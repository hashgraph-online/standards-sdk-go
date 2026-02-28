package hcs17

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

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

	operatorID := strings.TrimSpace(config.OperatorAccountID)
	if operatorID == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	operatorKey := strings.TrimSpace(config.OperatorPrivateKey)
	if operatorKey == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	parsedOperatorID, err := hedera.AccountIDFromString(operatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	parsedOperatorKey, err := shared.ParsePrivateKey(operatorKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(parsedOperatorID, parsedOperatorKey)

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
		operatorID:   parsedOperatorID,
		operatorKey:  parsedOperatorKey,
	}, nil
}

// HederaClient returns the configured Hedera SDK client.
func (c *Client) HederaClient() *hedera.Client {
	return c.hederaClient
}

// MirrorClient returns the configured mirror node client.
func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

// CreateStateTopic creates the requested resource.
func (c *Client) CreateStateTopic(
	ctx context.Context,
	options CreateTopicOptions,
) (string, error) {
	_ = ctx

	transaction := BuildCreateStateTopicTx(options)
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to execute state topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", fmt.Errorf("failed to get state topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", fmt.Errorf("failed to create HCS-17 topic")
	}
	return receipt.TopicID.String(), nil
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	topicID string,
	message StateHashMessage,
	transactionMemo string,
) (hedera.TransactionReceipt, error) {
	_ = ctx

	transaction, err := BuildStateHashMessageTx(topicID, message, transactionMemo)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to execute HCS-17 message transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to get HCS-17 message receipt: %w", err)
	}
	return receipt, nil
}

// ComputeAndPublish computes the requested state payload.
func (c *Client) ComputeAndPublish(
	ctx context.Context,
	options ComputeAndPublishOptions,
) (ComputeAndPublishResult, error) {
	topicStates := make([]TopicState, 0, len(options.Topics))
	for _, topicID := range options.Topics {
		messages, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
			Limit: 1,
			Order: "desc",
		})
		if err != nil {
			return ComputeAndPublishResult{}, err
		}
		runningHash := ""
		if len(messages) > 0 {
			runningHash = messages[0].RunningHash
		}
		topicStates = append(topicStates, TopicState{
			TopicID:           topicID,
			LatestRunningHash: runningHash,
		})
	}

	calculated, err := c.CalculateAccountStateHash(AccountStateInput{
		AccountID: options.AccountID,
		PublicKey: options.AccountPublicKey,
		Topics:    topicStates,
	})
	if err != nil {
		return ComputeAndPublishResult{}, err
	}

	message := c.CreateStateHashMessage(
		calculated.StateHash,
		options.AccountID,
		options.Topics,
		options.Memo,
		nil,
	)
	receipt, err := c.SubmitMessage(ctx, options.PublishTopicID, message, "")
	if err != nil {
		return ComputeAndPublishResult{}, err
	}

	return ComputeAndPublishResult{
		StateHash: calculated.StateHash,
		Receipt:   receipt,
	}, nil
}

// ValidateTopic validates the provided input value.
func (c *Client) ValidateTopic(ctx context.Context, topicID string) (bool, *TopicMemo, error) {
	info, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return false, nil, err
	}
	parsed, err := ParseTopicMemo(info.Memo)
	if err != nil {
		return false, nil, err
	}
	return true, parsed, nil
}

// GetRecentMessages returns the requested value.
func (c *Client) GetRecentMessages(
	ctx context.Context,
	topicID string,
	limit int,
	order string,
) ([]MessageRecord, error) {
	queryLimit := limit
	if queryLimit <= 0 {
		queryLimit = 25
	}
	queryOrder := strings.TrimSpace(order)
	if queryOrder == "" {
		queryOrder = "desc"
	}

	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		Limit: queryLimit,
		Order: queryOrder,
	})
	if err != nil {
		return nil, err
	}

	results := make([]MessageRecord, 0, len(items))
	for _, item := range items {
		decoded, decodeErr := base64.StdEncoding.DecodeString(item.Message)
		if decodeErr != nil {
			continue
		}
		var message StateHashMessage
		if unmarshalErr := json.Unmarshal(decoded, &message); unmarshalErr != nil {
			continue
		}
		if validationErrors := ValidateStateHashMessage(message); len(validationErrors) > 0 {
			continue
		}

		results = append(results, MessageRecord{
			Message:            message,
			ConsensusTimestamp: item.ConsensusTimestamp,
			SequenceNumber:     item.SequenceNumber,
			Payer:              item.PayerAccountID,
		})
	}

	return results, nil
}

