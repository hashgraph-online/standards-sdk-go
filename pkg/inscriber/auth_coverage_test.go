package inscriber

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewAuthClient(t *testing.T) {
	client := NewAuthClient("")
	if client.baseURL != "https://kiloscribe.com" {
		t.Fatal("expected default URL")
	}

	client2 := NewAuthClient("https://custom.com/api/")
	if client2.baseURL != "https://custom.com" {
		t.Fatal("expected trailing space and /api removed")
	}
}

func TestNormalizeChallengeMessage(t *testing.T) {
	str, obj, err := normalizeChallengeMessage(json.RawMessage(`"hello"`))
	if err != nil || str != "hello" || obj != "hello" {
		t.Fatal("expected hello string")
	}

	_, _, err = normalizeChallengeMessage(json.RawMessage(`""`))
	if err == nil {
		t.Fatal("expected empty string err")
	}

	str, obj, err = normalizeChallengeMessage(json.RawMessage(`{"a": 1}`))
	if err != nil || obj == nil {
		t.Fatal("expected obj")
	}

	_, _, err = normalizeChallengeMessage(json.RawMessage(``))
	if err == nil {
		t.Fatal("expected empty err")
	}
}

func TestAuthenticateSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/request-signature", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "test-challenge"}`))
	})
	mux.HandleFunc("/api/auth/authenticate", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"apiKey": "secret-key", "user": {"sessionToken": "token"}}`))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	pk, _ := hedera.PrivateKeyGenerateEcdsa()

	client := NewAuthClient(ts.URL)
	res, err := client.Authenticate(context.Background(), "0.0.1", pk.String(), NetworkTestnet)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.APIKey != "secret-key" {
		t.Fatal("expected secret-key")
	}
}

func TestAuthenticateFailReqSig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client := NewAuthClient(ts.URL)
	_, err := client.Authenticate(context.Background(), "0.0.1", pk.String(), NetworkTestnet)
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestAuthenticateFailOnAuth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/request-signature", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message": "test-challenge"}`))
	})
	mux.HandleFunc("/api/auth/authenticate", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	pk, _ := hedera.PrivateKeyGenerateEcdsa()

	client := NewAuthClient(ts.URL)
	_, err := client.Authenticate(context.Background(), "0.0.1", pk.String(), NetworkTestnet)
	
	if err == nil {
		t.Fatal("expected auth post to fail")
	}
}
