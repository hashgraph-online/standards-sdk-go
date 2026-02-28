package hcs7

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient *hedera.Client
	mirrorClient *mirror.Client
	operatorID   hedera.AccountID
	operatorKey  hedera.PrivateKey
}

// NewClient creates a new Client.
func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.OperatorAccountID) == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	if strings.TrimSpace(config.OperatorPrivateKey) == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	operatorID, err := hedera.AccountIDFromString(strings.TrimSpace(config.OperatorAccountID))
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	operatorKey, err := shared.ParsePrivateKey(config.OperatorPrivateKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(operatorID, operatorKey)

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
		APIKey:  config.MirrorAPIKey,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		hederaClient: hederaClient,
		mirrorClient: mirrorClient,
		operatorID:   operatorID,
		operatorKey:  operatorKey,
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// CreateRegistry creates the requested resource.
func (c *Client) CreateRegistry(ctx context.Context, options CreateRegistryOptions) (CreateRegistryResult, error) {
	_ = ctx

	ttl := options.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl < 3600 {
		return CreateRegistryResult{}, fmt.Errorf("TTL must be at least 3600 seconds")
	}

	transaction := BuildCreateRegistryTx(CreateRegistryTxParams{
		TTL:       ttl,
		AdminKey:  c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey: c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
	})

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return CreateRegistryResult{}, fmt.Errorf("failed to execute create topic transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return CreateRegistryResult{}, fmt.Errorf("failed to get create topic receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return CreateRegistryResult{}, fmt.Errorf("topic ID missing in create topic receipt")
	}

	return CreateRegistryResult{
		Success:       true,
		TopicID:       receipt.TopicID.String(),
		TransactionID: response.TransactionID.String(),
	}, nil
}

// RegisterConfig performs the requested operation.
func (c *Client) RegisterConfig(ctx context.Context, options RegisterConfigOptions) (RegistryOperationResult, error) {
	_ = ctx

	message, err := c.buildConfigMessage(options)
	if err != nil {
		return RegistryOperationResult{}, err
	}

	analyticsMemo := strings.TrimSpace(options.AnalyticsMemo)
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(0)
	}

	submitKey := c.resolvePrivateKey(options.SubmitKey)
	return c.submitMessage(options.RegistryTopicID, message, submitKey, analyticsMemo)
}

// RegisterMetadata performs the requested operation.
func (c *Client) RegisterMetadata(ctx context.Context, options RegisterMetadataOptions) (RegistryOperationResult, error) {
	_ = ctx

	if len(options.Tags) == 0 {
		return RegistryOperationResult{}, fmt.Errorf("tags are required")
	}
	data := map[string]any{
		"weight": options.Weight,
		"tags":   options.Tags,
	}
	for key, value := range options.Data {
		data[key] = value
	}

	message := Message{
		P:       "hcs-7",
		Op:      OperationRegister,
		TopicID: strings.TrimSpace(options.MetadataTopicID),
		Data:    data,
		Memo:    strings.TrimSpace(options.Memo),
	}

	analyticsMemo := strings.TrimSpace(options.AnalyticsMemo)
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(1)
	}

	submitKey := c.resolvePrivateKey(options.SubmitKey)
	return c.submitMessage(options.RegistryTopicID, message, submitKey, analyticsMemo)
}

// GetRegistry performs the requested operation.
func (c *Client) GetRegistry(ctx context.Context, topicID string, options QueryRegistryOptions) (RegistryTopic, error) {
	info, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return RegistryTopic{}, err
	}
	memoInfo, ok := ParseTopicMemo(info.Memo)
	if !ok {
		return RegistryTopic{}, fmt.Errorf("topic %s is not an HCS-7 registry", topicID)
	}

	order := strings.TrimSpace(options.Order)
	if order == "" {
		order = "asc"
	}
	sequenceNumber := ""
	if options.Skip > 0 {
		sequenceNumber = fmt.Sprintf("gt:%d", options.Skip)
	}

	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: sequenceNumber,
		Limit:          options.Limit,
		Order:          order,
	})
	if err != nil {
		return RegistryTopic{}, err
	}

	entries := make([]RegistryEntry, 0, len(items))
	for _, item := range items {
		message, err := decodeMessage(item.Message)
		if err != nil {
			continue
		}
		if err := ValidateMessage(message); err != nil {
			continue
		}
		entries = append(entries, RegistryEntry{
			SequenceNumber: item.SequenceNumber,
			Timestamp:      item.ConsensusTimestamp,
			Payer:          item.PayerAccountID,
			Message:        message,
		})
	}

	return RegistryTopic{
		TopicID: topicID,
		TTL:     memoInfo.TTL,
		Entries: entries,
	}, nil
}

