package hcs2

import (
	"fmt"
	"regexp"
	"strings"
)

var topicIDPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidateMessage validates the provided input value.
func ValidateMessage(message Message) error {
	if !regexp.MustCompile(`^hcs-\d+$`).MatchString(message.P) {
		return fmt.Errorf("protocol must be in format hcs-N")
	}

	if message.Op != OperationRegister &&
		message.Op != OperationUpdate &&
		message.Op != OperationDelete &&
		message.Op != OperationMigrate {
		return fmt.Errorf("operation %q is not supported", message.Op)
	}

	if len(message.Memo) > 500 {
		return fmt.Errorf("memo must not exceed 500 characters")
	}

	switch message.Op {
	case OperationRegister:
		if !topicIDPattern.MatchString(strings.TrimSpace(message.TopicID)) {
			return fmt.Errorf("register requires valid t_id")
		}
	case OperationUpdate:
		if strings.TrimSpace(message.UID) == "" {
			return fmt.Errorf("update requires uid")
		}
		if !topicIDPattern.MatchString(strings.TrimSpace(message.TopicID)) {
			return fmt.Errorf("update requires valid t_id")
		}
	case OperationDelete:
		if strings.TrimSpace(message.UID) == "" {
			return fmt.Errorf("delete requires uid")
		}
	case OperationMigrate:
		if !topicIDPattern.MatchString(strings.TrimSpace(message.TopicID)) {
			return fmt.Errorf("migrate requires valid t_id")
		}
	}

	if message.TTL < 0 {
		return fmt.Errorf("ttl cannot be negative")
	}

	return nil
}
