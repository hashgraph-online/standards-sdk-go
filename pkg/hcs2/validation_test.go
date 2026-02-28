package hcs2

import "testing"

func TestValidateMessage_Register(t *testing.T) {
	err := ValidateMessage(Message{
		P:       "hcs-2",
		Op:      OperationRegister,
		TopicID: "0.0.123",
		Memo:    "ok",
	})
	if err != nil {
		t.Fatalf("expected valid message, got error: %v", err)
	}
}

func TestValidateMessage_Invalid(t *testing.T) {
	err := ValidateMessage(Message{
		P:  "hcs-2",
		Op: OperationUpdate,
	})
	if err == nil {
		t.Fatalf("expected invalid message to fail validation")
	}
}
