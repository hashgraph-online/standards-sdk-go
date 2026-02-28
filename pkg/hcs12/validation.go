package hcs12

import (
	"fmt"
	"regexp"
	"strings"
)

var topicIDPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidatePayload performs the requested operation.
func ValidatePayload(payload map[string]any) error {
	protocolValue, ok := payload["p"].(string)
	if !ok || strings.TrimSpace(protocolValue) != "hcs-12" {
		return fmt.Errorf("payload p must be hcs-12")
	}
	operationValue, ok := payload["op"].(string)
	if !ok || strings.TrimSpace(operationValue) == "" {
		return fmt.Errorf("payload op is required")
	}

	switch strings.TrimSpace(operationValue) {
	case "register":
		if topicValue, hasTopic := payload["t_id"]; hasTopic {
			if topicID, ok := topicValue.(string); !ok || !topicIDPattern.MatchString(strings.TrimSpace(topicID)) {
				return fmt.Errorf("payload t_id must be a Hedera topic ID")
			}
		}
	case "add-action":
		topicID, ok := payload["t_id"].(string)
		if !ok || !topicIDPattern.MatchString(strings.TrimSpace(topicID)) {
			return fmt.Errorf("add-action requires valid t_id")
		}
	case "add-block":
		blockTopicID, ok := payload["block_t_id"].(string)
		if !ok || !topicIDPattern.MatchString(strings.TrimSpace(blockTopicID)) {
			return fmt.Errorf("add-block requires valid block_t_id")
		}
	case "update":
	default:
		return fmt.Errorf("unsupported payload op %q", operationValue)
	}
	return nil
}
