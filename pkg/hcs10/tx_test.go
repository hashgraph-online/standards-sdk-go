package hcs10

import "testing"

func TestBuildCreateTopicTx(t *testing.T) {
	transaction, err := BuildCreateTopicTx(CreateTopicTxParams{
		TopicType: TopicTypeOutbound,
		TTL:       60,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

func TestBuildSubmitMessageTx(t *testing.T) {
	transaction, err := BuildSubmitMessageTx("0.0.1001", BuildMessagePayload("0.0.100@0.0.200", "hello", ""), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}
