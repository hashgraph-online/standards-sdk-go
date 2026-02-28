package hcs27

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildTopicMemo(t *testing.T) {
	memo := BuildTopicMemo(86400)
	if memo != "hcs-27:0:86400:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildTopicMemoDefault(t *testing.T) {
	memo := BuildTopicMemo(0)
	if memo != "hcs-27:0:86400:0" {
		t.Fatalf("expected default TTL, got: %s", memo)
	}
}

func TestBuildTopicMemoNegative(t *testing.T) {
	memo := BuildTopicMemo(-1)
	if memo != "hcs-27:0:86400:0" {
		t.Fatalf("expected default TTL for negative, got: %s", memo)
	}
}

func TestBuildTopicMemoCustom(t *testing.T) {
	memo := BuildTopicMemo(3600)
	if memo != "hcs-27:0:3600:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseTopicMemo(t *testing.T) {
	parsed, ok := ParseTopicMemo("hcs-27:0:86400:0")
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if parsed.IndexedFlag != 0 {
		t.Fatalf("unexpected indexed flag: %d", parsed.IndexedFlag)
	}
	if parsed.TTLSeconds != 86400 {
		t.Fatalf("unexpected TTL: %d", parsed.TTLSeconds)
	}
	if parsed.TopicType != 0 {
		t.Fatalf("unexpected topic type: %d", parsed.TopicType)
	}
}

func TestParseTopicMemoInvalid(t *testing.T) {
	cases := []string{
		"",
		"hcs-2:0:86400",
		"hcs-27:0:86400",
		"hcs-27:0:86400:0:extra",
		"bad:0:86400:0",
		"hcs-27:x:86400:0",
		"hcs-27:0:abc:0",
		"hcs-27:0:86400:x",
	}
	for _, c := range cases {
		_, ok := ParseTopicMemo(c)
		if ok {
			t.Fatalf("expected parse to fail for %q", c)
		}
	}
}

func TestBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo()
	if memo != "hcs-27:op:0:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestValidateCheckpointMessageProtocol(t *testing.T) {
	msg := CheckpointMessage{Protocol: "bad", Operation: OperationName}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error for invalid protocol")
	}
}

func TestValidateCheckpointMessageOperation(t *testing.T) {
	msg := CheckpointMessage{Protocol: ProtocolID, Operation: "bad"}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error for invalid operation")
	}
}

func TestValidateCheckpointMessageMemoTooLong(t *testing.T) {
	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Memo:      strings.Repeat("x", 300),
		Metadata:  json.RawMessage(`{}`),
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error for memo too long")
	}
}

func TestValidateCheckpointMessageEmptyMetadata(t *testing.T) {
	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error for empty metadata")
	}
}

func buildValidMetadata() CheckpointMetadata {
	rootHash := sha256.Sum256([]byte("test"))
	rootHashB64 := base64.RawURLEncoding.EncodeToString(rootHash[:])
	return CheckpointMetadata{
		Type: "ans-checkpoint-v1",
		Stream: StreamID{
			Registry: "0.0.12345",
			LogID:    "log-1",
		},
		Log: &LogProfile{
			Algorithm: "sha-256",
			Leaf:      "rfc6962",
			Merkle:    "rfc6962",
		},
		Root: RootCommitment{
			TreeSize:    10,
			RootHashB64: rootHashB64,
		},
		BatchRange: BatchRange{
			Start: 0,
			End:   10,
		},
	}
}

func TestValidateCheckpointMessageSuccess(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
	}
	result, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "ans-checkpoint-v1" {
		t.Fatalf("unexpected type: %s", result.Type)
	}
}

func TestValidateMetadataType(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Type = "bad"
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestValidateMetadataRegistry(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Stream.Registry = ""
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing registry")
	}
}

func TestValidateMetadataLogID(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Stream.LogID = ""
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing log_id")
	}
}

func TestValidateMetadataLog(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Log = nil
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing log")
	}
}

