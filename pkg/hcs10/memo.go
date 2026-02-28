package hcs10

import (
	"fmt"
	"regexp"
	"strings"
)

const DefaultTTL int64 = 60

var (
	inboundMemoPattern    = regexp.MustCompile(`^hcs-10:0:(\d+):0:(\d+\.\d+\.\d+)$`)
	outboundMemoPattern   = regexp.MustCompile(`^hcs-10:0:(\d+):1$`)
	connectionMemoPattern = regexp.MustCompile(`^hcs-10:1:(\d+):2:(\d+\.\d+\.\d+):(\d+)$`)
	registryMemoPattern   = regexp.MustCompile(`^hcs-10:0:(\d+):3(?::(.+))?$`)
)

// BuildInboundMemo performs the requested operation.
func BuildInboundMemo(ttl int64, accountID string) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-10:0:%d:0:%s", resolvedTTL, strings.TrimSpace(accountID))
}

// BuildOutboundMemo performs the requested operation.
func BuildOutboundMemo(ttl int64) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-10:0:%d:1", resolvedTTL)
}

// BuildConnectionMemo performs the requested operation.
func BuildConnectionMemo(ttl int64, inboundTopicID string, connectionID int64) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	return fmt.Sprintf("hcs-10:1:%d:2:%s:%d", resolvedTTL, strings.TrimSpace(inboundTopicID), connectionID)
}

// BuildRegistryMemo performs the requested operation.
func BuildRegistryMemo(ttl int64, metadataTopicID string) string {
	resolvedTTL := ttl
	if resolvedTTL <= 0 {
		resolvedTTL = DefaultTTL
	}
	trimmedMetadataTopicID := strings.TrimSpace(metadataTopicID)
	if trimmedMetadataTopicID == "" {
		return fmt.Sprintf("hcs-10:0:%d:3", resolvedTTL)
	}
	return fmt.Sprintf("hcs-10:0:%d:3:%s", resolvedTTL, trimmedMetadataTopicID)
}

// ParseMemoType performs the requested operation.
func ParseMemoType(memo string) (TopicType, bool) {
	trimmed := strings.TrimSpace(memo)
	switch {
	case inboundMemoPattern.MatchString(trimmed):
		return TopicTypeInbound, true
	case outboundMemoPattern.MatchString(trimmed):
		return TopicTypeOutbound, true
	case connectionMemoPattern.MatchString(trimmed):
		return TopicTypeConnection, true
	case registryMemoPattern.MatchString(trimmed):
		return TopicTypeRegistry, true
	default:
		return TopicTypeInbound, false
	}
}

// BuildTransactionMemo performs the requested operation.
func BuildTransactionMemo(operationEnum int, topicTypeEnum int) string {
	return fmt.Sprintf("hcs-10:op:%d:%d", operationEnum, topicTypeEnum)
}

