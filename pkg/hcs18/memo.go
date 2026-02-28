package hcs18

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultTTL int64 = 86400

var topicMemoPattern = regexp.MustCompile(`^hcs-18:0(?::(\d+))?$`)

// BuildDiscoveryMemo performs the requested operation.
func BuildDiscoveryMemo(ttlSeconds int64, memoOverride string) string {
	trimmedOverride := strings.TrimSpace(memoOverride)
	if trimmedOverride != "" {
		return trimmedOverride
	}
	if ttlSeconds > 0 {
		return fmt.Sprintf("hcs-18:0:%d", ttlSeconds)
	}
	return "hcs-18:0"
}

// ParseDiscoveryMemo performs the requested operation.
func ParseDiscoveryMemo(memo string) (*TopicMemo, bool) {
	matches := topicMemoPattern.FindStringSubmatch(strings.TrimSpace(memo))
	if len(matches) == 0 {
		return nil, false
	}
	ttl := DefaultTTL
	if len(matches) > 1 && strings.TrimSpace(matches[1]) != "" {
		_, err := fmt.Sscanf(matches[1], "%d", &ttl)
		if err != nil || ttl <= 0 {
			return nil, false
		}
	}
	return &TopicMemo{
		Protocol: "hcs-18",
		Type:     0,
		TTL:      ttl,
	}, true
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo(operation DiscoveryOperation) string {
	return fmt.Sprintf("hcs-18:op:%d", opCode(operation))
}

func opCode(operation DiscoveryOperation) int {
	switch operation {
	case OperationAnnounce:
		return 0
	case OperationPropose:
		return 1
	case OperationRespond:
		return 2
	case OperationComplete:
		return 3
	case OperationWithdraw:
		return 4
	default:
		return 0
	}
}
