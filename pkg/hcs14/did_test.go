package hcs14

import "testing"

func TestCreateUAIDAIDAndParse(t *testing.T) {
	uaid, err := CreateUAIDAID(CanonicalAgentData{
		Registry: "ANS",
		Name:     "Support Agent",
		Version:  "1.0.1",
		Protocol: "a2a",
		NativeID: "ote.agent.cs3p.com",
		Skills:   []int{3, 2, 9},
	}, RoutingParams{
		UID:      "ans://v1.0.1.ote.agent.cs3p.com",
		Registry: "ans",
		Proto:    "a2a",
		NativeID: "ote.agent.cs3p.com",
		Version:  "1.0.1",
	}, true)
	if err != nil {
		t.Fatalf("CreateUAIDAID failed: %v", err)
	}

	parsed, err := ParseUAID(uaid)
	if err != nil {
		t.Fatalf("ParseUAID failed: %v", err)
	}
	if parsed.Target != "aid" {
		t.Fatalf("unexpected target: %s", parsed.Target)
	}
	if parsed.ID == "" {
		t.Fatalf("identifier must not be empty")
	}
	if parsed.Params["registry"] != "ans" {
		t.Fatalf("expected registry=ans, got %s", parsed.Params["registry"])
	}
	if parsed.Params["proto"] != "a2a" {
		t.Fatalf("expected proto=a2a, got %s", parsed.Params["proto"])
	}
	if parsed.Params["nativeId"] != "ote.agent.cs3p.com" {
		t.Fatalf("expected nativeId=ote.agent.cs3p.com, got %s", parsed.Params["nativeId"])
	}
}

func TestCreateUAIDFromDIDAddsSrcWhenSanitized(t *testing.T) {
	uaid, err := CreateUAIDFromDID(
		"did:web:example.com;service=agent",
		RoutingParams{
			UID:      "0",
			Registry: "self",
			Proto:    "mcp",
			NativeID: "example.com",
		},
	)
	if err != nil {
		t.Fatalf("CreateUAIDFromDID failed: %v", err)
	}

	parsed, err := ParseUAID(uaid)
	if err != nil {
		t.Fatalf("ParseUAID failed: %v", err)
	}
	if parsed.Target != "did" {
		t.Fatalf("expected did target, got %s", parsed.Target)
	}
	if parsed.ID != "example.com" {
		t.Fatalf("expected sanitized id example.com, got %s", parsed.ID)
	}
	if parsed.Params["src"] == "" {
		t.Fatalf("expected src parameter to be included for sanitized DID")
	}
}

func TestCreateUAIDAIDHCS10RequiresCAIP10(t *testing.T) {
	_, err := CreateUAIDAID(CanonicalAgentData{
		Registry: "ans",
		Name:     "HCS-10 Agent",
		Version:  "1.0.0",
		Protocol: "hcs-10",
		NativeID: "0.0.1234",
		Skills:   []int{1},
	}, RoutingParams{}, true)
	if err == nil {
		t.Fatalf("expected CAIP-10 validation error")
	}
}