func TestValidateMetadataLogAlg(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Log.Algorithm = "bad"
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for bad log.alg")
	}
}

func TestValidateMetadataLogLeaf(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Log.Leaf = ""
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing log.leaf")
	}
}

func TestValidateMetadataLogMerkle(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Log.Merkle = ""
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing log.merkle")
	}
}

func TestValidateMetadataRootHash(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Root.RootHashB64 = "!invalid!"
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for invalid root hash")
	}
}

func TestValidateMetadataPrevious(t *testing.T) {
	metadata := buildValidMetadata()
	prevHash := sha256.Sum256([]byte("prev"))
	metadata.Previous = &PreviousCommitment{
		TreeSize:    5,
		RootHashB64: base64.RawURLEncoding.EncodeToString(prevHash[:]),
	}
	err := validateMetadata(metadata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMetadataPreviousInvalidHash(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Previous = &PreviousCommitment{
		TreeSize:    5,
		RootHashB64: "!invalid!",
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for invalid prev root hash")
	}
}

func TestValidateMetadataPreviousTreeSizeTooLarge(t *testing.T) {
	metadata := buildValidMetadata()
	prevHash := sha256.Sum256([]byte("prev"))
	metadata.Previous = &PreviousCommitment{
		TreeSize:    100,
		RootHashB64: base64.RawURLEncoding.EncodeToString(prevHash[:]),
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for prev.treeSize > root.treeSize")
	}
}

func TestValidateMetadataBatchRangeInvalid(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.BatchRange.Start = 5
	metadata.BatchRange.End = 3
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for batch_range.end < start")
	}
}

func TestValidateMetadataBatchRangeEndExceedsTreeSize(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.BatchRange.End = 100
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for batch_range.end > treeSize")
	}
}

func TestValidateMetadataSignature(t *testing.T) {
	metadata := buildValidMetadata()
	sigBytes := sha256.Sum256([]byte("sig"))
	metadata.Signature = &Signature{
		Algorithm: "ed25519",
		KeyID:     "did:hedera:testnet:0.0.12345",
		Signature: base64.RawURLEncoding.EncodeToString(sigBytes[:]),
	}
	err := validateMetadata(metadata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMetadataSignatureMissingAlg(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Signature = &Signature{
		Algorithm: "",
		KeyID:     "kid",
		Signature: "c2ln",
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing sig.alg")
	}
}

func TestValidateMetadataSignatureMissingKid(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Signature = &Signature{
		Algorithm: "ed25519",
		KeyID:     "",
		Signature: "c2ln",
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing sig.kid")
	}
}

func TestValidateMetadataSignatureMissingSig(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Signature = &Signature{
		Algorithm: "ed25519",
		KeyID:     "kid",
		Signature: "",
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for missing sig.b64u")
	}
}

func TestValidateMetadataSignatureInvalidB64(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Signature = &Signature{
		Algorithm: "ed25519",
		KeyID:     "kid",
		Signature: "!invalid!",
	}
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for invalid sig.b64u")
	}
}

func TestValidateMetadataTreeSizeZeroWithBatch(t *testing.T) {
	metadata := buildValidMetadata()
	metadata.Root.TreeSize = 0
	metadata.BatchRange.Start = 0
	metadata.BatchRange.End = 5
	err := validateMetadata(metadata)
	if err == nil {
		t.Fatal("expected error for treeSize=0 with batch_range.end>0")
	}
}

func TestValidateCheckpointMessageWithReference(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	reference := "hcs://1/0.0.99999"
	referenceBytes, _ := json.Marshal(reference)

	resolver := func(ctx context.Context, ref string) ([]byte, error) {
		return metadataBytes, nil
	}

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
	}
	result, err := ValidateCheckpointMessage(context.Background(), msg, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != "ans-checkpoint-v1" {
		t.Fatalf("unexpected type: %s", result.Type)
	}
}

func TestValidateCheckpointMessageReferenceNonHCS1(t *testing.T) {
	reference := "https://example.com"
	referenceBytes, _ := json.Marshal(reference)

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error for non-HCS-1 reference")
	}
}

