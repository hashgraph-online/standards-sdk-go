package hcs2

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient    *hedera.Client
	mirrorClient    *mirror.Client
	operatorID      hedera.AccountID
	operatorKey     hedera.PrivateKey
	registryTypeMap map[string]RegistryType
	mutex           sync.RWMutex
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
		hederaClient:    hederaClient,
		mirrorClient:    mirrorClient,
		operatorID:      operatorID,
		operatorKey:     operatorKey,
		registryTypeMap: map[string]RegistryType{},
	}, nil
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// CreateRegistry creates the requested resource.
func (c *Client) CreateRegistry(
	ctx context.Context,
	options CreateRegistryOptions,
) (CreateRegistryResult, error) {
	registryType := options.RegistryType
	if registryType != RegistryTypeIndexed && registryType != RegistryTypeNonIndexed {
		registryType = RegistryTypeIndexed
	}
	ttl := options.TTL
	if ttl <= 0 {
		ttl = 86400
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(BuildTopicMemo(registryType, ttl))

	adminKey, err := c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin)
	if err != nil {
		return CreateRegistryResult{}, err
	}
	if adminKey != nil {
		transaction.SetAdminKey(*adminKey)
	}

	submitKey, err := c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit)
	if err != nil {
		return CreateRegistryResult{}, err
	}
	if submitKey != nil {
		transaction.SetSubmitKey(*submitKey)
	}

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

	topicID := receipt.TopicID.String()
	c.mutex.Lock()
	c.registryTypeMap[topicID] = registryType
	c.mutex.Unlock()

	return CreateRegistryResult{
		Success:       true,
		TopicID:       topicID,
		TransactionID: response.TransactionID.String(),
	}, nil
}

// RegisterEntry registers the requested resource.
func (c *Client) RegisterEntry(
	ctx context.Context,
	registryTopicID string,
	options RegisterEntryOptions,
	protocol string,
) (OperationResult, error) {
	if protocol == "" {
		protocol = "hcs-2"
	}
	message := Message{
		P:        protocol,
		Op:       OperationRegister,
		TopicID:  options.TargetTopicID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationRegister, registryType)
	}

	return c.submitMessage(registryTopicID, message, analyticsMemo)
}

// UpdateEntry updates the requested resource.
func (c *Client) UpdateEntry(
	ctx context.Context,
	registryTopicID string,
	options UpdateEntryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}
	if registryType != RegistryTypeIndexed {
		return OperationResult{}, fmt.Errorf("update is only valid for indexed registries")
	}

	message := Message{
		P:        "hcs-2",
		Op:       OperationUpdate,
		TopicID:  options.TargetTopicID,
		UID:      options.UID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationUpdate, registryType)
	}

	return c.submitMessage(registryTopicID, message, analyticsMemo)
}

// DeleteEntry deletes the requested resource.
func (c *Client) DeleteEntry(
	ctx context.Context,
	registryTopicID string,
	options DeleteEntryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}
	if registryType != RegistryTypeIndexed {
		return OperationResult{}, fmt.Errorf("delete is only valid for indexed registries")
	}

	message := Message{
		P:    "hcs-2",
		Op:   OperationDelete,
		UID:  options.UID,
		Memo: options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationDelete, registryType)
	}

	return c.submitMessage(registryTopicID, message, analyticsMemo)
}

// MigrateRegistry performs the requested operation.
func (c *Client) MigrateRegistry(
	ctx context.Context,
	registryTopicID string,
	options MigrateRegistryOptions,
) (OperationResult, error) {
	registryType, err := c.resolveRegistryType(ctx, registryTopicID, options.RegistryType)
	if err != nil {
		return OperationResult{}, err
	}

	message := Message{
		P:        "hcs-2",
		Op:       OperationMigrate,
		TopicID:  options.TargetTopicID,
		Metadata: options.Metadata,
		Memo:     options.Memo,
	}
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	analyticsMemo := options.AnalyticsMemo
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo(OperationMigrate, registryType)
	}

	return c.submitMessage(registryTopicID, message, analyticsMemo)
}

// GetRegistry returns the requested value.
func (c *Client) GetRegistry(
	ctx context.Context,
	topicID string,
	options QueryRegistryOptions,
) (TopicRegistry, error) {
	topicInfo, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return TopicRegistry{}, err
	}

	memoInfo, ok := ParseTopicMemo(topicInfo.Memo)
	if !ok {
		return TopicRegistry{}, fmt.Errorf("topic %s is not an HCS-2 registry", topicID)
	}

	order := options.Order
	if order == "" {
		order = "asc"
	}

	sequenceNumber := ""
	if options.Skip > 0 {
		sequenceNumber = fmt.Sprintf("gt:%d", options.Skip)
	}

	messages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: sequenceNumber,
		Limit:          options.Limit,
		Order:          order,
	})
	if err != nil {
		return TopicRegistry{}, err
	}

	entries := make([]RegistryEntry, 0, len(messages))
	var latestEntry *RegistryEntry

	for _, item := range messages {
		var message Message
		if err := mirror.DecodeMessageJSON(item, &message); err != nil {
			continue
		}
		if err := ValidateMessage(message); err != nil {
			continue
		}

		entry := RegistryEntry{
			TopicID:            topicID,
			Sequence:           item.SequenceNumber,
			Timestamp:          item.ConsensusTimestamp,
			Payer:              item.PayerAccountID,
			Message:            message,
			ConsensusTimestamp: item.ConsensusTimestamp,
			RegistryType:       memoInfo.RegistryType,
		}
		entries = append(entries, entry)

		if latestEntry == nil || entry.Timestamp > latestEntry.Timestamp {
			copyEntry := entry
			latestEntry = &copyEntry
		}
	}

	if memoInfo.RegistryType == RegistryTypeNonIndexed {
		if latestEntry == nil {
			entries = []RegistryEntry{}
		} else {
			entries = []RegistryEntry{*latestEntry}
		}
	}

	c.mutex.Lock()
	c.registryTypeMap[topicID] = memoInfo.RegistryType
	c.mutex.Unlock()

	return TopicRegistry{
		TopicID:      topicID,
		RegistryType: memoInfo.RegistryType,
		TTL:          memoInfo.TTL,
		Entries:      entries,
		LatestEntry:  latestEntry,
	}, nil
}

