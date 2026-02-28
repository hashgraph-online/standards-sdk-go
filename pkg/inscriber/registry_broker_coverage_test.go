package inscriber

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewBrokerClient(t *testing.T) {
	client, err := NewBrokerClient("", "valid-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != "https://registry.hashgraphonline.com/api/v1" {
		t.Fatal("expected default URL")
	}

	_, err = NewBrokerClient("https://custom.com/", "   ")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}

	client2, _ := NewBrokerClient("https://custom.com/", "key")
	if client2.baseURL != "https://custom.com" {
		t.Fatal("expected trimmed Custom URL")
	}
}

func TestBrokerClientCreateQuote(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/inscribe/content/quote" {
			w.Write([]byte(`{"totalCostHbar": 5.5}`))
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	resp, err := client.CreateQuote(context.Background(), BrokerQuoteRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCostHBAR != 5.5 {
		t.Fatal("unexpected total cost")
	}

	clientBad, _ := NewBrokerClient(ts.URL, "bad")
	_, err = clientBad.CreateQuote(context.Background(), BrokerQuoteRequest{})
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestBrokerClientCreateJob(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscribe/content" {
			w.Write([]byte(`{"jobId": "job-123"}`))
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	resp, err := client.CreateJob(context.Background(), BrokerQuoteRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.JobID != "job-123" {
		t.Fatal("unexpected job ID")
	}
}

func TestBrokerClientGetJob(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscribe/content/job-123" {
			w.Write([]byte(`{"status": "completed"}`))
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	resp, err := client.GetJob(context.Background(), "job-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "completed" {
		t.Fatal("unexpected status")
	}
}

func TestBrokerClientWaitForJobSuccess(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscribe/content/job-waitFor" {
			attempts++
			if attempts < 2 {
				w.Write([]byte(`{"status": "processing"}`))
			} else {
				w.Write([]byte(`{"status": "completed", "hrl": "hrl:123"}`))
			}
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	client.pollInterval = 10 * time.Millisecond

	resp, err := client.WaitForJob(context.Background(), "job-waitFor", 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "completed" || resp.HRL != "hrl:123" {
		t.Fatal("unexpected completed response")
	}
}

func TestBrokerClientWaitForJobFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "failed", "error": "custom error"}`))
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	client.pollInterval = 10 * time.Millisecond

	_, err := client.WaitForJob(context.Background(), "job-failed", 1*time.Second)
	if err == nil || err.Error() != "custom error" {
		t.Fatalf("expected custom error, got %v", err)
	}
}

func TestBrokerClientWaitForJobTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "processing"}`))
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	client.pollInterval = 10 * time.Millisecond

	_, err := client.WaitForJob(context.Background(), "job-processing", 30*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestBrokerClientInscribeAndWait(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscribe/content" {
			w.Write([]byte(`{"id": "job-1"}`))
		} else if r.URL.Path == "/inscribe/content/job-1" {
			w.Write([]byte(`{"id": "job-1", "status": "completed"}`))
		}
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	client.pollInterval = 10 * time.Millisecond

	res, err := client.InscribeAndWait(context.Background(), BrokerQuoteRequest{}, 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.JobID != "job-1" || !res.Confirmed {
		t.Fatal("unexpected InscribeAndWait result")
	}
}

func TestBrokerClientInscribeAndWaitMissingID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "processing"}`))
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	_, err := client.InscribeAndWait(context.Background(), BrokerQuoteRequest{}, 1*time.Second)
	if err == nil {
		t.Fatal("expected error missing ID")
	}
}

func TestPostJSONError(t *testing.T) {
	client := &BrokerClient{baseURL: "http://[::1]:namedport"}
	err := client.postJSON(context.Background(), "/", nil, nil)
	if err == nil {
		t.Fatal("expected post error")
	}

	err = client.getJSON(context.Background(), "/", nil)
	if err == nil {
		t.Fatal("expected get error")
	}
}

func TestPostJSONDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{bad-json`))
	}))
	defer ts.Close()

	client, _ := NewBrokerClient(ts.URL, "secret")
	err := client.postJSON(context.Background(), "/", nil, &BrokerJobResponse{})
	if err == nil {
		t.Fatal("expected decode error")
	}
	err = client.getJSON(context.Background(), "/", &BrokerJobResponse{})
	if err == nil {
		t.Fatal("expected get decode error")
	}
}

func TestBuildBrokerQuoteRequest(t *testing.T) {
	req, err := buildBrokerQuoteRequest(InscriptionInput{
		Type: InscriptionInputTypeBuffer,
		Buffer: []byte("data"),
		FileName: "test.txt",
		MimeType: "text/plain",
	}, InscribeViaRegistryBrokerOptions{
		Mode: ModeFile,
	})
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Base64 != encodeBufferToBase64([]byte("data")) {
		t.Fatal("unexpected content b64")
	}
	if req.Mode != ModeFile {
		t.Fatal("unexpected mode")
	}
}

func TestInscribeViaRegistryBroker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscribe/content" {
			w.Write([]byte(`{"id": "job-1"}`))
		} else if r.URL.Path == "/inscribe/content/job-1" {
			w.Write([]byte(`{"id": "job-1", "status": "completed"}`))
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	options := InscribeViaRegistryBrokerOptions{
		APIKey: "secret",
		BaseURL: ts.URL,
	}

	_, err := InscribeViaRegistryBroker(ctx, InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = GetRegistryBrokerQuote(ctx, InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, options)
	// URL returns 404 for /inscribe/content/quote in our mock, so expect error
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestInscribeSkillViaRegistryBroker(t *testing.T) {
	ctx := context.Background()
	_, err := InscribeSkillViaRegistryBroker(ctx, InscriptionInput{Type: InscriptionInputTypeBuffer, Buffer: []byte("test"), FileName: "t.txt"}, SkillInscriptionOptions{})
	if err == nil {
		t.Fatal("expected missing api key error")
	}
}
