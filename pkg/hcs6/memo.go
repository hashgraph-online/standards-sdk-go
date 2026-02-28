package hcs6

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultTTL int64 = 86400

var topicMemoPattern = regexp.MustCompile(`^hcs-6:(\d):(\d+)$`)

// BuildTopicMemo performs the requested operation.
func BuildTopicMemo(ttl int64) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-6:%d:%d", RegistryTypeNonIndexed, resolvedTTL)
}

// ParseTopicMemo performs the requested operation.
func ParseTopicMemo(memo string) (*TopicMemo, bool) {
	matches := topicMemoPattern.FindStringSubmatch(strings.TrimSpace(memo))
	if len(matches) != 3 {
		return nil, false
	}
	if matches[1] != "1" {
		return nil, false
	}

	var ttl int64
	_, err := fmt.Sscanf(matches[2], "%d", &ttl)
	if err != nil || ttl <= 0 {
		return nil, false
	}

	return &TopicMemo{
		Protocol:     "hcs-6",
		RegistryType: RegistryTypeNonIndexed,
		TTL:          ttl,
	}, true
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo() string {
	return "hcs-6:op:0:1"
}

// BuildHRL performs the requested operation.
func BuildHRL(topicID string) string {
	return "hcs://6/" + strings.TrimSpace(topicID)
}
