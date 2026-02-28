package hcs11

import (
	"testing"
)

func TestAgentBuilderFullSuccess(t *testing.T) {
	profile, err := NewAgentBuilder().
		SetName("TestAgent").
		SetAlias("agent-alias").
		SetBio("A test agent").
		SetDescription("description override").
		SetCapabilities([]AIAgentCapability{AIAgentCapabilityTextGeneration}).
		SetType(AIAgentTypeAutonomous).
		SetModel("gpt-4").
		SetCreator("creator-1").
		AddSocial("twitter", "@test").
		AddProperty("key1", "value1").
		SetInboundTopicID("0.0.100").
		SetOutboundTopicID("0.0.200").
		SetBaseAccount("0.0.12345").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.DisplayName != "description override" {
		if profile.Bio != "description override" {
			t.Fatalf("SetDescription should call SetBio")
		}
	}
	if profile.AIAgent.Type != AIAgentTypeAutonomous {
		t.Fatalf("unexpected agent type: %v", profile.AIAgent.Type)
	}
	if profile.AIAgent.Model != "gpt-4" {
		t.Fatalf("unexpected model: %s", profile.AIAgent.Model)
	}
	if profile.AIAgent.Creator != "creator-1" {
		t.Fatalf("unexpected creator: %s", profile.AIAgent.Creator)
	}
	if len(profile.Socials) != 1 {
		t.Fatalf("expected 1 social, got %d", len(profile.Socials))
	}
	if profile.Properties["key1"] != "value1" {
		t.Fatal("expected property key1=value1")
	}
	if profile.InboundTopicID != "0.0.100" {
		t.Fatalf("unexpected inbound topic: %s", profile.InboundTopicID)
	}
	if profile.OutboundTopicID != "0.0.200" {
		t.Fatalf("unexpected outbound topic: %s", profile.OutboundTopicID)
	}
	if profile.BaseAccount != "0.0.12345" {
		t.Fatalf("unexpected base account: %s", profile.BaseAccount)
	}
}

