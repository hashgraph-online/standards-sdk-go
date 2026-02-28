package hcs6

import "testing"

func TestBuildCreateRegistryTxPrimary(t *testing.T) {
	transaction := BuildCreateRegistryTx(CreateRegistryTxParams{
		TTL: 3600,
	})
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

func TestBuildRegisterEntryTxPrimary(t *testing.T) {
	transaction, err := BuildRegisterEntryTx(RegisterEntryTxParams{
		RegistryTopicID: "0.0.123",
		TargetTopicID:   "0.0.456",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}
