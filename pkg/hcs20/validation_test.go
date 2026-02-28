package hcs20

import (
	"encoding/json"
	"testing"
)

func TestNormalizeTick(t *testing.T) {
	if NormalizeTick("  LoYaL  ") != "loyal" {
		t.Fatal("expected lowercase/trimmed tick")
	}
}

func TestNormalizeAccountID(t *testing.T) {
	normalized, err := NormalizeAccountID("0.0.12345-abcde")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if normalized != "0.0.12345" {
		t.Fatalf("expected 0.0.12345, got %s", normalized)
	}
}

func TestNormalizeAccountIDInvalid(t *testing.T) {
	_, err := NormalizeAccountID("not-a-hedera-account")
	if err == nil {
		t.Fatal("expected invalid account error")
	}
}

func TestValidateDeployMessage(t *testing.T) {
	message := Message{
		Protocol:  "hcs-20",
		Operation: "deploy",
		Name:      "Loyalty",
		Tick:      "LOYAL",
		Max:       "100000",
		Limit:     "1000",
	}
	if err := ValidateMessage(message); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateMintMessage(t *testing.T) {
	message := Message{
		Protocol:  "HCS-20",
		Operation: "MINT",
		Tick:      "loyal",
		Amount:    "100",
		To:        "0.0.1001",
	}
	if err := ValidateMessage(message); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateTransferMessage(t *testing.T) {
	message := Message{
		Protocol:  "hcs-20",
		Operation: "transfer",
		Tick:      "loyal",
		Amount:    "10",
		From:      "0.0.1001",
		To:        "0.0.1002",
	}
	if err := ValidateMessage(message); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateRegisterRequiresPrivate(t *testing.T) {
	message := Message{
		Protocol:  "hcs-20",
		Operation: "register",
		Name:      "Topic A",
		TopicID:   "0.0.4567",
	}
	if err := ValidateMessage(message); err == nil {
		t.Fatal("expected validation failure when private field missing")
	}
}

func TestValidateNumberTooLong(t *testing.T) {
	message := Message{
		Protocol:  "hcs-20",
		Operation: "mint",
		Tick:      "loyal",
		Amount:    "12345678901234567890",
		To:        "0.0.1001",
	}
	if err := ValidateMessage(message); err == nil {
		t.Fatal("expected validation failure for oversized number string")
	}
}

func TestBuildMessagePayload(t *testing.T) {
	isPrivate := true
	payload, normalized, err := BuildMessagePayload(Message{
		Protocol:  "hcs-20",
		Operation: "register",
		Name:      "Loyal",
		TopicID:   "0.0.1234",
		Private:   &isPrivate,
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if normalized.Operation != "register" {
		t.Fatalf("expected normalized operation register, got %s", normalized.Operation)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("expected valid JSON payload: %v", err)
	}
}

func TestParseMessageBytes(t *testing.T) {
	payload := []byte(`{"p":"hcs-20","op":"burn","tick":"LOYAL","amt":"5","from":"0.0.12345"}`)
	message, err := ParseMessageBytes(payload)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if message.Tick != "loyal" {
		t.Fatalf("expected normalized tick loyal, got %s", message.Tick)
	}
}
