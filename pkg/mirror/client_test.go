package mirror

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientTestnet(t *testing.T) {
	client, err := NewClient(Config{Network: "testnet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://testnet.mirrornode.hedera.com" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
}

func TestNewClientMainnet(t *testing.T) {
	client, err := NewClient(Config{Network: "mainnet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://mainnet-public.mirrornode.hedera.com" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
}

func TestNewClientCustomBaseURL(t *testing.T) {
	client, err := NewClient(Config{
		Network: "testnet",
		BaseURL: "https://custom.example.com/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://custom.example.com" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
}

func TestNewClientUnsupportedNetwork(t *testing.T) {
	_, err := NewClient(Config{Network: "badnet"})
	if err == nil {
		t.Fatal("expected error for unsupported network")
	}
}

func TestNewClientWithAPIKey(t *testing.T) {
	client, err := NewClient(Config{
		Network: "testnet",
		APIKey:  "my-api-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.apiKey != "my-api-key" {
		t.Fatalf("expected api key 'my-api-key', got %q", client.apiKey)
	}
}

func TestNewClientWithHeaders(t *testing.T) {
	client, err := NewClient(Config{
		Network: "testnet",
		Headers: map[string]string{"X-Custom": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.headers["X-Custom"] != "test" {
		t.Fatalf("expected header X-Custom=test, got %q", client.headers["X-Custom"])
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	customHTTP := &http.Client{}
	client, err := NewClient(Config{
		Network:    "testnet",
		HTTPClient: customHTTP,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.httpClient != customHTTP {
		t.Fatal("expected custom http client to be used")
	}
}

func TestBaseURL(t *testing.T) {
	client, err := NewClient(Config{Network: "testnet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != "https://testnet.mirrornode.hedera.com" {
		t.Fatalf("unexpected BaseURL(): %s", client.BaseURL())
	}
}

func TestGetTopicInfoEmpty(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetTopicInfo(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty topic ID")
	}
}

func TestGetTopicInfoWhitespace(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetTopicInfo(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for whitespace topic ID")
	}
}

func TestGetTopicInfoSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/topics/0.0.12345" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TopicInfo{
			TopicID: "0.0.12345",
			Memo:    "test-memo",
			Deleted: false,
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
	})
	info, err := client.GetTopicInfo(context.Background(), "0.0.12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.TopicID != "0.0.12345" {
		t.Fatalf("unexpected topic ID: %s", info.TopicID)
	}
	if info.Memo != "test-memo" {
		t.Fatalf("unexpected memo: %s", info.Memo)
	}
}

func TestGetAccountEmpty(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetAccount(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty account ID")
	}
}

func TestGetAccountSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/accounts/0.0.12345" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AccountInfo{
			Account: "0.0.12345",
			Memo:    "account-memo",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
	})
	info, err := client.GetAccount(context.Background(), "0.0.12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Memo != "account-memo" {
		t.Fatalf("unexpected memo: %s", info.Memo)
	}
}

func TestGetAccountMemo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AccountInfo{
			Account: "0.0.12345",
			Memo:    "the-memo",
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
	})
	memo, err := client.GetAccountMemo(context.Background(), "0.0.12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if memo != "the-memo" {
		t.Fatalf("expected 'the-memo', got %q", memo)
	}
}

func TestGetTopicMessagesEmpty(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetTopicMessages(context.Background(), "", MessageQueryOptions{})
	if err == nil {
		t.Fatal("expected error for empty topic ID")
	}
}

func TestGetTopicMessagesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topicMessagesResponse{
			Messages: []TopicMessage{
				{
					SequenceNumber: 1,
					Message:        base64.StdEncoding.EncodeToString([]byte("hello")),
					TopicID:        "0.0.12345",
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
	})
	messages, err := client.GetTopicMessages(context.Background(), "0.0.12345", MessageQueryOptions{
		SequenceNumber: "gt:0",
		Limit:          10,
		Order:          "asc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
}

func TestGetTopicMessagesPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			resp := topicMessagesResponse{
				Messages: []TopicMessage{{SequenceNumber: 1, Message: base64.StdEncoding.EncodeToString([]byte("a"))}},
			}
			resp.Links.Next = "/api/v1/topics/0.0.1/messages?page=2"
			json.NewEncoder(w).Encode(resp)
		} else {
			json.NewEncoder(w).Encode(topicMessagesResponse{
				Messages: []TopicMessage{{SequenceNumber: 2, Message: base64.StdEncoding.EncodeToString([]byte("b"))}},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	messages, err := client.GetTopicMessages(context.Background(), "0.0.1", MessageQueryOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestGetTopicMessageBySequenceValidation(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetTopicMessageBySequence(context.Background(), "0.0.1", 0)
	if err == nil {
		t.Fatal("expected error for zero sequence")
	}
	_, err = client.GetTopicMessageBySequence(context.Background(), "0.0.1", -1)
	if err == nil {
		t.Fatal("expected error for negative sequence")
	}
}

func TestGetTopicMessageBySequenceSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topicMessagesResponse{
			Messages: []TopicMessage{{
				SequenceNumber: 5,
				Message:        base64.StdEncoding.EncodeToString([]byte("test")),
			}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	msg, err := client.GetTopicMessageBySequence(context.Background(), "0.0.1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.SequenceNumber != 5 {
		t.Fatalf("expected sequence 5, got %d", msg.SequenceNumber)
	}
}

func TestGetTopicMessageBySequenceNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topicMessagesResponse{Messages: []TopicMessage{}})
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	msg, err := client.GetTopicMessageBySequence(context.Background(), "0.0.1", 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Fatal("expected nil for not found")
	}
}

func TestDecodeMessageDataEmpty(t *testing.T) {
	_, err := DecodeMessageData(TopicMessage{Message: ""})
	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestDecodeMessageDataWhitespace(t *testing.T) {
	_, err := DecodeMessageData(TopicMessage{Message: "   "})
	if err == nil {
		t.Fatal("expected error for whitespace message")
	}
}

func TestDecodeMessageDataSuccess(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("hello world"))
	data, err := DecodeMessageData(TopicMessage{Message: encoded})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}
}

func TestDecodeMessageJSONSuccess(t *testing.T) {
	type testPayload struct {
		Name string `json:"name"`
	}
	payload := testPayload{Name: "test"}
	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.StdEncoding.EncodeToString(payloadBytes)

	var result testPayload
	err := DecodeMessageJSON(TopicMessage{Message: encoded}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" {
		t.Fatalf("expected 'test', got %q", result.Name)
	}
}

func TestDecodeMessageJSONInvalidBase64(t *testing.T) {
	var result map[string]string
	err := DecodeMessageJSON(TopicMessage{Message: "not-base64!!"}, &result)
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecodeMessageJSONInvalidJSON(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("not json"))
	var result map[string]string
	err := DecodeMessageJSON(TopicMessage{Message: encoded}, &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetTransactionEmpty(t *testing.T) {
	client, _ := NewClient(Config{Network: "testnet"})
	_, err := client.GetTransaction(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty transaction ID")
	}
}

func TestGetTransactionSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactionsResponse{
			Transactions: []Transaction{{
				TransactionID: "0.0.1@123.456",
				Result:        "SUCCESS",
			}},
		})
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	tx, err := client.GetTransaction(context.Background(), "0.0.1@123.456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
	if tx.Result != "SUCCESS" {
		t.Fatalf("expected 'SUCCESS', got %q", tx.Result)
	}
}

func TestGetTransactionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactionsResponse{Transactions: []Transaction{}})
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	tx, err := client.GetTransaction(context.Background(), "0.0.1@123.456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx != nil {
		t.Fatal("expected nil for not found")
	}
}

func TestGetJSONServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	_, err := client.GetTopicInfo(context.Background(), "0.0.1")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestGetJSONInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewClient(Config{Network: "testnet", BaseURL: server.URL})
	_, err := client.GetTopicInfo(context.Background(), "0.0.1")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestResolveURL(t *testing.T) {
	client := &Client{baseURL: "https://example.com"}

	if url := client.resolveURL("/api/test"); url != "https://example.com/api/test" {
		t.Fatalf("unexpected URL: %s", url)
	}

	if url := client.resolveURL("api/test"); url != "https://example.com/api/test" {
		t.Fatalf("unexpected URL: %s", url)
	}

	if url := client.resolveURL("https://other.com/path"); url != "https://other.com/path" {
		t.Fatalf("unexpected URL: %s", url)
	}

	if url := client.resolveURL("http://other.com/path"); url != "http://other.com/path" {
		t.Fatalf("unexpected URL: %s", url)
	}
}

func TestGetJSONWithAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer my-key" {
			t.Fatalf("expected 'Bearer my-key', got %q", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TopicInfo{TopicID: "0.0.1"})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
		APIKey:  "my-key",
	})
	_, err := client.GetTopicInfo(context.Background(), "0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetJSONWithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Fatalf("expected X-Custom=value, got %q", r.Header.Get("X-Custom"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TopicInfo{TopicID: "0.0.1"})
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		Network: "testnet",
		BaseURL: server.URL,
		Headers: map[string]string{"X-Custom": "value"},
	})
	_, err := client.GetTopicInfo(context.Background(), "0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