func TestValidateCheckpointMessageReferenceNoResolver(t *testing.T) {
	reference := "hcs://1/0.0.99999"
	referenceBytes, _ := json.Marshal(reference)

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error when resolver is nil")
	}
}

func TestValidateCheckpointMessageWithDigest(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	reference := "hcs://1/0.0.99999"
	referenceBytes, _ := json.Marshal(reference)

	sum := sha256.Sum256(metadataBytes)
	digestB64 := base64.RawURLEncoding.EncodeToString(sum[:])

	resolver := func(ctx context.Context, ref string) ([]byte, error) {
		return metadataBytes, nil
	}

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
		MetadataDigest: &MetadataDigest{
			Algorithm: "sha-256",
			DigestB64: digestB64,
		},
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCheckpointMessageDigestMismatch(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	reference := "hcs://1/0.0.99999"
	referenceBytes, _ := json.Marshal(reference)

	resolver := func(ctx context.Context, ref string) ([]byte, error) {
		return metadataBytes, nil
	}

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
		MetadataDigest: &MetadataDigest{
			Algorithm: "sha-256",
			DigestB64: "wrongdigest",
		},
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, resolver)
	if err == nil {
		t.Fatal("expected error for digest mismatch")
	}
}

func TestValidateCheckpointMessageDigestBadAlg(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	reference := "hcs://1/0.0.99999"
	referenceBytes, _ := json.Marshal(reference)

	resolver := func(ctx context.Context, ref string) ([]byte, error) {
		return metadataBytes, nil
	}

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  referenceBytes,
		MetadataDigest: &MetadataDigest{
			Algorithm: "md5",
			DigestB64: "doesntmatter",
		},
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, resolver)
	if err == nil {
		t.Fatal("expected error for bad digest algorithm")
	}
}

func TestValidateCheckpointMessageDigestRequiresReference(t *testing.T) {
	metadata := buildValidMetadata()
	metadataBytes, _ := json.Marshal(metadata)

	msg := CheckpointMessage{
		Protocol:  ProtocolID,
		Operation: OperationName,
		Metadata:  metadataBytes,
		MetadataDigest: &MetadataDigest{
			Algorithm: "sha-256",
			DigestB64: "test",
		},
	}
	_, err := ValidateCheckpointMessage(context.Background(), msg, nil)
	if err == nil {
		t.Fatal("expected error when digest present without reference")
	}
}

func TestValidateCheckpointChain(t *testing.T) {
	rootHash1 := sha256.Sum256([]byte("root1"))
	rootHash2 := sha256.Sum256([]byte("root2"))

	records := []CheckpointRecord{
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream:   StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:     RootCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash2[:])},
				Previous: &PreviousCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
	}

	err := ValidateCheckpointChain(records)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCheckpointChainTreeSizeDecreased(t *testing.T) {
	rootHash1 := sha256.Sum256([]byte("root1"))
	rootHash2 := sha256.Sum256([]byte("root2"))

	records := []CheckpointRecord{
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream:   StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:     RootCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash2[:])},
				Previous: &PreviousCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
	}

	err := ValidateCheckpointChain(records)
	if err == nil {
		t.Fatal("expected error for tree size decrease")
	}
}

func TestValidateCheckpointChainMissingPrev(t *testing.T) {
	rootHash1 := sha256.Sum256([]byte("root1"))
	rootHash2 := sha256.Sum256([]byte("root2"))

	records := []CheckpointRecord{
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash2[:])},
			},
		},
	}

	err := ValidateCheckpointChain(records)
	if err == nil {
		t.Fatal("expected error for missing prev linkage")
	}
}