func TestAgentBuilderMissingName(t *testing.T) {
	_, err := NewAgentBuilder().Build()
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestAgentBuilderAddSocialEmpty(t *testing.T) {
	builder := NewAgentBuilder().SetName("test")
	builder.AddSocial("", "handle")
	builder.AddSocial("platform", "")
	profile, _ := builder.Build()
	if len(profile.Socials) != 0 {
		t.Fatalf("expected no socials for empty inputs")
	}
}

func TestAgentBuilderAddSocialDuplicate(t *testing.T) {
	builder := NewAgentBuilder().SetName("test").
		AddSocial("twitter", "@old").
		AddSocial("twitter", "@new")
	profile, _ := builder.Build()
	if len(profile.Socials) != 1 {
		t.Fatalf("expected 1 social after dedup, got %d", len(profile.Socials))
	}
	if profile.Socials[0].Handle != "@new" {
		t.Fatalf("expected updated handle '@new', got %q", profile.Socials[0].Handle)
	}
}

func TestAgentBuilderAddPropertyEmptyKey(t *testing.T) {
	builder := NewAgentBuilder().SetName("test")
	builder.AddProperty("", "value")
	profile, _ := builder.Build()
	if profile.Properties != nil && len(profile.Properties) > 0 {
		t.Fatal("expected no properties for empty key")
	}
}

func TestPersonBuilderFullSuccess(t *testing.T) {
	profile, err := NewPersonBuilder().
		SetName("John Doe").
		SetAlias("johnd").
		SetBio("A person").
		SetDescription("description override").
		SetProfileImage("https://example.com/img.png").
		SetBaseAccount("0.0.12345").
		SetInboundTopicID("0.0.100").
		SetOutboundTopicID("0.0.200").
		AddSocial("twitter", "@john").
		AddProperty("color", "blue").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Type != ProfileTypePersonal {
		t.Fatalf("unexpected type: %v", profile.Type)
	}
	if profile.ProfileImage != "https://example.com/img.png" {
		t.Fatalf("unexpected profile image: %s", profile.ProfileImage)
	}
	if profile.Alias != "johnd" {
		t.Fatalf("unexpected alias: %s", profile.Alias)
	}
}

func TestPersonBuilderMissingName(t *testing.T) {
	_, err := NewPersonBuilder().Build()
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestPersonBuilderAddSocialEmpty(t *testing.T) {
	builder := NewPersonBuilder().SetName("test")
	builder.AddSocial("", "handle")
	builder.AddSocial("platform", "")
	profile, _ := builder.Build()
	if len(profile.Socials) != 0 {
		t.Fatalf("expected no socials for empty inputs")
	}
}

func TestPersonBuilderAddSocialDuplicate(t *testing.T) {
	builder := NewPersonBuilder().SetName("test").
		AddSocial("github", "old").
		AddSocial("github", "new")
	profile, _ := builder.Build()
	if len(profile.Socials) != 1 {
		t.Fatalf("expected 1 social after dedup, got %d", len(profile.Socials))
	}
}

func TestPersonBuilderAddPropertyEmptyKey(t *testing.T) {
	builder := NewPersonBuilder().SetName("test")
	builder.AddProperty("", "val")
	profile, _ := builder.Build()
	if profile.Properties != nil && len(profile.Properties) > 0 {
		t.Fatal("expected no properties for empty key")
	}
}

func TestFloraBuilderFullSuccess(t *testing.T) {
	profile, err := NewFloraBuilder().
		SetDisplayName("MyFlora").
		SetBio("A flora group").
		SetMembers([]FloraMember{{AccountID: "0.0.1", Weight: 1}}).
		SetThreshold(1).
		SetTopics(FloraTopics{
			Communication: "0.0.100",
			Transaction:   "0.0.200",
			State:         "0.0.300",
		}).
		SetPolicies(map[string]any{"voting": "majority"}).
		SetMetadata(map[string]any{"created": "2024"}).
		AddMetadata("version", "1.0").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Type != ProfileTypeFlora {
		t.Fatalf("unexpected type: %v", profile.Type)
	}
	if len(profile.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(profile.Members))
	}
	if profile.Threshold != 1 {
		t.Fatalf("unexpected threshold: %d", profile.Threshold)
	}
	if profile.Topics.State != "0.0.300" {
		t.Fatalf("unexpected state topic: %s", profile.Topics.State)
	}
	if profile.Policies["voting"] != "majority" {
		t.Fatal("expected policy voting=majority")
	}
	if profile.Metadata["version"] != "1.0" {
		t.Fatal("expected metadata version=1.0")
	}
}

func TestFloraBuilderMissingName(t *testing.T) {
	_, err := NewFloraBuilder().
		SetMembers([]FloraMember{{AccountID: "0.0.1", Weight: 1}}).
		SetThreshold(1).
		SetTopics(FloraTopics{Communication: "0.0.100", Transaction: "0.0.200"}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestFloraBuilderMissingMembers(t *testing.T) {
	_, err := NewFloraBuilder().
		SetDisplayName("Flora").
		SetThreshold(1).
		SetTopics(FloraTopics{Communication: "0.0.100", Transaction: "0.0.200"}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing members")
	}
}

func TestFloraBuilderThresholdZero(t *testing.T) {
	_, err := NewFloraBuilder().
		SetDisplayName("Flora").
		SetMembers([]FloraMember{{AccountID: "0.0.1", Weight: 1}}).
		SetTopics(FloraTopics{Communication: "0.0.100", Transaction: "0.0.200"}).
		Build()
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestFloraBuilderThresholdExceedsMembers(t *testing.T) {
	_, err := NewFloraBuilder().
		SetDisplayName("Flora").
		SetMembers([]FloraMember{{AccountID: "0.0.1", Weight: 1}}).
		SetThreshold(5).
		SetTopics(FloraTopics{Communication: "0.0.100", Transaction: "0.0.200"}).
		Build()
	if err == nil {
		t.Fatal("expected error for threshold > member count")
	}
}

func TestFloraBuilderMissingTopics(t *testing.T) {
	_, err := NewFloraBuilder().
		SetDisplayName("Flora").
		SetMembers([]FloraMember{{AccountID: "0.0.1", Weight: 1}}).
		SetThreshold(1).
		Build()
	if err == nil {
		t.Fatal("expected error for missing topics")
	}
}

func TestFloraBuilderAddMetadataEmptyKey(t *testing.T) {
	builder := NewFloraBuilder().SetDisplayName("Flora")
	builder.AddMetadata("", "val")
	if builder.profile.Metadata != nil && len(builder.profile.Metadata) > 0 {
		t.Fatal("expected no metadata for empty key")
	}
}

func TestCopyStringAnyMapNil(t *testing.T) {
	result := copyStringAnyMap(nil)
	if result != nil {
		t.Fatal("expected nil for nil input")
	}
}

func TestCopyStringAnyMapEmpty(t *testing.T) {
	result := copyStringAnyMap(map[string]any{})
	if result != nil {
		t.Fatal("expected nil for empty map")
	}
}

func TestCopyStringAnyMapCopy(t *testing.T) {
	input := map[string]any{"key": "value"}
	result := copyStringAnyMap(input)
	if result["key"] != "value" {
		t.Fatal("expected key=value in copy")
	}
	input["key"] = "changed"
	if result["key"] == "changed" {
		t.Fatal("copy should not be affected by mutation")
	}
}

func TestMCPServerBuilderFullSuccess(t *testing.T) {
	profile, err := NewMCPServerBuilder().
		SetName("TestServer").
		SetAlias("srv").
		SetBio("A test server").
		SetDescription("desc override").
		SetVersion("1.0.0").
		SetConnectionInfo("https://example.com/sse", "sse").
		SetServerDescription("MCP server desc").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		SetHostRequirements("0.1.0").
		SetCapabilities([]string{"tools"}).
		AddResource("res1", "resource one").
		SetResources([]MCPServerResource{{Name: "res2", Description: "resource two"}}).
		AddTool("tool1", "tool one").
		SetTools([]MCPServerTool{{Name: "tool2", Description: "tool two"}}).
		SetMaintainer("maintainer@example.com").
		SetRepository("https://github.com/example").
		SetDocs("https://docs.example.com").
		SetVerification(MCPServerVerification{Type: VerificationTypeDNS, Value: "example.com"}).
		AddVerificationDNS("example.com", "dns-field").
		AddVerificationSignature("sig123").
		AddVerificationChallenge("/challenge").
		AddSocial("twitter", "@test").
		SetSocials([]SocialLink{{Platform: "github", Handle: "test"}}).
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Type != ProfileTypeMCPServer {
		t.Fatalf("unexpected type: %v", profile.Type)
	}
	if profile.MCPServer.Version != "1.0.0" {
		t.Fatalf("unexpected version: %s", profile.MCPServer.Version)
	}
}

func TestMCPServerBuilderMissingName(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetVersion("1.0").
		SetConnectionInfo("https://example.com", "sse").
		SetServerDescription("desc").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestMCPServerBuilderMissingVersion(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetName("Srv").
		SetConnectionInfo("https://example.com", "sse").
		SetServerDescription("desc").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestMCPServerBuilderMissingURL(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetName("Srv").
		SetVersion("1.0").
		SetConnectionInfo("", "sse").
		SetServerDescription("desc").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestMCPServerBuilderMissingTransport(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetName("Srv").
		SetVersion("1.0").
		SetConnectionInfo("https://example.com", "").
		SetServerDescription("desc").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing transport")
	}
}

func TestMCPServerBuilderMissingServices(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetName("Srv").
		SetVersion("1.0").
		SetConnectionInfo("https://example.com", "sse").
		SetServerDescription("desc").
		Build()
	if err == nil {
		t.Fatal("expected error for missing services")
	}
}

func TestMCPServerBuilderMissingDescription(t *testing.T) {
	_, err := NewMCPServerBuilder().
		SetName("Srv").
		SetVersion("1.0").
		SetConnectionInfo("https://example.com", "sse").
		SetServices([]MCPServerCapability{MCPServerCapabilityToolProvider}).
		Build()
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestMCPServerBuilderAddSocialEmpty(t *testing.T) {
	builder := NewMCPServerBuilder().SetName("Srv")
	builder.AddSocial("", "handle")
	builder.AddSocial("platform", "")
	if len(builder.profile.Socials) != 0 {
		t.Fatalf("expected no socials for empty inputs")
	}
}

func TestMCPServerBuilderAddSocialDuplicate(t *testing.T) {
	builder := NewMCPServerBuilder().SetName("Srv").
		AddSocial("twitter", "@old").
		AddSocial("twitter", "@new")
	if len(builder.profile.Socials) != 1 {
		t.Fatalf("expected 1 social after dedup, got %d", len(builder.profile.Socials))
	}
	if builder.profile.Socials[0].Handle != "@new" {
		t.Fatalf("expected updated handle '@new', got %q", builder.profile.Socials[0].Handle)
	}
}
