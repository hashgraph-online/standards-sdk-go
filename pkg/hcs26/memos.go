package hcs26

import (
	"fmt"
	"strings"
)

// BuildTopicMemo performs the requested operation.
func BuildTopicMemo(indexed bool, ttlSeconds int64, topicType TopicType) string {
	resolvedTTL := ttlSeconds
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTLSeconds
	}
	indexedValue := 1
	if indexed {
		indexedValue = 0
	}
	return fmt.Sprintf("%s:%d:%d:%d", Protocol, indexedValue, resolvedTTL, topicType)
}

// ParseTopicMemo performs the requested operation.
func ParseTopicMemo(memoRaw string) (*TopicMemo, bool) {
	matches := discoveryMemoRe.FindStringSubmatch(strings.TrimSpace(memoRaw))
	if len(matches) != 4 {
		return nil, false
	}

	var indexedValue int
	var ttl int64
	var topicType int
	_, err := fmt.Sscanf(matches[1], "%d", &indexedValue)
	if err != nil {
		return nil, false
	}
	_, err = fmt.Sscanf(matches[2], "%d", &ttl)
	if err != nil || ttl <= 0 {
		return nil, false
	}
	_, err = fmt.Sscanf(matches[3], "%d", &topicType)
	if err != nil {
		return nil, false
	}
	if indexedValue != 0 && indexedValue != 1 {
		return nil, false
	}
	if topicType < int(TopicTypeDiscovery) || topicType > int(TopicTypeReputation) {
		return nil, false
	}

	return &TopicMemo{
		Protocol:   Protocol,
		Indexed:    indexedValue == 0,
		TTLSeconds: ttl,
		TopicType:  TopicType(topicType),
	}, true
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo(operation Operation, topicType TopicType) string {
	return fmt.Sprintf("%s:op:%d:%d", Protocol, operation, topicType)
}

// ParseTransactionMemo performs the requested operation.
func ParseTransactionMemo(memoRaw string) (*TransactionMemo, bool) {
	matches := txMemoRe.FindStringSubmatch(strings.TrimSpace(memoRaw))
	if len(matches) != 3 {
		return nil, false
	}

	var operation int
	var topicType int
	_, err := fmt.Sscanf(matches[1], "%d", &operation)
	if err != nil {
		return nil, false
	}
	_, err = fmt.Sscanf(matches[2], "%d", &topicType)
	if err != nil {
		return nil, false
	}
	if operation < int(OperationRegister) || operation > int(OperationMigrate) {
		return nil, false
	}
	if topicType < int(TopicTypeDiscovery) || topicType > int(TopicTypeReputation) {
		return nil, false
	}

	return &TransactionMemo{
		Protocol:  Protocol,
		Operation: Operation(operation),
		TopicType: TopicType(topicType),
	}, true
}
