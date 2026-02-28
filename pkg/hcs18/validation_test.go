package hcs18

import "testing"

func TestValidateAnnounceMessage(t *testing.T) {
	err := ValidateMessage(BuildAnnounceMessage(AnnounceData{
		Account: "0.0.1234",
		Petal: PetalDescriptor{
			Name:     "worker",
			Priority: 1,
		},
		Capabilities: CapabilityDetails{
			Protocols: []string{"hcs-10"},
		},
	}))
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

func TestValidateRespondMessage(t *testing.T) {
	err := ValidateMessage(BuildRespondMessage(RespondData{
		Responder:   "0.0.1234",
		ProposalSeq: 1,
		Decision:    "accept",
	}))
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

