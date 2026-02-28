package hcs6

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type CreateRegistryTxParams struct {
	TTL          int64
	AdminKey     hedera.Key
	SubmitKey    hedera.Key
	MemoOverride string
}

type RegisterEntryTxParams struct {
	RegistryTopicID string
	TargetTopicID   string
	Memo            string
	AnalyticsMemo   string
}

// BuildCreateRegistryTx builds and returns the configured value.
func BuildCreateRegistryTx(params CreateRegistryTxParams) *hedera.TopicCreateTransaction {
	resolvedMemo := strings.TrimSpace(params.MemoOverride)
	if resolvedMemo == "" {
		resolvedMemo = BuildTopicMemo(params.TTL)
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(resolvedMemo)
	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}

	return transaction
}

// BuildRegisterEntryTx builds and returns the configured value.
func BuildRegisterEntryTx(params RegisterEntryTxParams) (*hedera.TopicMessageSubmitTransaction, error) {
	trimmedTopicID := strings.TrimSpace(params.RegistryTopicID)
	if trimmedTopicID == "" {
		return nil, fmt.Errorf("registry topic ID is required")
	}
	parsedTopicID, err := hedera.TopicIDFromString(trimmedTopicID)
	if err != nil {
		return nil, fmt.Errorf("invalid registry topic ID: %w", err)
	}

	message := Message{
		P:       "hcs-6",
		Op:      OperationRegister,
		TopicID: strings.TrimSpace(params.TargetTopicID),
		Memo:    strings.TrimSpace(params.Memo),
	}
	if err := ValidateMessage(message); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HCS-6 message: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(payload)
	if strings.TrimSpace(params.AnalyticsMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(params.AnalyticsMemo))
	}

	return transaction, nil
}
