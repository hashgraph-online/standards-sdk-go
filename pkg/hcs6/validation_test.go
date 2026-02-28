package hcs6

import "testing"

func TestValidateMessage(t *testing.T) {
	err := ValidateMessage(Message{
		P:       "hcs-6",
		Op:      OperationRegister,
		TopicID: "0.0.123",
	})
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

func TestValidateMessageInvalidTopicID(t *testing.T) {
	err := ValidateMessage(Message{
		P:       "hcs-6",
		Op:      OperationRegister,
		TopicID: "invalid",
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

