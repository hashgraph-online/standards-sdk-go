package inscriber

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseString(t *testing.T) {
	if parseString("test") != "test" {
		t.Fatal("expected test")
	}
	if parseString(float64(5.5)) != "5.5" {
		t.Fatal("expected 5.5")
	}
	if parseString(int64(42)) != "42" {
		t.Fatal("expected 42")
	}
	if parseString(int(42)) != "42" {
		t.Fatal("expected 42")
	}
	if parseString(nil) != "" {
		t.Fatal("expected empty string")
	}
}

func TestParseFloat(t *testing.T) {
	if parseFloat(float64(5.5)) != 5.5 {
		t.Fatal("expected 5.5")
	}
	if parseFloat(float32(5.5)) != float64(float32(5.5)) {
		t.Fatal("expected 5.5")
	}
	if parseFloat(int(42)) != 42.0 {
		t.Fatal("expected 42")
	}
	if parseFloat(int64(42)) != 42.0 {
		t.Fatal("expected 42")
	}
	if parseFloat("5.5") != 5.5 {
		t.Fatal("expected 5.5")
	}
	if parseFloat("abc") != 0 {
		t.Fatal("expected 0")
	}
	if parseFloat(nil) != 0 {
		t.Fatal("expected 0")
	}
}

func TestFirstNonEmptyString(t *testing.T) {
	m := map[string]any{"key1": "", "key2": "val2"}
	if firstNonEmptyString(m, "key1", "key2") != "val2" {
		t.Fatal("expected val2")
	}
	if firstNonEmptyString(m, "key3") != "" {
		t.Fatal("expected empty")
	}
}

func TestNormalizeWebSocketURL(t *testing.T) {
	if normalizeWebSocketURL("") != "" {
		t.Fatal("expected empty")
	}
	if normalizeWebSocketURL("  http://example.com  ") != "http://example.com" {
		t.Fatal("expected http://example.com")
	}
	if normalizeWebSocketURL("invalid url %") != "invalid url %" {
		t.Fatal("expected unchanged for invalid URL")
	}
}

func TestResolveWebSocketBaseURL(t *testing.T) {
	client := &Client{webSocketBaseURL: "ws://custom"}
	url, err := client.resolveWebSocketBaseURL(context.Background())
	if err != nil || url != "ws://custom" {
		t.Fatal("expected custom url")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"recommended": "ws://rec", "servers": [{"url": "ws://1", "status": "active"}]}`))
	}))
	defer ts.Close()

	client2 := &Client{baseURL: ts.URL, httpClient: &http.Client{}}
	url2, err := client2.resolveWebSocketBaseURL(context.Background())
	if err != nil || url2 != "ws://rec" {
		t.Fatal("expected rec url")
	}

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"servers": [{"url": "ws://1", "status": "active"}]}`))
	}))
	defer ts2.Close()

	client3 := &Client{baseURL: ts2.URL, httpClient: &http.Client{}}
	url3, err := client3.resolveWebSocketBaseURL(context.Background())
	if err != nil || url3 != "ws://1" {
		t.Fatal("expected active url")
	}

	ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"servers": [{"url": "ws://2", "status": "inactive"}]}`))
	}))
	defer ts3.Close()

	client4 := &Client{baseURL: ts3.URL, httpClient: &http.Client{}}
	url4, err := client4.resolveWebSocketBaseURL(context.Background())
	if err != nil || url4 != "ws://2" {
		t.Fatal("expected inactive url")
	}

	ts4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"servers": []}`))
	}))
	defer ts4.Close()

	client5 := &Client{baseURL: ts4.URL, httpClient: &http.Client{}}
	_, err = client5.resolveWebSocketBaseURL(context.Background())
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestMatchesInscriptionEvent(t *testing.T) {
	if !matchesInscriptionEvent("", map[string]any{}) {
		t.Fatal("expected true")
	}
	
	m := map[string]any{"tx_id": "0.0.1-1-1"}
	if !matchesInscriptionEvent("0.0.1-1-1", m) {
		t.Fatal("expected true")
	}

	m2 := map[string]any{"tx_id": "0.0.1-2-2"}
	if matchesInscriptionEvent("0.0.1-1-1", m2) {
		t.Fatal("expected false")
	}
}

func TestParseInscriptionEvent(t *testing.T) {
	m := map[string]any{
		"id": "1",
		"status": "completed",
		"tx_id": "tx1",
		"transactionId": "tx1",
		"topicId": "t1",
		"error": "e1",
	}
	job := parseInscriptionEvent(m)
	if job.ID != "1" || !job.Completed || job.TxID != "tx1" || job.TransactionID != "tx1" || job.TopicID != "t1" || job.Error != "e1" {
		t.Fatal("unexpected job fields")
	}
}

func TestWaitForInscriptionWebSocketFailures(t *testing.T) {
	client := &Client{webSocketBaseURL: "\x00invalid"}
	// This will make socketio.NewClient fail or return error
	_, err := client.waitForInscriptionWebSocket(context.Background(), "0.0.1", nil)
	if err == nil {
		t.Fatal("expected failure to connect to websocket")
	}

	client2 := &Client{baseURL: "http://[::1]:namedport", httpClient: &http.Client{}}
	_, err = client2.waitForInscriptionWebSocket(context.Background(), "0.0.1", nil)
	// This will fail because resolveWebSocketBaseURL fails
	if err == nil {
		t.Fatal("expected err during resolve")
	}
}
