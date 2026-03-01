package hcs27

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestPrepareCheckpointPayload_InlineMetadata(t *testing.T) {
	metadata := testCheckpointMetadata("")
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	message := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
		Memo:      "inline payload",
	}

	client := &Client{}
	preparedMessage, payload, resolvedMetadata, err := client.prepareCheckpointPayload(
		context.Background(),
		message,
		metadataBytes,
	)
	if err != nil {
		t.Fatalf("prepareCheckpointPayload failed: %v", err)
	}
	if len(payload) > 1024 {
		t.Fatalf("expected payload <= 1024 bytes, got %d", len(payload))
	}
	if preparedMessage.MetadataDigest != nil {
		t.Fatalf("expected metadata_digest to be nil for inline payload")
	}
	if len(resolvedMetadata) != 0 {
		t.Fatalf("expected no inline resolved metadata for non-overflow payload")
	}

	var decodedMetadata CheckpointMetadata
	if err := json.Unmarshal(preparedMessage.Metadata, &decodedMetadata); err != nil {
		t.Fatalf("expected inline metadata object, got: %v", err)
	}
}

func TestPrepareCheckpointPayload_MetadataOverflowUsesHCS1Reference(t *testing.T) {
	metadata := testCheckpointMetadata(strings.Repeat("leaf-profile-", 120))
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	message := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
		Memo:      "overflow payload",
	}

	expectedReference := "hcs://1/0.0.900000"
	sum := sha256.Sum256(metadataBytes)
	expectedDigest := base64.RawURLEncoding.EncodeToString(sum[:])

	client := &Client{
		publishMetadataOverride: func(ctx context.Context, payload []byte) (string, *MetadataDigest, error) {
			if string(payload) != string(metadataBytes) {
				t.Fatalf("published metadata payload mismatch")
			}
			return expectedReference, &MetadataDigest{
				Algorithm: "sha-256",
				DigestB64: expectedDigest,
			}, nil
		},
	}

	preparedMessage, payload, resolvedMetadata, err := client.prepareCheckpointPayload(
		context.Background(),
		message,
		metadataBytes,
	)
	if err != nil {
		t.Fatalf("prepareCheckpointPayload failed: %v", err)
	}
	if len(payload) > 1024 {
		t.Fatalf("expected overflow pointer payload <= 1024 bytes, got %d", len(payload))
	}
	if preparedMessage.MetadataDigest == nil {
		t.Fatalf("expected metadata_digest to be set for overflow payload")
	}
	if string(resolvedMetadata) != string(metadataBytes) {
		t.Fatalf("resolved metadata bytes mismatch for overflow payload")
	}

	var reference string
	if err := json.Unmarshal(preparedMessage.Metadata, &reference); err != nil {
		t.Fatalf("expected metadata reference string, got: %v", err)
	}
	if reference != expectedReference {
		t.Fatalf("metadata reference mismatch: got %s want %s", reference, expectedReference)
	}

	_, err = ValidateCheckpointMessage(context.Background(), preparedMessage, func(ctx context.Context, hcs1Reference string) ([]byte, error) {
		if hcs1Reference != expectedReference {
			t.Fatalf("unexpected HCS-1 reference: %s", hcs1Reference)
		}
		return metadataBytes, nil
	})
	if err != nil {
		t.Fatalf("expected prepared overflow message to validate, got: %v", err)
	}
}

func testCheckpointMetadata(leafProfile string) CheckpointMetadata {
	leaf := "sha256(jcs(event))"
	if strings.TrimSpace(leafProfile) != "" {
		leaf = leafProfile
	}

	return CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: StreamID{
			Registry: "ans",
			LogID:    "default",
		},
		Log: &LogProfile{
			Algorithm: "sha-256",
			Leaf:      leaf,
			Merkle:    "rfc6962",
		},
		Root: RootCommitment{
			TreeSize:    1,
			RootHashB64: base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
		},
		BatchRange: BatchRange{
			Start: 1,
			End:   1,
		},
	}
}

func TestExtractChunkTransactionID(t *testing.T) {
	testCases := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "string transaction id",
			input: "0.0.100@1772212929.454563067",
			want:  "0.0.100@1772212929.454563067",
		},
		{
			name: "map with transaction_valid_start",
			input: map[string]any{
				"account_id":              "0.0.100",
				"transaction_valid_start": "1772212929.454563067",
			},
			want: "0.0.100@1772212929.454563067",
		},
		{
			name: "map with valid_start_timestamp",
			input: map[string]string{
				"account_id":            "0.0.100",
				"valid_start_timestamp": "1772212929.454563067",
			},
			want: "0.0.100@1772212929.454563067",
		},
		{
			name:  "invalid input",
			input: map[string]any{"account_id": "0.0.100"},
			want:  "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := extractChunkTransactionID(testCase.input)
			if got != testCase.want {
				t.Fatalf("extractChunkTransactionID mismatch: got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestNormalizeHCS1PayloadWrappedBrotliDataURL(t *testing.T) {
	original := []byte(`{"type":"ans-checkpoint-v1","stream":{"registry":"ans","log_id":"overflow"}}`)

	var compressed bytes.Buffer
	writer := brotli.NewWriter(&compressed)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("failed to brotli-compress payload: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to finalize brotli payload: %v", err)
	}

	wrapped := map[string]any{
		"o": 0,
		"c": "data:application/json;base64," + base64.StdEncoding.EncodeToString(compressed.Bytes()),
	}
	wrappedBytes, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal wrapped payload: %v", err)
	}

	normalized, err := normalizeHCS1Payload(wrappedBytes)
	if err != nil {
		t.Fatalf("normalizeHCS1Payload returned error: %v", err)
	}
	if string(normalized) != string(original) {
		t.Fatalf("normalized payload mismatch: got %s", string(normalized))
	}
}

func TestNormalizeHCS1PayloadWrappedPlainDataURL(t *testing.T) {
	original := []byte(`{"type":"ans-checkpoint-v1"}`)
	wrapped := map[string]any{
		"o": 0,
		"c": "data:application/json;base64," + base64.StdEncoding.EncodeToString(original),
	}
	wrappedBytes, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal wrapped payload: %v", err)
	}

	normalized, err := normalizeHCS1Payload(wrappedBytes)
	if err != nil {
		t.Fatalf("normalizeHCS1Payload returned error: %v", err)
	}
	if string(normalized) != string(original) {
		t.Fatalf("normalized payload mismatch: got %s", string(normalized))
	}
}
