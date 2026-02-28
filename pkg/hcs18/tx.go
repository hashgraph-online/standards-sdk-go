package hcs18

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type CreateDiscoveryTopicTxParams struct {
	TTLSeconds   int64
	AdminKey     hedera.Key
	SubmitKey    hedera.Key
	MemoOverride string
}

// BuildCreateDiscoveryTopicTx builds and returns the configured value.
func BuildCreateDiscoveryTopicTx(params CreateDiscoveryTopicTxParams) *hedera.TopicCreateTransaction {
	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(BuildDiscoveryMemo(params.TTLSeconds, params.MemoOverride))
	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}
	return transaction
}

// BuildSubmitDiscoveryMessageTx builds and returns the configured value.
func BuildSubmitDiscoveryMessageTx(topicID string, message DiscoveryMessage, transactionMemo string) (*hedera.TopicMessageSubmitTransaction, error) {
	if err := ValidateMessage(message); err != nil {
		return nil, err
	}
	parsedTopicID, err := hedera.TopicIDFromString(strings.TrimSpace(topicID))
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HCS-18 message: %w", err)
	}

	resolvedMemo := strings.TrimSpace(transactionMemo)
	if resolvedMemo == "" {
		resolvedMemo = BuildTransactionMemo(message.Op)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(payload).
		SetTransactionMemo(resolvedMemo)
	return transaction, nil
}

// BuildAnnounceMessage performs the requested operation.
func BuildAnnounceMessage(data AnnounceData) DiscoveryMessage {
	return DiscoveryMessage{
		P:    "hcs-18",
		Op:   OperationAnnounce,
		Data: data,
	}
}

// BuildProposeMessage performs the requested operation.
func BuildProposeMessage(data ProposeData) DiscoveryMessage {
	return DiscoveryMessage{
		P:    "hcs-18",
		Op:   OperationPropose,
		Data: data,
	}
}

// BuildRespondMessage performs the requested operation.
func BuildRespondMessage(data RespondData) DiscoveryMessage {
	return DiscoveryMessage{
		P:    "hcs-18",
		Op:   OperationRespond,
		Data: data,
	}
}

// BuildCompleteMessage performs the requested operation.
func BuildCompleteMessage(data CompleteData) DiscoveryMessage {
	return DiscoveryMessage{
		P:    "hcs-18",
		Op:   OperationComplete,
		Data: data,
	}
}

// BuildWithdrawMessage performs the requested operation.
func BuildWithdrawMessage(data WithdrawData) DiscoveryMessage {
	return DiscoveryMessage{
		P:    "hcs-18",
		Op:   OperationWithdraw,
		Data: data,
	}
}

