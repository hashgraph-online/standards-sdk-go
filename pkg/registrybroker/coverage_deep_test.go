package registrybroker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	e1 := &RegistryBrokerError{Message: "bad", Status: 400, StatusText: "Bad Request"}
	if e1.Error() == "" {
		t.Fatal("expected error string")
	}

	e2 := &RegistryBrokerError{Message: "bad"}
	if e2.Error() != "bad" {
		t.Fatal("expected bad")
	}

	var e3 *RegistryBrokerError
	if e3.Error() != "registry broker request failed" {
		t.Fatal("expected nil error")
	}

	pe1 := &RegistryBrokerParseError{Message: "parse", Cause: fmt.Errorf("inner")}
	if pe1.Error() == "" {
		t.Fatal("expected error")
	}

	pe2 := &RegistryBrokerParseError{Message: "parse"}
	if pe2.Error() != "parse" {
		t.Fatal("expected parse")
	}

	var pe3 *RegistryBrokerParseError
	if pe3.Error() != "registry broker parse error" {
		t.Fatal("expected nil error")
	}
}

func TestEncryptionHelpers(t *testing.T) {
	_, ok := parseCipherEnvelope("not a map")
	if ok {
		t.Fatal("expected false")
	}

	env, ok := parseCipherEnvelope(map[string]any{
		"algorithm":  "aes-256-gcm",
		"ciphertext": "abc",
		"nonce":      "def",
		"keyLocator": map[string]any{"sessionId": "s1"},
		"recipients": []any{
			map[string]any{"uaid": "u1", "email": "e@e.com"},
			"not-a-map",
		},
	})
	if !ok {
		t.Fatal("expected true")
	}
	if env.Algorithm != "aes-256-gcm" {
		t.Fatal("expected algo")
	}

	_, ok2 := parseCipherEnvelope(map[string]any{"algorithm": "x"})
	if ok2 {
		t.Fatal("expected false for missing fields")
	}

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{})
	kp, err := client.GenerateEphemeralKeyPair()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	pub, err := derivePublicKeyFromPrivateKey(kp.PrivateKey)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if pub != kp.PublicKey {
		t.Fatal("expected matching public key")
	}

	_, err = derivePublicKeyFromPrivateKey("invalid-hex")
	if err == nil {
		t.Fatal("expected err")
	}

	result := serialiseAuthConfig(nil)
	if len(result) != 0 {
		t.Fatal("expected empty")
	}

	result2 := serialiseAuthConfig(&AgentAuthConfig{
		Type: "bearer", Token: "tok", Username: "u", Password: "p",
		HeaderName: "h", HeaderValue: "v",
		Headers: map[string]string{"X-Key": "val"},
	})
	if result2["type"] != "bearer" {
		t.Fatal("expected bearer")
	}
	if result2["headers"] == nil {
		t.Fatal("expected headers")
	}

	m := map[string]any{"s": "str", "f": 1.5, "i": 42, "n": nil, "other": []int{}}
	if mapStringField(m, "s") != "str" {
		t.Fatal("expected str")
	}
	if mapStringField(m, "f") != "1.5" {
		t.Fatal("expected 1.5")
	}
	if mapStringField(m, "missing") != "" {
		t.Fatal("expected empty")
	}
	if mapStringField(m, "n") != "" {
		t.Fatal("expected empty")
	}
	if mapStringField(m, "other") != "" {
		t.Fatal("expected empty")
	}
}

func TestFeedbackIndexEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true, "data": []}`))
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: ts.URL})
	ctx := context.Background()

	_, _ = client.ListAgentFeedbackIndex(ctx, AgentFeedbackIndexOptions{})
	_, _ = client.ListAgentFeedbackEntriesIndex(ctx, AgentFeedbackIndexOptions{})
	_, _ = client.GetX402Minimums(ctx)
	_, _ = client.SkillsConfig(ctx)
}

func TestChatStartAndHistory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{"success": true, "data": map[string]any{"sessionId": "s1"}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{BaseURL: ts.URL})
	ctx := context.Background()

	_, _ = client.StartChat(ctx, StartChatOptions{UAID: "test"})
	_, _ = client.StartChat(ctx, StartChatOptions{AgentURL: "http://example.com"})
	_, _ = client.StartChat(ctx, StartChatOptions{})

	_, _ = client.AcceptConversation(ctx, AcceptConversationOptions{SessionID: "s1"})
	_, _ = client.AcceptConversation(ctx, AcceptConversationOptions{})

	handle := client.CreatePlaintextConversationHandle("s1", nil, nil, "uaid", "")
	_, _ = handle.FetchHistory(ctx, ChatHistoryFetchOptions{})
}

func TestIdentitiesMatch(t *testing.T) {
	if identitiesMatch(nil, nil) != true {
		t.Fatal("expected match")
	}
	if identitiesMatch(&RecipientIdentity{UAID: "a"}, nil) {
		t.Fatal("expected mismatch")
	}
	if !identitiesMatch(&RecipientIdentity{UAID: "a"}, &RecipientIdentity{UAID: "a"}) {
		t.Fatal("expected match")
	}
	if identitiesMatch(&RecipientIdentity{UAID: "a"}, &RecipientIdentity{UAID: "b"}) {
		t.Fatal("expected mismatch")
	}
	if !identitiesMatch(&RecipientIdentity{LedgerAccountID: "x"}, &RecipientIdentity{LedgerAccountID: "x"}) {
		t.Fatal("expected match")
	}
	if !identitiesMatch(&RecipientIdentity{UserID: "u"}, &RecipientIdentity{UserID: "u"}) {
		t.Fatal("expected match")
	}
	if !identitiesMatch(&RecipientIdentity{Email: "e@e.com"}, &RecipientIdentity{Email: "e@e.com"}) {
		t.Fatal("expected match")
	}
	if identitiesMatch(&RecipientIdentity{}, &RecipientIdentity{}) {
		t.Fatal("expected mismatch empty")
	}
}

func TestExtractInsufficientCreditsDetails(t *testing.T) {
	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{})
	_, ok := client.extractInsufficientCreditsDetails(fmt.Errorf("generic error"))
	if ok {
		t.Fatal("expected false")
	}

	_, ok2 := client.extractInsufficientCreditsDetails(&RegistryBrokerError{
		Status: http.StatusPaymentRequired,
		Body:   map[string]any{"shortfallCredits": 100.0},
	})
	if !ok2 {
		t.Fatal("expected true")
	}
}

func TestDecryptHistoryEntryFromContext(t *testing.T) {
	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{})

	// nil entry
	result := client.DecryptHistoryEntryFromContext(nil, nil)
	if result != nil {
		t.Fatal("expected nil")
	}

	// no context, has content
	entry := JSONObject{"content": "hello"}
	result2 := client.DecryptHistoryEntryFromContext(entry, nil)
	if result2 == nil || *result2 != "hello" {
		t.Fatal("expected hello")
	}

	// no context, empty content
	entry2 := JSONObject{"content": ""}
	result3 := client.DecryptHistoryEntryFromContext(entry2, nil)
	if result3 != nil {
		t.Fatal("expected nil")
	}

	// with context but no cipherEnvelope
	ctx := &ConversationContextState{SharedSecret: make([]byte, 32)}
	entry3 := JSONObject{"content": "plain"}
	result4 := client.DecryptHistoryEntryFromContext(entry3, ctx)
	if result4 == nil || *result4 != "plain" {
		t.Fatal("expected plain")
	}
}

func TestResolveDecryptionContext(t *testing.T) {
	client, _ := NewRegistryBrokerClient(RegistryBrokerClientOptions{})

	// with shared secret in options
	ctx := client.ResolveDecryptionContext("s1", ChatHistoryFetchOptions{
		SharedSecret: []byte("secret"),
	})
	if ctx == nil {
		t.Fatal("expected ctx")
	}

	// no registered contexts
	ctx2 := client.ResolveDecryptionContext("s2", ChatHistoryFetchOptions{})
	if ctx2 != nil {
		t.Fatal("expected nil")
	}

	// register a context
	client.RegisterConversationContextForEncryption(ConversationContextInput{
		SessionID:    "s3",
		SharedSecret: []byte("secret"),
		Identity:     &RecipientIdentity{UAID: "u1"},
	})

	ctx3 := client.ResolveDecryptionContext("s3", ChatHistoryFetchOptions{
		Identity: &RecipientIdentity{UAID: "u1"},
	})
	if ctx3 == nil {
		t.Fatal("expected non-nil context")
	}

	// Without identity
	ctx4 := client.ResolveDecryptionContext("s3", ChatHistoryFetchOptions{})
	if ctx4 == nil {
		t.Fatal("expected non-nil context fallback")
	}
}
