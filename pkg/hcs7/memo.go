package hcs7

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultTTL int64 = 86400

var topicMemoPattern = regexp.MustCompile(`^hcs-7:indexed:(\d+)$`)

// BuildTopicMemo performs the requested operation.
func BuildTopicMemo(ttl int64) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-7:indexed:%d", resolvedTTL)
}

// ParseTopicMemo performs the requested operation.
func ParseTopicMemo(memo string) (*TopicMemo, bool) {
	matches := topicMemoPattern.FindStringSubmatch(strings.TrimSpace(memo))
	if len(matches) != 2 {
		return nil, false
	}

	var ttl int64
	_, err := fmt.Sscanf(matches[1], "%d", &ttl)
	if err != nil || ttl <= 0 {
		return nil, false
	}

	return &TopicMemo{
		Protocol: "hcs-7",
		Indexed:  true,
		TTL:      ttl,
	}, true
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo(operationEnum int) string {
	return fmt.Sprintf("hcs-7:op:%d:0", operationEnum)
}

