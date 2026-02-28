package hcs14

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
)

const defaultANSIntegrationUAID = "uaid:aid:ans-godaddy-ote;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1"

func TestHCS14Integration_ANSDNSWebResolution(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}

	inputUAID := strings.TrimSpace(os.Getenv("HCS14_INTEGRATION_UAID"))
	if inputUAID == "" {
		inputUAID = defaultANSIntegrationUAID
	}

	client := NewClient(ClientOptions{})
	result, err := client.Resolve(context.Background(), inputUAID)
	if err != nil {
		t.Fatalf("failed to resolve UAID: %v", err)
	}
	if result == nil {
		t.Fatalf("resolve returned nil result")
	}
	if result.Error != nil {
		t.Fatalf("resolve returned error: %s (%s)", result.Error.Message, result.Error.Code)
	}
	if !result.Metadata.Resolved {
		t.Fatalf("expected resolved metadata")
	}
	if strings.TrimSpace(result.Metadata.Endpoint) == "" {
		t.Fatalf("expected endpoint in metadata")
	}

	parsedUAID, err := ParseUAID(inputUAID)
	if err != nil {
		t.Fatalf("failed to parse input UAID: %v", err)
	}
	expectedHost := normalizeDomain(parsedUAID.Params["nativeId"])

	endpointURL, err := url.Parse(result.Metadata.Endpoint)
	if err != nil {
		t.Fatalf("resolved endpoint is not a valid URL: %v", err)
	}
	if normalizeDomain(endpointURL.Hostname()) != expectedHost {
		t.Fatalf("resolved endpoint host mismatch: got %s expected %s", endpointURL.Hostname(), expectedHost)
	}

	t.Logf("resolved profile=%s endpoint=%s agentCardUrl=%s", result.Metadata.Profile, result.Metadata.Endpoint, result.Metadata.AgentCardURL)
}
