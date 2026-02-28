package hcs18

import "testing"

func TestBuildSubmitDiscoveryMessageTx(t *testing.T) {
	transaction, err := BuildSubmitDiscoveryMessageTx("0.0.1001", BuildAnnounceMessage(AnnounceData{
		Account: "0.0.1234",
		Petal: PetalDescriptor{
			Name:     "worker",
			Priority: 1,
		},
		Capabilities: CapabilityDetails{
			Protocols: []string{"hcs-10"},
		},
	}), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

