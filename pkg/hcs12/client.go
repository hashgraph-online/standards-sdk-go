package hcs12

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

// CreateRegistryTopic creates the requested resource.
func (c *Client) CreateRegistryTopic(ctx context.Context, options CreateRegistryTopicOptions) (CreateTopicResult, error) {
	_ = ctx

	transaction, err := BuildCreateRegistryTopicTx(CreateRegistryTopicTxParams{
		RegistryType: options.RegistryType,
		TTL:          options.TTL,
		AdminKey:     c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride: options.MemoOverride,
	})
	if err != nil {
		return CreateTopicResult{}, err
	}
	if strings.TrimSpace(options.TransactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(options.TransactionMemo))
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return CreateTopicResult{}, fmt.Errorf("failed to execute topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return CreateTopicResult{}, fmt.Errorf("failed to get topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return CreateTopicResult{}, fmt.Errorf("topic create receipt missing topic ID")
	}
	return CreateTopicResult{
		Success:       true,
		TopicID:       receipt.TopicID.String(),
		TransactionID: response.TransactionID.String(),
	}, nil
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	topicID string,
	payload any,
	transactionMemo string,
) (SubmitMessageResult, error) {
	_ = ctx

	transaction, err := BuildSubmitMessageTx(topicID, payload, transactionMemo)
	if err != nil {
		return SubmitMessageResult{}, err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return SubmitMessageResult{}, fmt.Errorf("failed to execute message submit transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return SubmitMessageResult{}, fmt.Errorf("failed to get message submit receipt: %w", err)
	}
	return SubmitMessageResult{
		Success:        true,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}, nil
}

// RegisterAction performs the requested operation.
func (c *Client) RegisterAction(
	ctx context.Context,
	topicID string,
	registration ActionRegistration,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":           "hcs-12",
		"op":          "register",
		"name":        registration.Name,
		"version":     registration.Version,
		"description": registration.Description,
		"author":      registration.Author,
		"tags":        registration.Tags,
		"t_id":        registration.TID,
	}
	return c.SubmitMessage(ctx, topicID, payload, transactionMemo)
}

// RegisterAssembly performs the requested operation.
func (c *Client) RegisterAssembly(
	ctx context.Context,
	topicID string,
	registration AssemblyRegistration,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":           "hcs-12",
		"op":          "register",
		"name":        registration.Name,
		"version":     registration.Version,
		"description": registration.Description,
		"author":      registration.Author,
		"tags":        registration.Tags,
	}
	return c.SubmitMessage(ctx, topicID, payload, transactionMemo)
}

// RegisterHashLink performs the requested operation.
func (c *Client) RegisterHashLink(
	ctx context.Context,
	topicID string,
	registration HashLinksRegistration,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":           "hcs-12",
		"op":          "register",
		"t_id":        registration.TID,
		"name":        registration.Name,
		"description": registration.Description,
		"tags":        registration.Tags,
		"category":    registration.Category,
		"featured":    registration.Featured,
		"icon":        registration.Icon,
		"author":      registration.Author,
		"website":     registration.Website,
	}
	return c.SubmitMessage(ctx, topicID, payload, transactionMemo)
}

// AddActionToAssembly performs the requested operation.
func (c *Client) AddActionToAssembly(
	ctx context.Context,
	assemblyTopicID string,
	operation AssemblyAddAction,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":     "hcs-12",
		"op":    "add-action",
		"t_id":  operation.TID,
		"alias": operation.Alias,
	}
	return c.SubmitMessage(ctx, assemblyTopicID, payload, transactionMemo)
}

// AddBlockToAssembly performs the requested operation.
func (c *Client) AddBlockToAssembly(
	ctx context.Context,
	assemblyTopicID string,
	operation AssemblyAddBlock,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":          "hcs-12",
		"op":         "add-block",
		"block_t_id": operation.BlockID,
		"data":       operation.Data,
	}
	return c.SubmitMessage(ctx, assemblyTopicID, payload, transactionMemo)
}

// UpdateAssembly performs the requested operation.
func (c *Client) UpdateAssembly(
	ctx context.Context,
	assemblyTopicID string,
	operation AssemblyUpdate,
	transactionMemo string,
) (SubmitMessageResult, error) {
	payload := map[string]any{
		"p":    "hcs-12",
		"op":   "update",
		"data": operation.Data,
	}
	return c.SubmitMessage(ctx, assemblyTopicID, payload, transactionMemo)
}

// GetEntries performs the requested operation.
func (c *Client) GetEntries(ctx context.Context, topicID string, options QueryOptions) ([]RegistryEntry, error) {
	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: strings.TrimSpace(options.SequenceNumber),
		Limit:          options.Limit,
		Order:          strings.TrimSpace(options.Order),
	})
	if err != nil {
		return nil, err
	}

	entries := make([]RegistryEntry, 0, len(items))
	for _, item := range items {
		payload, err := decodePayload(item.Message)
		if err != nil {
			continue
		}
		if err := ValidatePayload(payload); err != nil {
			continue
		}
		entries = append(entries, RegistryEntry{
			SequenceNumber:     item.SequenceNumber,
			ConsensusTimestamp: item.ConsensusTimestamp,
			Payer:              item.PayerAccountID,
			Payload:            payload,
		})
	}
	return entries, nil
}

func decodePayload(encoded string) (map[string]any, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}
	return payload, nil
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
