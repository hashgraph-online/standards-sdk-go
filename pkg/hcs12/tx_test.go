package hcs12

import "testing"

func TestBuildCreateRegistryTopicTx(t *testing.T) {
	transaction, err := BuildCreateRegistryTopicTx(CreateRegistryTopicTxParams{
		RegistryType: RegistryTypeAction,
		TTL:          3600,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

func TestBuildSubmitMessageTx(t *testing.T) {
	transaction, err := BuildSubmitMessageTx("0.0.1001", map[string]any{
		"p":    "hcs-12",
		"op":   "register",
		"name": "demo",
	}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}
