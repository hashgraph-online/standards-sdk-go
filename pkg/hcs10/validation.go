package hcs10

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	topicIDPattern   = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	accountIDPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

// ValidateMessage performs the requested operation.
func ValidateMessage(message Message) error {
	if strings.TrimSpace(message.P) != "hcs-10" {
		return fmt.Errorf("message p must be hcs-10")
	}
	switch message.Op {
	case OperationConnectionRequest:
		if strings.TrimSpace(message.OperatorID) == "" {
			return fmt.Errorf("connection_request requires operator_id")
		}
	case OperationConnectionCreated:
		if !topicIDPattern.MatchString(strings.TrimSpace(message.ConnectionTopicID)) {
			return fmt.Errorf("connection_created requires valid connection_topic_id")
		}
		if strings.TrimSpace(message.OperatorID) == "" {
			return fmt.Errorf("connection_created requires operator_id")
		}
	case OperationMessage:
		if strings.TrimSpace(message.OperatorID) == "" {
			return fmt.Errorf("message requires operator_id")
		}
		if strings.TrimSpace(message.Data) == "" {
			return fmt.Errorf("message requires data")
		}
	case OperationRegister:
		if !accountIDPattern.MatchString(strings.TrimSpace(message.AccountID)) {
			return fmt.Errorf("register requires valid account_id")
		}
	case OperationDelete:
		if strings.TrimSpace(message.UID) == "" {
			return fmt.Errorf("delete requires uid")
		}
	case OperationCloseConnection, OperationTransaction:
	default:
		return fmt.Errorf("unsupported operation %q", message.Op)
	}
	if len(strings.TrimSpace(message.Memo)) > 500 {
		return fmt.Errorf("memo exceeds 500 characters")
	}
	return nil
}