// GetLatestMessage returns the requested value.
func (c *Client) GetLatestMessage(ctx context.Context, topicID string) (*MessageRecord, error) {
	items, err := c.GetRecentMessages(ctx, topicID, 1, "desc")
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

// CalculateAccountStateHash calculates the requested value.
func (c *Client) CalculateAccountStateHash(input AccountStateInput) (StateHashResult, error) {
	if strings.TrimSpace(input.AccountID) == "" {
		return StateHashResult{}, fmt.Errorf("account ID is required")
	}
	publicKeyString, err := normalizePublicKeyValue(input.PublicKey)
	if err != nil {
		return StateHashResult{}, err
	}

	sortedTopics := append([]TopicState{}, input.Topics...)
	sort.Slice(sortedTopics, func(index int, other int) bool {
		return sortedTopics[index].TopicID < sortedTopics[other].TopicID
	})

	var builder strings.Builder
	for _, topic := range sortedTopics {
		builder.WriteString(topic.TopicID)
		builder.WriteString(topic.LatestRunningHash)
	}
	builder.WriteString(publicKeyString)

	hash := sha512.Sum384([]byte(builder.String()))
	stateHash := hex.EncodeToString(hash[:])

	return StateHashResult{
		StateHash:  stateHash,
		AccountID:  input.AccountID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		TopicCount: len(input.Topics),
	}, nil
}

// CalculateCompositeStateHash calculates the requested value.
func (c *Client) CalculateCompositeStateHash(input CompositeStateInput) (CompositeStateHashResult, error) {
	if strings.TrimSpace(input.CompositeAccountID) == "" {
		return CompositeStateHashResult{}, fmt.Errorf("composite account ID is required")
	}

	sortedMembers := append([]CompositeMemberState{}, input.MemberStates...)
	sort.Slice(sortedMembers, func(index int, other int) bool {
		return sortedMembers[index].AccountID < sortedMembers[other].AccountID
	})

	sortedTopics := append([]TopicState{}, input.CompositeTopics...)
	sort.Slice(sortedTopics, func(index int, other int) bool {
		return sortedTopics[index].TopicID < sortedTopics[other].TopicID
	})

	var builder strings.Builder
	for _, member := range sortedMembers {
		builder.WriteString(member.AccountID)
		builder.WriteString(member.StateHash)
	}
	for _, topic := range sortedTopics {
		builder.WriteString(topic.TopicID)
		builder.WriteString(topic.LatestRunningHash)
	}
	builder.WriteString(input.CompositePublicKeyFingerprint)

	hash := sha512.Sum384([]byte(builder.String()))
	stateHash := hex.EncodeToString(hash[:])

	return CompositeStateHashResult{
		StateHash:           stateHash,
		AccountID:           input.CompositeAccountID,
		Timestamp:           time.Now().UTC().Format(time.RFC3339Nano),
		TopicCount:          len(input.CompositeTopics),
		MemberCount:         len(input.MemberStates),
		CompositeTopicCount: len(input.CompositeTopics),
	}, nil
}

// CalculateKeyFingerprint calculates the requested value.
func (c *Client) CalculateKeyFingerprint(keys []hedera.PublicKey, threshold int) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("keys are required")
	}
	if threshold <= 0 {
		return "", fmt.Errorf("threshold must be positive")
	}

	keyStrings := make([]string, 0, len(keys))
	for _, key := range keys {
		keyStrings = append(keyStrings, key.String())
	}
	sort.Strings(keyStrings)

	payload := map[string]any{
		"threshold": threshold,
		"keys":      keyStrings,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode key fingerprint payload: %w", err)
	}

	hash := sha512.Sum384(encodedPayload)
	return hex.EncodeToString(hash[:]), nil
}

// CreateStateHashMessage creates the requested resource.
func (c *Client) CreateStateHashMessage(
	stateHash string,
	accountID string,
	topicIDs []string,
	memo string,
	epoch *int64,
) StateHashMessage {
	return StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: stateHash,
		Topics:    topicIDs,
		AccountID: accountID,
		Epoch:     epoch,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Memo:      memo,
	}
}

// VerifyStateHash performs the requested operation.
func (c *Client) VerifyStateHash(input any, expectedHash string) (bool, error) {
	switch typed := input.(type) {
	case AccountStateInput:
		calculated, err := c.CalculateAccountStateHash(typed)
		if err != nil {
			return false, err
		}
		return calculated.StateHash == expectedHash, nil
	case CompositeStateInput:
		calculated, err := c.CalculateCompositeStateHash(typed)
		if err != nil {
			return false, err
		}
		return calculated.StateHash == expectedHash, nil
	default:
		return false, fmt.Errorf("unsupported state hash input type")
	}
}

func normalizePublicKeyValue(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return "", fmt.Errorf("public key string is required")
		}
		return strings.TrimSpace(typed), nil
	case hedera.PublicKey:
		if strings.TrimSpace(typed.String()) == "" {
			return "", fmt.Errorf("public key is required")
		}
		return typed.String(), nil
	default:
		return "", fmt.Errorf("public key must be a string or hedera.PublicKey")
	}
}
