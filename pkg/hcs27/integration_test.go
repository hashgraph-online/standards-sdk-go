package hcs27

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS27Integration_CheckpointChain(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live Hedera integration tests")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	client, err := NewClient(ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create HCS-27 client: %v", err)
	}

	ctx := context.Background()

	topicID, transactionID, err := client.CreateCheckpointTopic(ctx, CreateTopicOptions{
		TTLSeconds:          600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create checkpoint topic: %v", err)
	}
	t.Logf("created checkpoint topic %s (tx=%s)", topicID, transactionID)

	rootOne := hashB64URL("go-sdk-hcs27-root-1")
	rootTwo := hashB64URL("go-sdk-hcs27-root-2")

	firstCheckpoint := CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: StreamID{
			Registry: "ans",
			LogID:    "default",
		},
		Log: &LogProfile{
			Algorithm: "sha-256",
			Leaf:      "sha256(jcs(event))",
			Merkle:    "rfc6962",
		},
		Root: RootCommitment{
			TreeSize:    1,
			RootHashB64: rootOne,
		},
		BatchRange: BatchRange{
			Start: 1,
			End:   1,
		},
	}

	firstResult, err := client.PublishCheckpoint(
		ctx,
		topicID,
		firstCheckpoint,
		"go-sdk checkpoint 1",
		"",
	)
	if err != nil {
		t.Fatalf("failed to publish first checkpoint: %v", err)
	}
	t.Logf("published first checkpoint seq=%d tx=%s", firstResult.SequenceNumber, firstResult.TransactionID)

	secondCheckpoint := CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: StreamID{
			Registry: "ans",
			LogID:    "default",
		},
		Log: &LogProfile{
			Algorithm: "sha-256",
			Leaf:      "sha256(jcs(event))",
			Merkle:    "rfc6962",
		},
		Root: RootCommitment{
			TreeSize:    2,
			RootHashB64: rootTwo,
		},
		Previous: &PreviousCommitment{
			TreeSize:    1,
			RootHashB64: rootOne,
		},
		BatchRange: BatchRange{
			Start: 2,
			End:   2,
		},
	}

	secondResult, err := client.PublishCheckpoint(
		ctx,
		topicID,
		secondCheckpoint,
		"go-sdk checkpoint 2",
		"",
	)
	if err != nil {
		t.Fatalf("failed to publish second checkpoint: %v", err)
	}
	t.Logf("published second checkpoint seq=%d tx=%s", secondResult.SequenceNumber, secondResult.TransactionID)

	records := waitForCheckpointRecords(t, ctx, client, topicID, 2, 30, 3*time.Second)

	if err := ValidateCheckpointChain(records); err != nil {
		t.Fatalf("checkpoint chain validation failed: %v", err)
	}
}

