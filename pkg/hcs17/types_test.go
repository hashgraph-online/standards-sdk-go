package hcs17

import "testing"

func TestGenerateAndParseTopicMemo(t *testing.T) {
	memo := GenerateTopicMemo(7200)
	if memo != "hcs-17:0:7200" {
		t.Fatalf("unexpected generated topic memo: %s", memo)
	}

	parsed, err := ParseTopicMemo(memo)
	if err != nil {
		t.Fatalf("ParseTopicMemo failed: %v", err)
	}
	if parsed.Type != HCS17TopicTypeState {
		t.Fatalf("unexpected topic type: %d", parsed.Type)
	}
	if parsed.TTLSeconds != 7200 {
		t.Fatalf("unexpected ttl: %d", parsed.TTLSeconds)
	}
}

func TestParseTopicMemoRejectsInvalidMemo(t *testing.T) {
	if _, err := ParseTopicMemo("invalid"); err == nil {
		t.Fatalf("expected invalid memo error")
	}
}

func TestValidateStateHashMessage(t *testing.T) {
	message := StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "abc",
		Topics:    []string{"0.0.1"},
		AccountID: "0.0.2",
	}
	if errors := ValidateStateHashMessage(message); len(errors) != 0 {
		t.Fatalf("expected valid message, got errors: %+v", errors)
	}

	invalid := StateHashMessage{}
	if errors := ValidateStateHashMessage(invalid); len(errors) == 0 {
		t.Fatalf("expected validation errors for invalid message")
	}
}
