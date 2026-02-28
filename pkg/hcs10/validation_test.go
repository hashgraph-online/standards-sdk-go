package hcs10

import "testing"

func TestValidateMessagePayload(t *testing.T) {
	err := ValidateMessage(BuildMessagePayload("0.0.100@0.0.200", "hello", "memo"))
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

func TestValidateRegisterMessage(t *testing.T) {
	err := ValidateMessage(BuildRegistryRegisterMessage("0.0.1234", "0.0.400", ""))
	if err != nil {
		t.Fatalf("expected valid register message, got %v", err)
	}
}
