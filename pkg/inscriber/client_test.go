package inscriber

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStartInscriptionIncludesParityFields(t *testing.T) {
	var captured map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", request.Method)
		}
		if request.URL.Path != "/inscriptions/start-inscription" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"tx_id":"0.0.123@1772243084.050614451",
			"status":"pending",
			"topic_id":"0.0.7777",
			"transactionBytes":"AQIDBA==",
			"totalCost":12345,
			"totalMessages":3
		}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		APIKey:  "test-key",
		Network: NetworkTestnet,
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	job, err := client.StartInscription(context.Background(), StartInscriptionRequest{
		File: FileInput{
			Type:     "base64",
			Base64:   "dGVzdA==",
			FileName: "example.txt",
			MimeType: "text/plain",
		},
		HolderID:     "0.0.123",
		Mode:         ModeBulkFiles,
		Metadata:     map[string]any{"kind": "skill"},
		Tags:         []string{"alpha", "beta"},
		ChunkSize:    1024,
		FileStandard: "hcs-1",
	})
	if err != nil {
		t.Fatalf("StartInscription failed: %v", err)
	}

	if captured["chunkSize"] != float64(1024) {
		t.Fatalf("expected chunkSize 1024, got %#v", captured["chunkSize"])
	}
	if _, ok := captured["metadata"]; !ok {
		t.Fatalf("expected metadata to be sent")
	}
	if _, ok := captured["tags"]; !ok {
		t.Fatalf("expected tags to be sent")
	}
	if job.TotalCost != 12345 {
		t.Fatalf("expected totalCost parsing to preserve response field")
	}
	if job.TotalMessages != 3 {
		t.Fatalf("expected totalMessages parsing to preserve response field")
	}
}

func TestIsRetryableWaitError(t *testing.T) {
	if !isRetryableWaitError(errors.New("read: operation timed out")) {
		t.Fatalf("expected timeout errors to be retryable")
	}
	if isRetryableWaitError(context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded to be non-retryable")
	}
}
