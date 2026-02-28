package hcs14

import (
	"context"
	"testing"
)

func TestClientResolveEndpoints(t *testing.T) {
	client := NewClient(ClientOptions{})

	ctx := context.Background()

	// Registry getter
	if client.Registry() == nil {
		t.Fatal("expected registry")
	}
	if client.UaidDNSResolver() == nil {
		t.Fatal("expected uaid dns resolver")
	}
	if client.AnsDNSResolver() == nil {
		t.Fatal("expected ans dns resolver")
	}
	if client.AidDNSResolver() == nil {
		t.Fatal("expected aid dns resolver")
	}
	if client.UaidDidResolver() == nil {
		t.Fatal("expected uaid did resolver")
	}

	// ResolveProfile
	_, _ = client.ResolveProfile(ctx, "uaid:test", "test-profile")
	_, _ = client.ResolveUAIDDNSWeb(ctx, "uaid:test")
	_, _ = client.ResolveANSDNSWeb(ctx, "uaid:test")
	_, _ = client.ResolveAIDDNSWeb(ctx, "uaid:test")
	_, _ = client.ResolveUAIDDidResolution(ctx, "uaid:test")
}
