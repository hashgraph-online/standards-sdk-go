package hcs17

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func BuildCreateStateTopicTx(options CreateTopicOptions) *hedera.TopicCreateTransaction {
	ttlSeconds := options.TTLSeconds
	if ttlSeconds <= 0 {
		ttlSeconds = 86400
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(GenerateTopicMemo(ttlSeconds))
	if options.AdminKey != nil {
		transaction.SetAdminKey(options.AdminKey)
	}
	if options.SubmitKey != nil {
		transaction.SetSubmitKey(options.SubmitKey)
	}
	if strings.TrimSpace(options.TransactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(options.TransactionMemo))
	}
	return transaction
}

func BuildStateHashMessageTx(topicID string, message StateHashMessage, transactionMemo string) (*hedera.TopicMessageSubmitTransaction, error) {
	if strings.TrimSpace(topicID) == "" {
		return nil, fmt.Errorf("topic ID is required")
	}
	if validationErrors := ValidateStateHashMessage(message); len(validationErrors) > 0 {
		return nil, fmt.Errorf("invalid HCS-17 message: %s", strings.Join(validationErrors, ", "))
	}

	parsedTopicID, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	payload := StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: message.StateHash,
		Topics:    message.Topics,
		AccountID: message.AccountID,
		Epoch:     message.Epoch,
		Timestamp: message.Timestamp,
		Memo:      message.Memo,
	}
	if strings.TrimSpace(payload.Timestamp) == "" {
		payload.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode HCS-17 message payload: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(encodedPayload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}

	return transaction, nil
}
