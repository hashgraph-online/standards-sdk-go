package registrybroker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSearchBuildsExpectedQuery(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", request.Method)
		}
		if !strings.HasSuffix(request.URL.Path, "/search") {
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("q") != "agent" {
			t.Fatalf("expected q=agent, got %s", query.Get("q"))
		}
		if query.Get("registry") != "ans" {
			t.Fatalf("expected registry=ans, got %s", query.Get("registry"))
		}
		writer.Header().Set("content-type", "application/json")
		_, _ = writer.Write([]byte(`{"hits":[],"total":0,"limit":20,"page":1}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.Search(context.Background(), SearchParams{
		Q:        "agent",
		Registry: "ans",
		Limit:    20,
		Page:     1,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
}

func TestRegisterAgentAutoTopUpFlow(t *testing.T) {
	t.Parallel()

	var mutex sync.Mutex
	callOrder := []string{}
	quoteCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		callOrder = append(callOrder, request.Method+" "+request.URL.Path)
		mutex.Unlock()

		writer.Header().Set("content-type", "application/json")

		switch {
		case strings.HasSuffix(request.URL.Path, "/register/quote"):
			quoteCount++
			if quoteCount == 1 {
				_, _ = writer.Write([]byte(`{"shortfallCredits":5,"creditsPerHbar":10,"requiredCredits":5}`))
				return
			}
			_, _ = writer.Write([]byte(`{"shortfallCredits":0,"creditsPerHbar":10,"requiredCredits":5}`))
		case strings.HasSuffix(request.URL.Path, "/credits/purchase"):
			_, _ = writer.Write([]byte(`{"success":true}`))
		case strings.HasSuffix(request.URL.Path, "/register"):
			_, _ = writer.Write([]byte(`{"status":"completed","success":true}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.RegisterAgent(context.Background(), AgentRegistrationRequest{
		"profile": JSONObject{
			"display_name": "go-sdk-test",
		},
	}, &RegisterAgentOptions{
		AutoTopUp: &AutoTopUpOptions{
			AccountID:  "0.0.123",
			PrivateKey: "302e020100300506032b6570042204200000000000000000000000000000000000000000000000000000000000000001",
		},
	})
	if err != nil {
		t.Fatalf("register agent failed: %v", err)
	}

	if len(callOrder) < 4 {
		t.Fatalf("expected at least 4 calls, got %d", len(callOrder))
	}
}

func TestCreateSessionHistoryAutoTopUp(t *testing.T) {
	t.Parallel()

	sessionAttempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")

		switch {
		case strings.HasSuffix(request.URL.Path, "/chat/session"):
			sessionAttempts++
			if sessionAttempts == 1 {
				writer.WriteHeader(http.StatusPaymentRequired)
				_, _ = writer.Write([]byte(`{"message":"chat history credits required"}`))
				return
			}
			_, _ = writer.Write([]byte(`{"sessionId":"session-1"}`))
		case strings.HasSuffix(request.URL.Path, "/credits/purchase"):
			_, _ = writer.Write([]byte(`{"success":true}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
		HistoryAutoTop: &HistoryAutoTopUpOptions{
			AccountID:  "0.0.123",
			PrivateKey: "302e020100300506032b6570042204200000000000000000000000000000000000000000000000000000000000000001",
		},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.CreateSession(context.Background(), CreateSessionRequestPayload{
		UAID:              "uaid:aid:example",
		HistoryTTLSeconds: intPointer(3600),
	}, true)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	if result["sessionId"] != "session-1" {
		t.Fatalf("unexpected session result: %#v", result)
	}
}

func TestWaitForRegistrationCompletion(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		if !strings.Contains(request.URL.Path, "/register/progress/") {
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"error":"not found"}`))
			return
		}
		calls++
		if calls == 1 {
			_, _ = writer.Write([]byte(`{"progress":{"status":"pending"}}`))
			return
		}
		_, _ = writer.Write([]byte(`{"progress":{"status":"completed","attemptId":"a1"}}`))
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.WaitForRegistrationCompletion(context.Background(), "a1", RegistrationProgressWaitOptions{
		Interval: 25 * time.Millisecond,
		Timeout:  2 * time.Second,
	})
	if err != nil {
		t.Fatalf("wait for registration failed: %v", err)
	}
	if result["status"] != "completed" {
		t.Fatalf("unexpected progress status: %#v", result)
	}
}

func TestLedgerAuthenticateFlow(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		switch {
		case strings.HasSuffix(request.URL.Path, "/auth/ledger/challenge"):
			_, _ = writer.Write([]byte(`{"challengeId":"c1","message":"sign-me"}`))
		case strings.HasSuffix(request.URL.Path, "/auth/ledger/verify"):
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode verify payload: %v", err)
			}
			if body["signature"] != "sig" {
				t.Fatalf("unexpected signature payload: %#v", body)
			}
			_, _ = writer.Write([]byte(`{"key":"api-key-value","accountId":"0.0.123"}`))
		default:
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer server.Close()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.AuthenticateWithLedger(context.Background(), LedgerAuthenticationOptions{
		AccountID: "0.0.123",
		Network:   "hedera:testnet",
		Sign: func(message string) (LedgerAuthenticationSignerResult, error) {
			if message != "sign-me" {
				t.Fatalf("unexpected challenge message %s", message)
			}
			return LedgerAuthenticationSignerResult{
				Signature:     "sig",
				SignatureKind: "raw",
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("authenticate with ledger failed: %v", err)
	}

	headers := client.GetDefaultHeaders()
	if headers["x-api-key"] != "api-key-value" {
		t.Fatalf("expected API key to be updated, got %s", headers["x-api-key"])
	}
}

func intPointer(value int) *int {
	return &value
}
