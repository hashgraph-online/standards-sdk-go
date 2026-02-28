package hcs18

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

// CreateDiscoveryTopic creates the requested resource.
func (c *Client) CreateDiscoveryTopic(
	ctx context.Context,
	options CreateDiscoveryTopicOptions,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	transaction := BuildCreateDiscoveryTopicTx(CreateDiscoveryTopicTxParams{
		TTLSeconds:   options.TTLSeconds,
		AdminKey:     c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride: options.MemoOverride,
	})
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute discovery topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get discovery topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to create discovery topic")
	}
	return receipt.TopicID.String(), receipt, nil
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	topicID string,
	message DiscoveryMessage,
	transactionMemo string,
) (hedera.TransactionReceipt, error) {
	_ = ctx

	transaction, err := BuildSubmitDiscoveryMessageTx(topicID, message, transactionMemo)
	if err != nil {
		return hedera.TransactionReceipt{}, err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to execute discovery message transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return hedera.TransactionReceipt{}, fmt.Errorf("failed to get discovery message receipt: %w", err)
	}
	return receipt, nil
}

// Announce performs the requested operation.
func (c *Client) Announce(ctx context.Context, topicID string, data AnnounceData, memo string) (hedera.TransactionReceipt, error) {
	return c.SubmitMessage(ctx, topicID, BuildAnnounceMessage(data), memo)
}

// Propose performs the requested operation.
func (c *Client) Propose(ctx context.Context, topicID string, data ProposeData, memo string) (hedera.TransactionReceipt, error) {
	return c.SubmitMessage(ctx, topicID, BuildProposeMessage(data), memo)
}

// Respond performs the requested operation.
func (c *Client) Respond(ctx context.Context, topicID string, data RespondData, memo string) (hedera.TransactionReceipt, error) {
	return c.SubmitMessage(ctx, topicID, BuildRespondMessage(data), memo)
}

// Complete performs the requested operation.
func (c *Client) Complete(ctx context.Context, topicID string, data CompleteData, memo string) (hedera.TransactionReceipt, error) {
	return c.SubmitMessage(ctx, topicID, BuildCompleteMessage(data), memo)
}

// Withdraw performs the requested operation.
func (c *Client) Withdraw(ctx context.Context, topicID string, data WithdrawData, memo string) (hedera.TransactionReceipt, error) {
	return c.SubmitMessage(ctx, topicID, BuildWithdrawMessage(data), memo)
}

// GetDiscoveryMessages performs the requested operation.
func (c *Client) GetDiscoveryMessages(
	ctx context.Context,
	topicID string,
	sequenceNumber string,
	limit int,
	order string,
) ([]MessageRecord, error) {
	items, err := c.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
		SequenceNumber: strings.TrimSpace(sequenceNumber),
		Limit:          limit,
		Order:          strings.TrimSpace(order),
	})
	if err != nil {
		return nil, err
	}

	records := make([]MessageRecord, 0, len(items))
	for _, item := range items {
		message, err := decodeDiscoveryMessage(item.Message)
		if err != nil {
			continue
		}
		if err := ValidateMessage(message); err != nil {
			continue
		}
		records = append(records, MessageRecord{
			Message:            message,
			ConsensusTimestamp: item.ConsensusTimestamp,
			SequenceNumber:     item.SequenceNumber,
			Payer:              item.PayerAccountID,
		})
	}
	return records, nil
}

// IsProposalReady performs the requested operation.
func (c *Client) IsProposalReady(proposal TrackedProposal) bool {
	acceptances := 0
	for _, response := range proposal.Responses {
		if response.Decision == "accept" {
			acceptances++
		}
	}
	requiredResponses := len(proposal.Data.Members) - 1
	return acceptances >= requiredResponses
}

func decodeDiscoveryMessage(encoded string) (DiscoveryMessage, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return DiscoveryMessage{}, err
	}

	var envelope struct {
		P    string            `json:"p"`
		Op   DiscoveryOperation `json:"op"`
		Data json.RawMessage   `json:"data"`
	}
	if err := json.Unmarshal(decoded, &envelope); err != nil {
		return DiscoveryMessage{}, err
	}

	message := DiscoveryMessage{
		P:  strings.TrimSpace(envelope.P),
		Op: envelope.Op,
	}
	switch envelope.Op {
	case OperationAnnounce:
		var payload AnnounceData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return DiscoveryMessage{}, err
		}
		message.Data = payload
	case OperationPropose:
		var payload ProposeData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return DiscoveryMessage{}, err
		}
		message.Data = payload
	case OperationRespond:
		var payload RespondData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return DiscoveryMessage{}, err
		}
		message.Data = payload
	case OperationComplete:
		var payload CompleteData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return DiscoveryMessage{}, err
		}
		message.Data = payload
	case OperationWithdraw:
		var payload WithdrawData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return DiscoveryMessage{}, err
		}
		message.Data = payload
	default:
		return DiscoveryMessage{}, fmt.Errorf("unsupported operation %q", envelope.Op)
	}
	return message, nil
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

