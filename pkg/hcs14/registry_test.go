package hcs14

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDIDResolver struct {
	document *DIDDocument
}

func (resolver *mockDIDResolver) Supports(did string) bool {
	return resolver.document != nil && resolver.document.ID == did
}

func (resolver *mockDIDResolver) Resolve(_ context.Context, did string) (*DIDDocument, error) {
	if resolver.document != nil && resolver.document.ID == did {
		return resolver.document, nil
	}
	return nil, nil
}

func TestResolverRegistryResolveAIDDNSProfile(t *testing.T) {
	registry := NewResolverRegistry()
	registry.RegisterUAIDProfileResolver(NewAIDDNSWebResolver(AIDDNSWebResolverOptions{
		DNSLookup: func(_ context.Context, hostname string) ([]string, error) {
			if hostname != "_agent.agent.example.com" {
				t.Fatalf("unexpected hostname: %s", hostname)
			}
			return []string{
				"v=aid1; p=hcs-10; u=https://agent.example.com/.well-known/agent.json",
			}, nil
		},
	}))

	uaid := "uaid:aid:QmAid123;uid=support;proto=a2a;nativeId=agent.example.com"
	profile, err := registry.ResolveUAIDProfile(context.Background(), uaid, ResolveUaidProfileOptions{
		ProfileID: AIDDNSWebProfileID,
	})
	if err != nil {
		t.Fatalf("ResolveUAIDProfile failed: %v", err)
	}
	if profile == nil || profile.Error != nil {
		t.Fatalf("unexpected profile result: %+v", profile)
	}
	if profile.Metadata.Profile != AIDDNSWebProfileID {
		t.Fatalf("unexpected profile ID: %s", profile.Metadata.Profile)
	}
	if profile.Metadata.Endpoint != "https://agent.example.com/.well-known/agent.json" {
		t.Fatalf("unexpected endpoint: %s", profile.Metadata.Endpoint)
	}
}

func TestResolverRegistryResolveUAIDDidResolutionProfile(t *testing.T) {
	baseDID := "did:key:z6MkhaXgBZDvotDkL5257f"
	src := "z" + base58Encode([]byte(baseDID))

	registry := NewResolverRegistry()
	registry.RegisterDIDResolver(&mockDIDResolver{
		document: &DIDDocument{
			ID:                 baseDID,
			VerificationMethod: []DIDVerificationMethod{{ID: baseDID + "#key-1", Type: "Ed25519VerificationKey2020", Controller: baseDID}},
			Authentication:     []string{baseDID + "#key-1"},
		},
	})
	registry.RegisterUAIDProfileResolver(NewUAIDDidResolutionResolver())

	uaid := "uaid:did:z6MkhaXgBZDvotDkL5257f;uid=0;proto=hcs-10;nativeId=agent.example.com;src=" + src
	profile, err := registry.ResolveUAIDProfile(context.Background(), uaid, ResolveUaidProfileOptions{
		ProfileID: UAIDDidResolutionProfileID,
	})
	if err != nil {
		t.Fatalf("ResolveUAIDProfile failed: %v", err)
	}
	if profile == nil || profile.Error != nil {
		t.Fatalf("unexpected profile result: %+v", profile)
	}
	if profile.DID != baseDID {
		t.Fatalf("unexpected DID: %s", profile.DID)
	}
	if profile.Metadata.Profile != UAIDDidResolutionProfileID {
		t.Fatalf("unexpected profile ID: %s", profile.Metadata.Profile)
	}
	if !profile.Metadata.Resolved || !profile.Metadata.BaseDIDResolved {
		t.Fatalf("expected resolved metadata")
	}
}

func TestResolverRegistryResolveUAIDDidResolutionProfileRequiresBaseDID(t *testing.T) {
	registry := NewResolverRegistry()
	registry.RegisterUAIDProfileResolver(NewUAIDDidResolutionResolver())

	uaid := "uaid:did:opaque-id;uid=0"
	profile, err := registry.ResolveUAIDProfile(context.Background(), uaid, ResolveUaidProfileOptions{
		ProfileID: UAIDDidResolutionProfileID,
	})
	if err != nil {
		t.Fatalf("ResolveUAIDProfile failed: %v", err)
	}
	if profile == nil || profile.Error == nil {
		t.Fatalf("expected error profile, got: %+v", profile)
	}
	if profile.Error.Code != "ERR_BASE_DID_UNDETERMINED" {
		t.Fatalf("unexpected error code: %s", profile.Error.Code)
	}
}

func TestClientResolveANSViaRegistryFlow(t *testing.T) {
	agentCard := map[string]any{
		"ansName": "ans://v1.0.1.ote.agent.cs3p.com",
		"endpoints": map[string]any{
			"a2a": map[string]any{"url": "https://ote.agent.cs3p.com/a2a/jsonrpc"},
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := json.NewEncoder(writer).Encode(agentCard); err != nil {
			t.Fatalf("failed to encode payload: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(ClientOptions{
		DNSLookup: func(_ context.Context, hostname string) ([]string, error) {
			switch hostname {
			case "_uaid.ote.agent.cs3p.com":
				return []string{
					"target=aid;id=ans-godaddy-ote;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1",
				}, nil
			case "_ans.ote.agent.cs3p.com":
				return []string{
					"v=ans1;version=1.0.1;url=" + server.URL,
				}, nil
			default:
				return []string{}, nil
			}
		},
		HTTP: ANSDNSWebResolverOptions{
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		},
	})

	uaid := "uaid:aid:ans-godaddy-ote;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1"
	result, err := client.Resolve(context.Background(), uaid)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result == nil || result.Error != nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Metadata.Profile != UAIDDNSWebProfileID {
		t.Fatalf("unexpected profile: %s", result.Metadata.Profile)
	}
	if result.Metadata.SelectedFollowupProfile != ANSDNSWebProfileID {
		t.Fatalf("unexpected selected follow-up profile: %s", result.Metadata.SelectedFollowupProfile)
	}
	if result.Metadata.Endpoint != "https://ote.agent.cs3p.com/a2a/jsonrpc" {
		t.Fatalf("unexpected endpoint: %s", result.Metadata.Endpoint)
	}
}
