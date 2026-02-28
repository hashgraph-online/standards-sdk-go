package hcs11

import (
	"context"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestClientGettersAndFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network: "testnet",
		Auth: Auth{
			OperatorID: "0.0.1",
			PrivateKey: pk.String(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if client.HederaClient() == nil || client.MirrorClient() == nil {
		t.Fatal("expected clients")
	}
	if client.OperatorID() != "0.0.1" {
		t.Fatal("expected id")
	}

	// Close the client so operations fail
	client.HederaClient().Close()
	ctx := context.Background()

	_ = client.CreatePersonalProfile("me", nil)

	_ = client.SetProfileForAccountMemo("0.0.1", 1)

	_, err = client.UpdateAccountMemoWithProfile(ctx, "0.0.1", "0.0.2")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.InscribeImage(ctx, []byte{1}, "img", InscribeImageOptions{})
	if err == nil { t.Fatal("expected fail") }

	resp, err := client.InscribeProfile(ctx, HCS11Profile{}, InscribeProfileOptions{})
	if err != nil || resp.Success { t.Fatal("expected no err but false success for invalid profile") }

	_, err = client.CreateAndInscribeProfile(ctx, HCS11Profile{}, false, InscribeProfileOptions{})
	if err != nil { t.Fatal("expected no generic error from wrapper") }
}

func TestGetAgentTypeFromMetadata(t *testing.T) {
	client := &Client{}
	
	res := client.GetAgentTypeFromMetadata(AgentMetadata{Type: "autonomous"})
	if res != AIAgentTypeAutonomous {
		t.Fatal("expected autonomous")
	}

	res2 := client.GetAgentTypeFromMetadata(AgentMetadata{})
	if res2 != AIAgentTypeManual {
		t.Fatal("expected manual")
	}
}

func TestAttachUAIDIfMissing(t *testing.T) {
	client := &Client{}
	ctx := context.Background()
	
	p1 := &HCS11Profile{UAID: "uaid:myagent"}
	err := client.AttachUAIDIfMissing(ctx, p1)
	if err != nil || p1.UAID != "uaid:myagent" {
		t.Fatal("should match uaid")
	}

	p2 := &HCS11Profile{}
	err = client.AttachUAIDIfMissing(ctx, p2)
	// without operator ID it just returns early
	if err != nil || p2.UAID != "" {
		t.Fatal("expected empty")
	}
}

func TestAssignProfileOptions(t *testing.T) {
	profile := &HCS11Profile{}
	opts := map[string]any{
		"alias":          "my-alias",
		"bio":            "my bio",
		"profileImage":   "img.png",
		"inboundTopicId": "0.0.1",
		"outboundTopicId": "0.0.2",
		"baseAccount":    "0.0.3",
	}
	assignProfileOptions(profile, opts)
	if profile.Alias != "my-alias" { t.Fatal("expected alias") }
	if profile.Bio != "my bio" { t.Fatal("expected bio") }
	if profile.ProfileImage != "img.png" { t.Fatal("expected img") }
	if profile.InboundTopicID != "0.0.1" { t.Fatal("expected inbound") }
	if profile.OutboundTopicID != "0.0.2" { t.Fatal("expected outbound") }
	if profile.BaseAccount != "0.0.3" { t.Fatal("expected base") }

	assignProfileOptions(nil, opts)
	assignProfileOptions(profile, nil)
}

func TestValidateProfileBranches(t *testing.T) {
	client := &Client{}

	// Flora missing members
	result := client.ValidateProfile(HCS11Profile{
		Version:     "1.0",
		DisplayName: "Flora",
		Type:        ProfileTypeFlora,
		Threshold:   0,
	})
	if result.Valid { t.Fatal("expected invalid flora") }

	// MCP Server invalid transport
	result2 := client.ValidateProfile(HCS11Profile{
		Version:     "1.0",
		DisplayName: "MCP",
		Type:        ProfileTypeMCPServer,
		MCPServer: &MCPServerDetails{
			Version:     "1.0",
			Description: "desc",
			ConnectionInfo: MCPServerConnectionInfo{
				URL:       "http://example.com",
				Transport: "http",
			},
			Services: []MCPServerCapability{MCPServerCapabilityToolProvider},
		},
	})
	if result2.Valid { t.Fatal("expected invalid transport") }

	// MCP Server invalid verification type
	result3 := client.ValidateProfile(HCS11Profile{
		Version:     "1.0",
		DisplayName: "MCP",
		Type:        ProfileTypeMCPServer,
		MCPServer: &MCPServerDetails{
			Version:     "1.0",
			Description: "desc",
			ConnectionInfo: MCPServerConnectionInfo{
				URL:       "http://example.com",
				Transport: "sse",
			},
			Services:     []MCPServerCapability{MCPServerCapabilityToolProvider},
			Verification: &MCPServerVerification{Type: "invalid"},
		},
	})
	if result3.Valid { t.Fatal("expected invalid verification") }

	// Invalid profile type
	result4 := client.ValidateProfile(HCS11Profile{
		Version:     "1.0",
		DisplayName: "Unknown",
		Type:        ProfileType(99),
	})
	if result4.Valid { t.Fatal("expected invalid type") }
}

func TestCreateAIAgentProfile(t *testing.T) {
	client := &Client{}

	_, err := client.CreateAIAgentProfile("Agent", AIAgentTypeAutonomous,
		[]AIAgentCapability{AIAgentCapabilityTextGeneration}, "gpt-4", nil)
	if err != nil { t.Fatalf("unexpected: %v", err) }

	_, err = client.CreateAIAgentProfile("Agent", AIAgentTypeAutonomous,
		[]AIAgentCapability{}, "gpt-4", nil)
	if err == nil { t.Fatal("expected validation error") }
}

func TestCreateMCPServerProfile(t *testing.T) {
	client := &Client{}

	_, err := client.CreateMCPServerProfile("Server", MCPServerDetails{
		Version:     "1.0",
		Description: "test",
		ConnectionInfo: MCPServerConnectionInfo{URL: "http://example.com", Transport: "sse"},
		Services:    []MCPServerCapability{MCPServerCapabilityToolProvider},
	}, nil)
	if err != nil { t.Fatalf("unexpected: %v", err) }

	_, err = client.CreateMCPServerProfile("Server", MCPServerDetails{}, nil)
	if err == nil { t.Fatal("expected validation error") }
}

func TestProfileToJSONAndParse(t *testing.T) {
	client := &Client{}

	profile := HCS11Profile{
		Version:     "1.0",
		Type:        ProfileTypePersonal,
		DisplayName: "Test",
	}

	jsonStr, err := client.ProfileToJSONString(profile)
	if err != nil { t.Fatalf("unexpected: %v", err) }

	parsed, err := client.ParseProfileFromString(jsonStr)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if parsed.DisplayName != "Test" { t.Fatal("expected Test") }

	_, err = client.ParseProfileFromString("invalid json")
	if err == nil { t.Fatal("expected error") }

	_, err = client.ParseProfileFromString(`{"version":"","display_name":"","type":99}`)
	if err == nil { t.Fatal("expected validation error") }
}

func TestGetCapabilitiesFromTags(t *testing.T) {
	client := &Client{}

	caps := client.GetCapabilitiesFromTags(nil)
	if len(caps) != 1 { t.Fatal("expected default") }

	caps2 := client.GetCapabilitiesFromTags([]string{"text_generation", "code_generation"})
	if len(caps2) != 2 { t.Fatal("expected 2") }

	caps3 := client.GetCapabilitiesFromTags([]string{"unknown_capability"})
	if len(caps3) != 1 { t.Fatal("expected fallback default") }
}
