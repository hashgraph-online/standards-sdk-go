package registrybroker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestInitializeAgentEnsuresEncryptionKeyByDefault(t *testing.T) {
	t.Parallel()

	var keyRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && request.URL.Path == "/api/v1/encryption/keys" {
			atomic.AddInt32(&keyRequests, 1)
			writer.Header().Set("content-type", "application/json")
			_, _ = writer.Write([]byte(`{"publicKey":"pub-key-1"}`))
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	result, err := InitializeAgent(context.Background(), InitializeAgentClientOptions{
		RegistryBrokerClientOptions: RegistryBrokerClientOptions{
			BaseURL: server.URL,
		},
		UAID: "uaid:aid:test",
	})
	if err != nil {
		t.Fatalf("initialize agent failed: %v", err)
	}
	if result == nil || result.Client == nil {
		t.Fatalf("expected initialized client result")
	}
	if atomic.LoadInt32(&keyRequests) != 1 {
		t.Fatalf("expected one key registration request, got %d", keyRequests)
	}
}

func TestInitializeAgentCanSkipEnsureEncryptionKey(t *testing.T) {
	t.Parallel()

	var keyRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost && request.URL.Path == "/api/v1/encryption/keys" {
			atomic.AddInt32(&keyRequests, 1)
		}
		writer.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	disabled := false
	result, err := InitializeAgent(context.Background(), InitializeAgentClientOptions{
		RegistryBrokerClientOptions: RegistryBrokerClientOptions{
			BaseURL: server.URL,
		},
		UAID:                "uaid:aid:test",
		EnsureEncryptionKey: &disabled,
	})
	if err != nil {
		t.Fatalf("initialize agent failed: %v", err)
	}
	if result == nil || result.Client == nil {
		t.Fatalf("expected initialized client result")
	}
	if atomic.LoadInt32(&keyRequests) != 0 {
		t.Fatalf("expected zero key registration requests, got %d", keyRequests)
	}
}

func TestRegisterAgentResponseHelpers(t *testing.T) {
	t.Parallel()

	if !IsPendingRegisterAgentResponse(JSONObject{"status": "pending"}) {
		t.Fatalf("expected pending response helper to match")
	}
	if !IsPartialRegisterAgentResponse(JSONObject{"status": "partial", "success": false}) {
		t.Fatalf("expected partial response helper to match")
	}
	if !IsSuccessRegisterAgentResponse(JSONObject{"status": "completed", "success": true}) {
		t.Fatalf("expected success response helper to match")
	}
}