func TestValidateCheckpointChainPrevTreeSizeMismatch(t *testing.T) {
	rootHash1 := sha256.Sum256([]byte("root1"))
	rootHash2 := sha256.Sum256([]byte("root2"))

	records := []CheckpointRecord{
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream:   StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:     RootCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash2[:])},
				Previous: &PreviousCommitment{TreeSize: 3, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
	}

	err := ValidateCheckpointChain(records)
	if err == nil {
		t.Fatal("expected error for prev.treeSize mismatch")
	}
}

func TestValidateCheckpointChainPrevRootHashMismatch(t *testing.T) {
	rootHash1 := sha256.Sum256([]byte("root1"))
	rootHash2 := sha256.Sum256([]byte("root2"))
	wrongHash := sha256.Sum256([]byte("wrong"))

	records := []CheckpointRecord{
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream: StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:   RootCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash1[:])},
			},
		},
		{
			EffectiveMetadata: CheckpointMetadata{
				Stream:   StreamID{Registry: "0.0.1", LogID: "log-1"},
				Root:     RootCommitment{TreeSize: 10, RootHashB64: base64.RawURLEncoding.EncodeToString(rootHash2[:])},
				Previous: &PreviousCommitment{TreeSize: 5, RootHashB64: base64.RawURLEncoding.EncodeToString(wrongHash[:])},
			},
		},
	}

	err := ValidateCheckpointChain(records)
	if err == nil {
		t.Fatal("expected error for prev root hash mismatch")
	}
}

func TestExtractChunkTransactionIDString(t *testing.T) {
	result := extractChunkTransactionID("tx-id-123")
	if result != "tx-id-123" {
		t.Fatalf("expected 'tx-id-123', got %q", result)
	}
}

func TestExtractChunkTransactionIDMap(t *testing.T) {
	result := extractChunkTransactionID(map[string]any{
		"account_id":                "0.0.1",
		"transaction_valid_start":   "123.456",
	})
	if result != "0.0.1@123.456" {
		t.Fatalf("expected '0.0.1@123.456', got %q", result)
	}
}

func TestExtractChunkTransactionIDMapFallback(t *testing.T) {
	result := extractChunkTransactionID(map[string]any{
		"account_id":             "0.0.1",
		"valid_start_timestamp":  "789.012",
	})
	if result != "0.0.1@789.012" {
		t.Fatalf("expected '0.0.1@789.012', got %q", result)
	}
}

func TestExtractChunkTransactionIDMapString(t *testing.T) {
	result := extractChunkTransactionID(map[string]string{
		"account_id":               "0.0.2",
		"transaction_valid_start":  "111.222",
	})
	if result != "0.0.2@111.222" {
		t.Fatalf("expected '0.0.2@111.222', got %q", result)
	}
}

func TestExtractChunkTransactionIDMapStringFallback(t *testing.T) {
	result := extractChunkTransactionID(map[string]string{
		"account_id":            "0.0.2",
		"valid_start_timestamp": "333.444",
	})
	if result != "0.0.2@333.444" {
		t.Fatalf("expected '0.0.2@333.444', got %q", result)
	}
}

func TestExtractChunkTransactionIDNil(t *testing.T) {
	result := extractChunkTransactionID(nil)
	if result != "" {
		t.Fatalf("expected empty string for nil, got %q", result)
	}
}

func TestExtractChunkTransactionIDUnknownType(t *testing.T) {
	result := extractChunkTransactionID(42)
	if result != "" {
		t.Fatalf("expected empty string for int, got %q", result)
	}
}

func TestExtractChunkTransactionIDEmptyMap(t *testing.T) {
	result := extractChunkTransactionID(map[string]any{})
	if result != "" {
		t.Fatalf("expected empty for empty map, got %q", result)
	}
}

func TestExtractChunkTransactionIDMapMissingFields(t *testing.T) {
	result := extractChunkTransactionID(map[string]any{
		"account_id": "0.0.1",
	})
	if result != "" {
		t.Fatalf("expected empty when missing valid_start, got %q", result)
	}
}
