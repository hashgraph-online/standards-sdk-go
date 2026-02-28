package hcs12

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// BuildCreateRegistryTopicTx builds and returns the configured value.
func BuildCreateRegistryTopicTx(params CreateRegistryTopicTxParams) (*hedera.TopicCreateTransaction, error) {
	resolvedMemo := strings.TrimSpace(params.MemoOverride)
	if resolvedMemo == "" {
		memo, err := BuildRegistryMemo(params.RegistryType, params.TTL)
		if err != nil {
			return nil, err
		}
		resolvedMemo = memo
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(resolvedMemo)
	if params.AdminKey != nil {
		transaction.SetAdminKey(params.AdminKey)
	}
	if params.SubmitKey != nil {
		transaction.SetSubmitKey(params.SubmitKey)
	}
	return transaction, nil
}

// BuildSubmitMessageTx builds and returns the configured value.
func BuildSubmitMessageTx(topicID string, payload any, transactionMemo string) (*hedera.TopicMessageSubmitTransaction, error) {
	parsedTopicID, err := hedera.TopicIDFromString(strings.TrimSpace(topicID))
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	var payloadMap map[string]any
	switch typed := payload.(type) {
	case map[string]any:
		payloadMap = typed
	default:
		encodedPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		if err := json.Unmarshal(encodedPayload, &payloadMap); err != nil {
			return nil, fmt.Errorf("failed to normalize payload: %w", err)
		}
	}

	if err := ValidatePayload(payloadMap); err != nil {
		return nil, err
	}

	encodedPayload, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(encodedPayload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	return transaction, nil
}
