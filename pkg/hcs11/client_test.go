package hcs11

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProfileBuilders(t *testing.T) {
	agentBuilder := NewAgentBuilder().
		SetName("Support Agent").
		SetType(AIAgentTypeAutonomous).
		SetCapabilities([]AIAgentCapability{AIAgentCapabilityTextGeneration}).
		SetModel("gpt-4o")

	agentProfile, err := agentBuilder.Build()
	if err != nil {
		t.Fatalf("agent builder failed: %v", err)
	}
	if agentProfile.Type != ProfileTypeAIAgent {
		t.Fatalf("unexpected agent profile type: %d", agentProfile.Type)
	}

	personProfile, err := NewPersonBuilder().
		SetName("Jane Doe").
		SetBio("profile").Build()
	if err != nil {
		t.Fatalf("person builder failed: %v", err)
	}
	if personProfile.Type != ProfileTypePersonal {
		t.Fatalf("unexpected person profile type: %d", personProfile.Type)
	}

	floraProfile, err := NewFloraBuilder().
		SetDisplayName("Flora").
		SetMembers([]FloraMember{{AccountID: "0.0.1001"}}).
		SetThreshold(1).
		SetTopics(FloraTopics{
			Communication: "0.0.2001",
			Transaction:   "0.0.2002",
			State:         "0.0.2003",
		}).
		Build()
	if err != nil {
		t.Fatalf("flora builder failed: %v", err)
	}
	if floraProfile.Type != ProfileTypeFlora {
		t.Fatalf("unexpected flora profile type: %d", floraProfile.Type)
	}
}

func TestClientValidateAndSerializeProfile(t *testing.T) {
	client, err := NewClient(ClientConfig{Network: "testnet"})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	profile, err := client.CreateAIAgentProfile(
		"Agent",
		AIAgentTypeManual,
		[]AIAgentCapability{AIAgentCapabilityTextGeneration},
		"gpt-4o",
		nil,
	)
	if err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	validation := client.ValidateProfile(profile)
	if !validation.Valid {
		t.Fatalf("profile should be valid: %+v", validation)
	}

	encoded, err := client.ProfileToJSONString(profile)
	if err != nil {
		t.Fatalf("profile serialization failed: %v", err)
	}
	decoded, err := client.ParseProfileFromString(encoded)
	if err != nil {
		t.Fatalf("profile parse failed: %v", err)
	}
	if decoded.DisplayName != "Agent" {
		t.Fatalf("unexpected decoded display name: %s", decoded.DisplayName)
	}
}

func TestClientFetchProfileByAccountID(t *testing.T) {
	profile := HCS11Profile{
		Version:         "1.0",
		Type:            ProfileTypeAIAgent,
		DisplayName:     "Agent",
		InboundTopicID:  "0.0.1001",
		OutboundTopicID: "0.0.1002",
		AIAgent: &AIAgentDetails{
			Type:         AIAgentTypeManual,
			Capabilities: []AIAgentCapability{AIAgentCapabilityTextGeneration},
			Model:        "gpt-4o",
		},
	}
	profileBytes, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case strings.HasPrefix(request.URL.Path, "/api/v1/accounts/"):
			_, _ = writer.Write([]byte(`{"account":"0.0.1234","memo":"hcs-11:hcs://1/0.0.987654"}`))
		case strings.HasPrefix(request.URL.Path, "/api/inscription-cdn/0.0.987654"):
			_, _ = writer.Write(profileBytes)
		default:
			writer.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		Network:           "testnet",
		MirrorBaseURL:     server.URL,
		KiloScribeBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	response, err := client.FetchProfileByAccountID(context.Background(), "0.0.1234", "testnet")
	if err != nil {
		t.Fatalf("FetchProfileByAccountID failed: %v", err)
	}
	if !response.Success {
		t.Fatalf("fetch failed: %s", response.Error)
	}
	if response.Profile == nil || response.Profile.DisplayName != "Agent" {
		t.Fatalf("unexpected fetched profile: %+v", response.Profile)
	}
	if response.TopicInfo == nil || response.TopicInfo.ProfileTopicID != "0.0.987654" {
		t.Fatalf("unexpected topic info: %+v", response.TopicInfo)
	}
}

func TestClientGetCapabilitiesFromTags(t *testing.T) {
	client, err := NewClient(ClientConfig{Network: "testnet"})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}
	capabilities := client.GetCapabilitiesFromTags([]string{"text_generation"})
	if len(capabilities) != 1 || capabilities[0] != int(AIAgentCapabilityTextGeneration) {
		t.Fatalf("unexpected capabilities: %+v", capabilities)
	}
}

func TestClientGetCapabilitiesFromTagsExpanded(t *testing.T) {
	client, err := NewClient(ClientConfig{Network: "testnet"})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	capabilities := client.GetCapabilitiesFromTags([]string{
		"image_generation",
		"workflow_automation",
		"unknown",
	})
	if len(capabilities) != 2 {
		t.Fatalf("expected two mapped capabilities, got %+v", capabilities)
	}
	if capabilities[0] != int(AIAgentCapabilityImageGeneration) {
		t.Fatalf("expected image generation capability, got %d", capabilities[0])
	}
	if capabilities[1] != int(AIAgentCapabilityWorkflowAutomation) {
		t.Fatalf("expected workflow automation capability, got %d", capabilities[1])
	}
}

func TestClientValidateProfileRejectsInvalidMCPTransport(t *testing.T) {
	client, err := NewClient(ClientConfig{Network: "testnet"})
	if err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	profile, createErr := client.CreateMCPServerProfile(
		"MCP",
		MCPServerDetails{
			Version: "2026-01-01",
			ConnectionInfo: MCPServerConnectionInfo{
				URL:       "https://example.com",
				Transport: "invalid",
			},
			Services:    []MCPServerCapability{MCPServerCapabilityToolProvider},
			Description: "server",
		},
		nil,
	)
	if createErr == nil {
		t.Fatalf("expected profile creation to fail due to invalid transport, got profile %+v", profile)
	}
}
