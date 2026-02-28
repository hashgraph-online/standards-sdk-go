package hcs6

import (
	"context"
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
	if !ValidateTTL(ttl) {
		return CreateRegistryResult{}, fmt.Errorf("TTL must be at least 3600 seconds")
	}

	transaction := BuildCreateRegistryTx(CreateRegistryTxParams{
		TTL:          ttl,
		AdminKey:     c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride: "",
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

// RegisterEntry registers the requested resource.
func (c *Client) RegisterEntry(
	ctx context.Context,
	registryTopicID string,
	options RegisterEntryOptions,
) (OperationResult, error) {
	_ = ctx

	analyticsMemo := strings.TrimSpace(options.AnalyticsMemo)
	if analyticsMemo == "" {
		analyticsMemo = BuildTransactionMemo()
	}

	transaction, err := BuildRegisterEntryTx(RegisterEntryTxParams{
		RegistryTopicID: registryTopicID,
		TargetTopicID:   options.TargetTopicID,
		Memo:            options.Memo,
		AnalyticsMemo:   analyticsMemo,
	})
	if err != nil {
		return OperationResult{}, err
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

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	registryTopicID string,
	message Message,
	transactionMemo string,
) (OperationResult, error) {
	_ = ctx
	if err := ValidateMessage(message); err != nil {
		return OperationResult{}, err
	}

	topicID, err := hedera.TopicIDFromString(strings.TrimSpace(registryTopicID))
	if err != nil {
		return OperationResult{}, fmt.Errorf("invalid registry topic ID: %w", err)
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return OperationResult{}, fmt.Errorf("failed to marshal HCS-6 message: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicID).
		SetMessage(payload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
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

// GetRegistry returns the requested value.
func (c *Client) GetRegistry(ctx context.Context, topicID string, options QueryRegistryOptions) (TopicRegistry, error) {
	topicInfo, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return TopicRegistry{}, err
	}
	parsedMemo, ok := ParseTopicMemo(topicInfo.Memo)
	if !ok {
		return TopicRegistry{}, fmt.Errorf("topic %s is not an HCS-6 registry", topicID)
	}

	order := strings.TrimSpace(options.Order)
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
			RegistryType:       RegistryTypeNonIndexed,
		}
		entries = append(entries, entry)
		if latestEntry == nil || entry.Timestamp > latestEntry.Timestamp {
			copyEntry := entry
			latestEntry = &copyEntry
		}
	}

	resolvedEntries := make([]RegistryEntry, 0, 1)
	if latestEntry != nil {
		resolvedEntries = append(resolvedEntries, *latestEntry)
	}

	return TopicRegistry{
		TopicID:      topicID,
		RegistryType: RegistryTypeNonIndexed,
		TTL:          parsedMemo.TTL,
		Entries:      resolvedEntries,
		LatestEntry:  latestEntry,
	}, nil
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
