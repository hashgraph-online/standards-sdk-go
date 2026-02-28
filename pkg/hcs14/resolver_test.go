package hcs14

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUaidDNSWebResolverBindingOnly(t *testing.T) {
	uaid := "uaid:aid:test123;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1"
	resolver := NewUaidDNSWebResolver(UaidDNSWebResolverOptions{
		DNSLookup: func(ctx context.Context, hostname string) ([]string, error) {
			if hostname != "_uaid.ote.agent.cs3p.com" {
				t.Fatalf("unexpected hostname: %s", hostname)
			}
			return []string{
				"target=aid;id=test123;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1;m=demo",
			}, nil
		},
		EnableFollowupResolution: false,
	})

	result, err := resolver.Resolve(context.Background(), uaid, nil)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected resolver error: %+v", result.Error)
	}
	if !result.Metadata.Resolved {
		t.Fatalf("expected resolved result")
	}
	if result.Metadata.Profile != UAIDDNSWebProfileID {
		t.Fatalf("unexpected profile: %s", result.Metadata.Profile)
	}
}

func TestANSDNSWebResolver(t *testing.T) {
	agentCard := map[string]any{
		"ansName": "ans://v1.0.1.ote.agent.cs3p.com",
		"endpoints": map[string]any{
			"a2a": map[string]any{"url": "https://ote.agent.cs3p.com/a2a/jsonrpc"},
			"mcp": map[string]any{"url": "https://ote.agent.cs3p.com/mcp"},
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(agentCard)
	}))
	defer server.Close()

	serverURL := server.URL
	resolver := NewANSDNSWebResolver(ANSDNSWebResolverOptions{
		DNSLookup: func(ctx context.Context, hostname string) ([]string, error) {
			if hostname != "_ans.ote.agent.cs3p.com" {
				t.Fatalf("unexpected hostname: %s", hostname)
			}
			return []string{
				fmt.Sprintf("v=ans1;version=1.0.1;url=%s", serverURL),
			}, nil
		},
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		SupportedSchemes: []string{"https"},
	})

	uaid := "uaid:aid:test123;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=a2a;nativeId=ote.agent.cs3p.com;version=1.0.1"
	result, err := resolver.Resolve(context.Background(), uaid)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected resolver error: %+v", result.Error)
	}
	if !result.Metadata.Resolved {
		t.Fatalf("expected resolved result")
	}
	if result.Metadata.Endpoint != "https://ote.agent.cs3p.com/a2a/jsonrpc" {
		t.Fatalf("unexpected endpoint selected: %s", result.Metadata.Endpoint)
	}
}

func TestANSDNSWebResolverCompatibilityCard(t *testing.T) {
	agentCard := map[string]any{
		"url": "https://ote.agent.cs3p.com/a2a/jsonrpc",
		"additionalInterfaces": []any{
			map[string]any{"url": "https://ote.agent.cs3p.com/mcp"},
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(agentCard)
	}))
	defer server.Close()

	resolver := NewANSDNSWebResolver(ANSDNSWebResolverOptions{
		DNSLookup: func(ctx context.Context, hostname string) ([]string, error) {
			return []string{
				fmt.Sprintf("v=ans1;version=1.0.1;url=%s", server.URL),
			}, nil
		},
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		SupportedSchemes: []string{"https"},
	})

	uaid := "uaid:aid:test123;uid=ans://v1.0.1.ote.agent.cs3p.com;registry=ans;proto=mcp;nativeId=ote.agent.cs3p.com;version=1.0.1"
	result, err := resolver.Resolve(context.Background(), uaid)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("unexpected resolver error: %+v", result.Error)
	}
	if result.Metadata.Endpoint != "https://ote.agent.cs3p.com/mcp" {
		t.Fatalf("unexpected endpoint selected: %s", result.Metadata.Endpoint)
	}
}
