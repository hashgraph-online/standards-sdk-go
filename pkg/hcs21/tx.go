package hcs21

import (
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// BuildRegistryMemo performs the requested operation.
func BuildRegistryMemo(ttl int64, indexed bool, topicType TopicType, metaTopicID string) (string, error) {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTopicTTL
	}
	indexedValue := 1
	if indexed {
		indexedValue = 0
	}

	trimmedMeta := strings.TrimSpace(metaTopicID)
	if trimmedMeta != "" && !metaPointerPattern.MatchString(trimmedMeta) {
		return "", &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "meta value must be short immutable pointer",
		}
	}

	memo := fmt.Sprintf("hcs-21:%d:%d:%d", indexedValue, resolvedTTL, topicType)
	if trimmedMeta != "" {
		memo = fmt.Sprintf("%s:%s", memo, trimmedMeta)
	}
	return memo, nil
}

// BuildCreateRegistryTopicTx builds and returns the configured value.
func BuildCreateRegistryTopicTx(
	ttl int64,
	indexed bool,
	topicType TopicType,
	metaTopicID string,
	adminKey hedera.Key,
	submitKey hedera.Key,
) (*hedera.TopicCreateTransaction, error) {
	memo, err := BuildRegistryMemo(ttl, indexed, topicType, metaTopicID)
	if err != nil {
		return nil, err
	}

	transaction := hedera.NewTopicCreateTransaction().SetTopicMemo(memo)
	if adminKey != nil {
		transaction.SetAdminKey(adminKey)
	}
	if submitKey != nil {
		transaction.SetSubmitKey(submitKey)
	}
	return transaction, nil
}

// BuildDeclarationMessageTx builds and returns the configured value.
func BuildDeclarationMessageTx(
	topicID string,
	declaration AdapterDeclaration,
	transactionMemo string,
) (*hedera.TopicMessageSubmitTransaction, error) {
	if err := ValidateDeclaration(declaration); err != nil {
		return nil, err
	}
	parsedTopicID, err := hedera.TopicIDFromString(strings.TrimSpace(topicID))
	if err != nil {
		return nil, fmt.Errorf("invalid topic ID: %w", err)
	}

	payload, err := json.Marshal(declaration)
	if err != nil {
		return nil, fmt.Errorf("failed to encode declaration: %w", err)
	}

	transaction := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(parsedTopicID).
		SetMessage(payload)
	if strings.TrimSpace(transactionMemo) != "" {
		transaction.SetTransactionMemo(strings.TrimSpace(transactionMemo))
	}
	return transaction, nil
}