func TestHCS27Integration_MetadataOverflowUsesHRL(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live Hedera integration tests")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	client, err := NewClient(ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create HCS-27 client: %v", err)
	}

	ctx := context.Background()
	topicID, transactionID, err := client.CreateCheckpointTopic(ctx, CreateTopicOptions{
		TTLSeconds:          600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create checkpoint topic: %v", err)
	}
	t.Logf("created overflow test checkpoint topic %s (tx=%s)", topicID, transactionID)

	overflowMetadata := CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: StreamID{
			Registry: "ans",
			LogID:    "overflow",
		},
		Log: &LogProfile{
			Algorithm: "sha-256",
			Leaf:      strings.Repeat("sha256(jcs(event))-", 90),
			Merkle:    "rfc6962",
		},
		Root: RootCommitment{
			TreeSize:    1,
			RootHashB64: hashB64URL("go-sdk-hcs27-overflow-root-1"),
		},
		BatchRange: BatchRange{
			Start: 1,
			End:   1,
		},
	}

	metadataBytes, err := json.Marshal(overflowMetadata)
	if err != nil {
		t.Fatalf("failed to encode overflow metadata: %v", err)
	}
	inlinePayload, err := json.Marshal(CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
		Memo:      "go-sdk overflow checkpoint",
	})
	if err != nil {
		t.Fatalf("failed to encode inline checkpoint payload: %v", err)
	}
	if len(inlinePayload) <= 1024 {
		t.Fatalf("overflow test setup invalid: inline payload is %d bytes", len(inlinePayload))
	}

	result, err := client.PublishCheckpoint(
		ctx,
		topicID,
		overflowMetadata,
		"go-sdk overflow checkpoint",
		"",
	)
	if err != nil {
		t.Fatalf("failed to publish overflow checkpoint: %v", err)
	}
	t.Logf("published overflow checkpoint seq=%d tx=%s", result.SequenceNumber, result.TransactionID)

	messages := waitForTopicMessages(t, ctx, client, topicID, 1, 40, 3*time.Second)
	if len(messages) != 1 {
		t.Fatalf("expected exactly 1 checkpoint message, got %d", len(messages))
	}

	var checkpointMessage CheckpointMessage
	if err := mirror.DecodeMessageJSON(messages[0], &checkpointMessage); err != nil {
		t.Fatalf("failed to decode checkpoint message payload: %v", err)
	}

	var metadataReference string
	if err := json.Unmarshal(checkpointMessage.Metadata, &metadataReference); err != nil {
		t.Fatalf("expected metadata to be HCS-1 reference string, got: %v", err)
	}
	if !strings.HasPrefix(metadataReference, "hcs://1/") {
		t.Fatalf("unexpected metadata reference format: %s", metadataReference)
	}
	if strings.Contains(metadataReference, "@") {
		t.Fatalf("metadata reference must be HRL topic form (no sequence suffix): %s", metadataReference)
	}
	t.Logf("resolved overflow metadata reference %s", metadataReference)
	if checkpointMessage.MetadataDigest == nil {
		t.Fatalf("expected metadata_digest to be present for overflow message")
	}

	resolvedBytes := waitForResolvedMetadata(t, ctx, client, metadataReference, 40, 3*time.Second)
	var resolvedMetadata CheckpointMetadata
	if err := json.Unmarshal(resolvedBytes, &resolvedMetadata); err != nil {
		t.Fatalf("resolved metadata is not valid checkpoint JSON: %v", err)
	}
	if resolvedMetadata.Stream.LogID != "overflow" {
		t.Fatalf("resolved metadata stream.log_id mismatch: got %s", resolvedMetadata.Stream.LogID)
	}
}

func hashB64URL(input string) string {
	hash := sha256.Sum256([]byte(input))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func waitForCheckpointRecords(
	t *testing.T,
	ctx context.Context,
	client *Client,
	topicID string,
	minCount int,
	maxAttempts int,
	interval time.Duration,
) []CheckpointRecord {
	t.Helper()

	var (
		records []CheckpointRecord
		err     error
	)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		records, err = client.GetCheckpoints(ctx, topicID, nil)
		if err == nil && len(records) >= minCount {
			return records
		}
		time.Sleep(interval)
	}

	if err != nil {
		t.Fatalf("failed to load checkpoints after retries: %v", err)
	}
	t.Fatalf("expected at least %d checkpoint records after retries, got %d", minCount, len(records))
	return nil
}

func waitForTopicMessages(
	t *testing.T,
	ctx context.Context,
	client *Client,
	topicID string,
	minCount int,
	maxAttempts int,
	interval time.Duration,
) []mirror.TopicMessage {
	t.Helper()

	var (
		messages []mirror.TopicMessage
		err      error
	)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		messages, err = client.mirrorClient.GetTopicMessages(ctx, topicID, mirror.MessageQueryOptions{
			Order: "asc",
		})
		if err == nil && len(messages) >= minCount {
			return messages
		}
		time.Sleep(interval)
	}

	if err != nil {
		t.Fatalf("failed to load topic messages after retries: %v", err)
	}
	t.Fatalf("expected at least %d topic messages after retries, got %d", minCount, len(messages))
	return nil
}

func waitForResolvedMetadata(
	t *testing.T,
	ctx context.Context,
	client *Client,
	reference string,
	maxAttempts int,
	interval time.Duration,
) []byte {
	t.Helper()

	var (
		resolved []byte
		err      error
	)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resolved, err = client.ResolveHCS1Reference(ctx, reference)
		if err == nil && len(resolved) > 0 {
			return resolved
		}
		time.Sleep(interval)
	}

	if err != nil {
		t.Fatalf("failed to resolve metadata reference %s after retries: %v", reference, err)
	}
	t.Fatalf("resolved metadata reference %s returned empty payload after retries", reference)
	return nil
}
