package hcs6

import (
	"fmt"
	"regexp"
	"strings"
)

var topicIDPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidateTopicID performs the requested operation.
func ValidateTopicID(topicID string) bool {
	return topicIDPattern.MatchString(strings.TrimSpace(topicID))
}

// ValidateTTL performs the requested operation.
func ValidateTTL(ttl int64) bool {
	return ttl >= 3600
}

// ValidateMessage performs the requested operation.
func ValidateMessage(message Message) error {
	if strings.TrimSpace(message.P) != "hcs-6" {
		return fmt.Errorf("message p must be hcs-6")
	}
	if message.Op != OperationRegister {
		return fmt.Errorf("message op must be register")
	}
	if !ValidateTopicID(message.TopicID) {
		return fmt.Errorf("message t_id must be a Hedera topic ID")
	}
	if len(strings.TrimSpace(message.Memo)) > 500 {
		return fmt.Errorf("message memo exceeds 500 characters")
	}
	return nil
}

