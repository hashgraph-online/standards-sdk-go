package registrybroker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStartConversationFallsBackToPlaintextWhenEncryptionPreferred(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/chat/session" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("content-type", "application/json")
		_, _ = writer.Write([]byte(`{"sessionId":"session-1","encryption":{"enabled":false}}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handle, err := client.StartConversation(context.Background(), StartConversationOptions{
		UAID: "uaid:aid:test",
		Encryption: &ConversationEncryptionOptions{
			Preference: "preferred",
		},
	})
	if err != nil {
		t.Fatalf("start conversation failed: %v", err)
	}
	if handle == nil {
		t.Fatalf("expected plaintext handle")
	}
	if handle.Mode != "plaintext" {
		t.Fatalf("expected plaintext mode, got %s", handle.Mode)
	}
}

func TestStartConversationFailsWhenEncryptionRequired(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/api/v1/chat/session" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("content-type", "application/json")
		_, _ = writer.Write([]byte(`{"sessionId":"session-2","encryption":{"enabled":false}}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.StartConversation(context.Background(), StartConversationOptions{
		UAID: "uaid:aid:test",
		Encryption: &ConversationEncryptionOptions{
			Preference: "required",
		},
	})
	if err == nil {
		t.Fatalf("expected encryption required error")
	}
	var unavailable *EncryptionUnavailableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected EncryptionUnavailableError, got %T", err)
	}
}

func TestCreateEncryptedSessionAndSendMessage(t *testing.T) {
	t.Parallel()

	bootstrapClient, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{})
	if err != nil {
		t.Fatalf("failed to bootstrap client: %v", err)
	}
	responderKeyPair, err := bootstrapClient.GenerateEphemeralKeyPair()
	if err != nil {
		t.Fatalf("failed to generate responder keypair: %v", err)
	}

	var messagePayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/chat/session":
			_, _ = writer.Write([]byte(`{
				"sessionId":"session-3",
				"encryption":{
					"enabled":true,
					"requester":{"uaid":"uaid:aid:requester"},
					"responder":{"uaid":"uaid:aid:responder"}
				}
			}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/chat/session/session-3/encryption-handshake":
			_, _ = writer.Write([]byte(`{"handshake":{"status":"pending"}}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/chat/session/session-3/encryption":
			_, _ = writer.Write([]byte(`{
				"sessionId":"session-3",
				"encryption":{
					"enabled":true,
					"requester":{"uaid":"uaid:aid:requester"},
					"responder":{"uaid":"uaid:aid:responder"},
					"handshake":{
						"status":"complete",
						"requester":{"ephemeralPublicKey":"02ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
						"responder":{"ephemeralPublicKey":"` + responderKeyPair.PublicKey + `"}
					}
				}
			}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/chat/message":
			if err := json.NewDecoder(request.Body).Decode(&messagePayload); err != nil {
				t.Fatalf("failed to decode message payload: %v", err)
			}
			_, _ = writer.Write([]byte(`{"ok":true}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handle, err := client.CreateEncryptedSession(context.Background(), StartEncryptedChatSessionOptions{
		UAID: "uaid:aid:responder",
	})
	if err != nil {
		t.Fatalf("create encrypted session failed: %v", err)
	}
	if handle.Mode != "encrypted" {
		t.Fatalf("expected encrypted mode, got %s", handle.Mode)
	}

	_, err = handle.Send(context.Background(), "hello world", nil, nil)
	if err != nil {
		t.Fatalf("send encrypted message failed: %v", err)
	}

	if messagePayload == nil {
		t.Fatalf("expected chat message payload to be captured")
	}
	cipherEnvelope, ok := messagePayload["cipherEnvelope"].(map[string]any)
	if !ok || len(cipherEnvelope) == 0 {
		t.Fatalf("expected cipherEnvelope in message payload, got %#v", messagePayload)
	}
	if strings.TrimSpace(stringField(messagePayload, "message")) != "[ciphertext omitted]" {
		t.Fatalf("expected placeholder message for encrypted payload, got %#v", messagePayload["message"])
	}
}
