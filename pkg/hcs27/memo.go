package hcs27

import (
	"fmt"
	"strconv"
	"strings"
)

func BuildTopicMemo(ttlSeconds int64) string {
	if ttlSeconds <= 0 {
		ttlSeconds = 86400
	}
	return fmt.Sprintf("hcs-27:0:%d:0", ttlSeconds)
}

func ParseTopicMemo(memo string) (*TopicMemo, bool) {
	parts := strings.Split(strings.TrimSpace(memo), ":")
	if len(parts) != 4 {
		return nil, false
	}
	if parts[0] != "hcs-27" {
		return nil, false
	}

	indexedFlag, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, false
	}

	ttlSeconds, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, false
	}

	topicType, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, false
	}

	return &TopicMemo{
		IndexedFlag: indexedFlag,
		TTLSeconds:  ttlSeconds,
		TopicType:   topicType,
	}, true
}

func BuildTransactionMemo() string {
	return "hcs-27:op:0:0"
}
