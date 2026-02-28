package hcs10

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// BuildCreateTopicTx builds and returns the configured value.
func BuildCreateTopicTx(params CreateTopicTxParams) (*hedera.TopicCreateTransaction, error) {
	ttl := params.TTL
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	memo := strings.TrimSpace(params.MemoOverride)
	if memo == "" {
		switch params.TopicType {
		case TopicTypeInbound:
			memo = BuildInboundMemo(ttl, params.AccountID)
		case TopicTypeOutbound:
			memo = BuildOutboundMemo(ttl)
		case TopicTypeConnection:
			memo = BuildConnectionMemo(ttl, params.InboundTopicID, params.ConnectionID)
		case TopicTypeRegistry:
			memo = BuildRegistryMemo(ttl, params.MetadataTopicID)
		default:
			return nil, fmt.Errorf("unsupported topic type %d", params.TopicType)
		}
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(memo)
	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}
	return transaction, nil
}

// BuildSubmitMessageTx builds and returns the configured value.
func BuildSubmitMessageTx(topicID string, message Message, transactionMemo string) (*hedera.TopicMessageSubmitTransaction, error) {
	if err := ValidateMessage(message); err != nil {
		return nil, err
	}

	parsedTopicID, err := hedera.TopicIDFromString(strings.TrimSpace(topicID))
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HCS-10 payload: %w", err)
	}
	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(payload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	return transaction, nil
}

// BuildConnectionRequestMessage performs the requested operation.
func BuildConnectionRequestMessage(operatorID string, memo string) Message {
	return Message{
		P:          "hcs-10",
		Op:         OperationConnectionRequest,
		OperatorID: strings.TrimSpace(operatorID),
		Memo:       strings.TrimSpace(memo),
	}
}

// BuildConnectionCreatedMessage performs the requested operation.
func BuildConnectionCreatedMessage(
	connectionTopicID string,
	connectedAccountID string,
	operatorID string,
	connectionID int64,
	memo string,
) Message {
	return Message{
		P:                  "hcs-10",
		Op:                 OperationConnectionCreated,
		ConnectionTopicID:  strings.TrimSpace(connectionTopicID),
		ConnectedAccountID: strings.TrimSpace(connectedAccountID),
		OperatorID:         strings.TrimSpace(operatorID),
		ConnectionID:       connectionID,
		Memo:               strings.TrimSpace(memo),
	}
}

// BuildMessagePayload performs the requested operation.
func BuildMessagePayload(operatorID string, data string, memo string) Message {
	return Message{
		P:          "hcs-10",
		Op:         OperationMessage,
		OperatorID: strings.TrimSpace(operatorID),
		Data:       data,
		Memo:       strings.TrimSpace(memo),
	}
}

// BuildRegistryRegisterMessage performs the requested operation.
func BuildRegistryRegisterMessage(accountID string, inboundTopicID string, memo string) Message {
	return Message{
		P:              "hcs-10",
		Op:             OperationRegister,
		AccountID:      strings.TrimSpace(accountID),
		InboundTopicID: strings.TrimSpace(inboundTopicID),
		Memo:           strings.TrimSpace(memo),
	}
}

// BuildRegistryDeleteMessage performs the requested operation.
func BuildRegistryDeleteMessage(uid string, memo string) Message {
	return Message{
		P:    "hcs-10",
		Op:   OperationDelete,
		UID:  strings.TrimSpace(uid),
		Memo: strings.TrimSpace(memo),
	}
}