func (c *Client) buildConfigMessage(options RegisterConfigOptions) (Message, error) {
	message := Message{
		P:    "hcs-7",
		Op:   OperationRegisterConfig,
		Type: options.Type,
		Memo: strings.TrimSpace(options.Memo),
	}
	switch options.Type {
	case ConfigTypeEVM:
		if options.EVM == nil {
			return Message{}, fmt.Errorf("EVM config is required")
		}
		message.Config = *options.EVM
	case ConfigTypeWASM:
		if options.WASM == nil {
			return Message{}, fmt.Errorf("WASM config is required")
		}
		message.Config = *options.WASM
	default:
		return Message{}, fmt.Errorf("unsupported config type %q", options.Type)
	}
	if err := ValidateMessage(message); err != nil {
		return Message{}, err
	}
	return message, nil
}

func (c *Client) submitMessage(topicID string, message Message, submitKey *hedera.PrivateKey, transactionMemo string) (RegistryOperationResult, error) {
	transaction, err := BuildSubmitMessageTx(topicID, message, transactionMemo)
	if err != nil {
		return RegistryOperationResult{}, err
	}

	if submitKey != nil {
		frozen, freezeErr := transaction.FreezeWith(c.hederaClient)
		if freezeErr != nil {
			return RegistryOperationResult{}, fmt.Errorf("failed to freeze transaction: %w", freezeErr)
		}
		frozen.Sign(*submitKey)
		response, executeErr := frozen.Execute(c.hederaClient)
		if executeErr != nil {
			return RegistryOperationResult{}, fmt.Errorf("failed to execute message submit transaction: %w", executeErr)
		}
		receipt, receiptErr := response.GetReceipt(c.hederaClient)
		if receiptErr != nil {
			return RegistryOperationResult{}, fmt.Errorf("failed to get message submit receipt: %w", receiptErr)
		}
		return RegistryOperationResult{
			Success:        true,
			TransactionID:  response.TransactionID.String(),
			SequenceNumber: int64(receipt.TopicSequenceNumber),
		}, nil
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return RegistryOperationResult{}, fmt.Errorf("failed to execute message submit transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return RegistryOperationResult{}, fmt.Errorf("failed to get message submit receipt: %w", err)
	}
	return RegistryOperationResult{
		Success:        true,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}, nil
}

func decodeMessage(encoded string) (Message, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return Message{}, err
	}
	var payload map[string]any
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return Message{}, err
	}
	message := Message{
		P:    readString(payload, "p"),
		Op:   Operation(readString(payload, "op")),
		Type: ConfigType(readString(payload, "t")),
		Memo: readString(payload, "m"),
	}
	if topicID := readString(payload, "t_id"); topicID != "" {
		message.TopicID = topicID
	}
	if configValue, ok := payload["c"]; ok {
		if message.Type == ConfigTypeEVM {
			var configPayload EvmConfigPayload
			bytes, _ := json.Marshal(configValue)
			if err := json.Unmarshal(bytes, &configPayload); err == nil {
				message.Config = configPayload
			}
		} else if message.Type == ConfigTypeWASM {
			var configPayload WasmConfigPayload
			bytes, _ := json.Marshal(configValue)
			if err := json.Unmarshal(bytes, &configPayload); err == nil {
				message.Config = configPayload
			}
		}
	}
	if dataValue, ok := payload["d"]; ok {
		if dataMap, ok := dataValue.(map[string]any); ok {
			message.Data = dataMap
		}
	}
	return message, nil
}

func readString(payload map[string]any, field string) string {
	value, ok := payload[field]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(typed)
}

func (c *Client) resolvePublicKey(raw string, useOperator bool) hedera.Key {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" && !useOperator {
		return nil
	}
	if useOperator {
		return c.operatorKey.PublicKey()
	}
	publicKey, err := hedera.PublicKeyFromString(trimmed)
	if err != nil {
		return nil
	}
	return publicKey
}

func (c *Client) resolvePrivateKey(raw string) *hedera.PrivateKey {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	privateKey, err := shared.ParsePrivateKey(trimmed)
	if err != nil {
		return nil
	}
	return &privateKey
}

