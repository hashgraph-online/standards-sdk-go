package hcs10

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

// CreateInboundTopic creates the requested resource.
func (c *Client) CreateInboundTopic(ctx context.Context, options CreateTopicOptions) (string, hedera.TransactionReceipt, error) {
	return c.createTopic(ctx, CreateTopicTxParams{
		TopicType:    TopicTypeInbound,
		TTL:          options.TTL,
		AccountID:    options.AccountID,
		AdminKey:     c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride: options.MemoOverride,
	}, options.TransactionMemo)
}

// CreateOutboundTopic creates the requested resource.
func (c *Client) CreateOutboundTopic(ctx context.Context, options CreateTopicOptions) (string, hedera.TransactionReceipt, error) {
	return c.createTopic(ctx, CreateTopicTxParams{
		TopicType:    TopicTypeOutbound,
		TTL:          options.TTL,
		AdminKey:     c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:    c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride: options.MemoOverride,
	}, options.TransactionMemo)
}

// CreateConnectionTopic creates the requested resource.
func (c *Client) CreateConnectionTopic(ctx context.Context, options CreateTopicOptions) (string, hedera.TransactionReceipt, error) {
	return c.createTopic(ctx, CreateTopicTxParams{
		TopicType:     TopicTypeConnection,
		TTL:           options.TTL,
		InboundTopicID: options.InboundTopicID,
		ConnectionID:  options.ConnectionID,
		AdminKey:      c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:     c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride:  options.MemoOverride,
	}, options.TransactionMemo)
}

// CreateRegistryTopic creates the requested resource.
func (c *Client) CreateRegistryTopic(ctx context.Context, options CreateTopicOptions) (CreateRegistryTopicResult, error) {
	topicID, receipt, err := c.createTopic(ctx, CreateTopicTxParams{
		TopicType:       TopicTypeRegistry,
		TTL:             options.TTL,
		MetadataTopicID: options.MetadataTopicID,
		AdminKey:        c.resolvePublicKey(options.AdminKey, options.UseOperatorAsAdmin),
		SubmitKey:       c.resolvePublicKey(options.SubmitKey, options.UseOperatorAsSubmit),
		MemoOverride:    options.MemoOverride,
	}, options.TransactionMemo)
	if err != nil {
		return CreateRegistryTopicResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	_ = receipt
	return CreateRegistryTopicResult{
		Success: true,
		TopicID: topicID,
	}, nil
}

func (c *Client) createTopic(
	ctx context.Context,
	params CreateTopicTxParams,
	transactionMemo string,
) (string, hedera.TransactionReceipt, error) {
	_ = ctx

	transaction, err := BuildCreateTopicTx(params)
	if err != nil {
		return "", hedera.TransactionReceipt{}, err
	}
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to execute topic create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("failed to get topic create receipt: %w", err)
	}
	if receipt.TopicID == nil {
		return "", hedera.TransactionReceipt{}, fmt.Errorf("topic create receipt missing topic ID")
	}
	return receipt.TopicID.String(), receipt, nil
}

// SubmitMessage submits the requested message payload.
func (c *Client) SubmitMessage(
	ctx context.Context,
	topicID string,
	message Message,
	transactionMemo string,
) (SubmitResult, error) {
	_ = ctx
	transaction, err := BuildSubmitMessageTx(topicID, message, transactionMemo)
	if err != nil {
		return SubmitResult{}, err
	}
	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("failed to execute message submit transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("failed to get message submit receipt: %w", err)
	}
	return SubmitResult{
		Success:        true,
		TransactionID:  response.TransactionID.String(),
		SequenceNumber: int64(receipt.TopicSequenceNumber),
	}, nil
}

// SendConnectionRequest performs the requested operation.
func (c *Client) SendConnectionRequest(
	ctx context.Context,
	inboundTopicID string,
	operatorID string,
	memo string,
) (SubmitResult, error) {
	return c.SubmitMessage(ctx, inboundTopicID, BuildConnectionRequestMessage(operatorID, memo), BuildTransactionMemo(3, 1))
}

// ConfirmConnection performs the requested operation.
func (c *Client) ConfirmConnection(
	ctx context.Context,
	inboundTopicID string,
	connectionTopicID string,
	connectedAccountID string,
	operatorID string,
	connectionID int64,
	memo string,
) (SubmitResult, error) {
	message := BuildConnectionCreatedMessage(connectionTopicID, connectedAccountID, operatorID, connectionID, memo)
	return c.SubmitMessage(ctx, inboundTopicID, message, BuildTransactionMemo(4, 1))
}

// SendMessage performs the requested operation.
func (c *Client) SendMessage(
	ctx context.Context,
	connectionTopicID string,
	operatorID string,
	data string,
	memo string,
) (SubmitResult, error) {
	message := BuildMessagePayload(operatorID, data, memo)
	return c.SubmitMessage(ctx, connectionTopicID, message, BuildTransactionMemo(6, 3))
}

// RegisterAgent performs the requested operation.
func (c *Client) RegisterAgent(
	ctx context.Context,
	registryTopicID string,
	accountID string,
	inboundTopicID string,
	memo string,
) (SubmitResult, error) {
	return c.SubmitMessage(ctx, registryTopicID, BuildRegistryRegisterMessage(accountID, inboundTopicID, memo), BuildTransactionMemo(0, 0))
}

// DeleteAgent performs the requested operation.
func (c *Client) DeleteAgent(
	ctx context.Context,
	registryTopicID string,
	uid string,
	memo string,
) (SubmitResult, error) {
	return c.SubmitMessage(ctx, registryTopicID, BuildRegistryDeleteMessage(uid, memo), BuildTransactionMemo(1, 0))
}

// GetMessageStream performs the requested operation.
func (c *Client) GetMessageStream(
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
	validOps := map[Operation]bool{
		OperationMessage:         true,
		OperationCloseConnection: true,
		OperationTransaction:     true,
	}
	records := make([]MessageRecord, 0, len(items))
	for _, item := range items {
		message, err := decodeMessage(item.Message)
		if err != nil {
			continue
		}
		if strings.TrimSpace(message.P) != "hcs-10" {
			continue
		}
		if !validOps[message.Op] {
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

// GetTopicInfo performs the requested operation.
func (c *Client) GetTopicInfo(ctx context.Context, topicID string) (TopicRecord, error) {
	info, err := c.mirrorClient.GetTopicInfo(ctx, topicID)
	if err != nil {
		return TopicRecord{}, err
	}
	return TopicRecord{
		TopicID: info.TopicID,
		Memo:    info.Memo,
	}, nil
}

func decodeMessage(encoded string) (Message, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return Message{}, err
	}
	var message Message
	if err := json.Unmarshal(decoded, &message); err != nil {
		return Message{}, err
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