// GetTopicInfo returns the requested value.
func (c *Client) GetTopicInfo(ctx context.Context, topicID string) (mirror.TopicInfo, error) {
	return c.mirrorClient.GetTopicInfo(ctx, topicID)
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	registryTopicID string,
	payload Message,
	transactionMemo string,
) (OperationResult, error) {
	_ = ctx
	return c.submitMessage(registryTopicID, payload, transactionMemo)
}

func (c *Client) submitMessage(
	registryTopicID string,
	message Message,
	transactionMemo string,
) (OperationResult, error) {
	topicID, err := hedera.TopicIDFromString(strings.TrimSpace(registryTopicID))
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid registry topic ID: %w", err)
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to marshal HCS-2 message: %w", err)
	}

	// If payload exceeds 1024 bytes, inscribe via HCS-1 and submit a reference.
	if len(payload) > 1024 {
		hrl, digest, inscribeErr := c.inscribeHCS1(payload)
		if inscribeErr != nil {
			return OperationResult{}, fmt.Errorf("failed to inscribe overflow payload via HCS-1: %w", inscribeErr)
		}

		wrapper := OverflowMessage{
			P:              message.P,
			Op:             message.Op,
			DataRef:        hrl,
			DataRefDigest:  digest,
		}
		payload, err = json.Marshal(wrapper)
		if err != nil {
			return OperationResult{}, fmt.Errorf("failed to marshal overflow wrapper: %w", err)
		}
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicID).
		SetMessage(payload)

	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(transactionMemo)
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to execute message submit transaction: %w", err)
	}

	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to get message submit receipt: %w", err)
	}

	return OperationResult{
		Success:        true,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}, nil
}

// inscribeHCS1 creates an HCS-1 topic, publishes the payload, and returns an HRL + SHA-256 digest.
func (c *Client) inscribeHCS1(payload []byte) (string, string, error) {
	createResp, err := hedera.NewTopicCreateTransaction().
		SetTopicMemo("hcs-1:0:0").
		SetAdminKey(c.operatorKey.PublicKey()).
		SetSubmitKey(c.operatorKey.PublicKey()).
		Execute(c.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to create HCS-1 overflow topic: %w", err)
	}
	createReceipt, err := createResp.GetReceipt(c.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to get HCS-1 overflow topic receipt: %w", err)
	}
	if createReceipt.TopicID == nil {
		return "", "", fmt.Errorf("HCS-1 overflow topic receipt missing topic ID")
	}
	dataTopic := *createReceipt.TopicID

	_, err = hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(dataTopic).
		SetMessage(payload).
		Execute(c.hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to publish HCS-1 overflow payload: %w", err)
	}

	hrl := fmt.Sprintf("hcs://1/%s", dataTopic.String())
	sum := sha256.Sum256(payload)
	digest := base64.RawURLEncoding.EncodeToString(sum[:])

	return hrl, digest, nil
}

func (c *Client) resolveRegistryType(
	ctx context.Context,
	topicID string,
	override *RegistryType,
) (RegistryType, error) {
	if override != nil {
		return *override, nil
	}

	c.mutex.RLock()
	cachedType, ok := c.registryTypeMap[topicID]
	c.mutex.RUnlock()
	if ok {
		return cachedType, nil
	}

	topicInfo, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return RegistryTypeIndexed, err
	}
	memoInfo, parsed := ParseTopicMemo(topicInfo.Memo)
	if !parsed {
		return RegistryTypeIndexed, fmt.Errorf("topic %s is not an HCS-2 registry", topicID)
	}

	c.mutex.Lock()
	c.registryTypeMap[topicID] = memoInfo.RegistryType
	c.mutex.Unlock()

	return memoInfo.RegistryType, nil
}

func (c *Client) resolvePublicKey(rawKey string, useOperator bool) (*hedera.PublicKey, error) {
	if useOperator {
		publicKey := c.operatorKey.PublicKey()
		return &publicKey, nil
	}

	if strings.TrimSpace(rawKey) == "" {
		return nil, nil
	}

	publicKey, pubErr := hedera.PublicKeyFromString(rawKey)
	if pubErr == nil {
		return &publicKey, nil
	}

	privateKey, prvErr := shared.ParsePrivateKey(rawKey)
	if prvErr != nil {
		return nil, fmt.Errorf("failed to parse key as public (%v) or private (%v)", pubErr, prvErr)
	}

	derivedPublicKey := privateKey.PublicKey()
	return &derivedPublicKey, nil
}
