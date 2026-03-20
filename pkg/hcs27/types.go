package hcs27

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

const (
	ProtocolID                 = "hcs-27"
	OperationName              = "register"
	checkpointMetadataType     = "ans-checkpoint-v1"
	merkleProfileRFC9162       = "rfc9162"
	legacyMerkleProfileRFC6962 = "rfc6962"
)

type StreamID struct {
	Registry string `json:"registry"`
	LogID    string `json:"log_id"`
}

type LogProfile struct {
	Algorithm string `json:"alg"`
	Leaf      string `json:"leaf"`
	Merkle    string `json:"merkle"`
}

type RootCommitment struct {
	TreeSize     string `json:"treeSize"`
	RootHashB64u string `json:"rootHashB64u"`
}

func (commitment *RootCommitment) UnmarshalJSON(data []byte) error {
	type rawRootCommitment struct {
		TreeSize     json.RawMessage `json:"treeSize"`
		RootHashB64u string          `json:"rootHashB64u"`
	}

	var raw rawRootCommitment
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	treeSize, err := decodeLegacyTreeSize(raw.TreeSize)
	if err != nil {
		return fmt.Errorf("treeSize must be a JSON string or number: %w", err)
	}

	commitment.TreeSize = treeSize
	commitment.RootHashB64u = raw.RootHashB64u
	return nil
}

type PreviousCommitment struct {
	TreeSize     string `json:"treeSize"`
	RootHashB64u string `json:"rootHashB64u"`
}

func (commitment *PreviousCommitment) UnmarshalJSON(data []byte) error {
	type rawPreviousCommitment struct {
		TreeSize     json.RawMessage `json:"treeSize"`
		RootHashB64u string          `json:"rootHashB64u"`
	}

	var raw rawPreviousCommitment
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	treeSize, err := decodeLegacyTreeSize(raw.TreeSize)
	if err != nil {
		return fmt.Errorf("treeSize must be a JSON string or number: %w", err)
	}

	commitment.TreeSize = treeSize
	commitment.RootHashB64u = raw.RootHashB64u
	return nil
}

type Signature struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Signature string `json:"b64u"`
}

type CheckpointMetadata struct {
	Type      string              `json:"type"`
	Stream    StreamID            `json:"stream"`
	Log       *LogProfile         `json:"log,omitempty"`
	Root      RootCommitment      `json:"root"`
	Previous  *PreviousCommitment `json:"prev,omitempty"`
	Signature *Signature          `json:"sig,omitempty"`
}

type MetadataDigest struct {
	Algorithm string `json:"alg"`
	DigestB64 string `json:"b64u"`
}

type InclusionProof struct {
	LeafHash  string   `json:"leafHash"`
	LeafIndex string   `json:"leafIndex"`
	TreeSize  string   `json:"treeSize"`
	Path      []string `json:"path"`
	RootHash  string   `json:"rootHash"`
	// RootSignature is carried through for draft parity but is not verified here.
	RootSignature string `json:"rootSignature,omitempty"`
	TreeVersion   int    `json:"treeVersion"`
}

type ConsistencyProof struct {
	OldTreeSize     string   `json:"oldTreeSize"`
	NewTreeSize     string   `json:"newTreeSize"`
	OldRootHash     string   `json:"oldRootHash"`
	NewRootHash     string   `json:"newRootHash"`
	ConsistencyPath []string `json:"consistencyPath"`
	TreeVersion     int      `json:"treeVersion"`
}

type CheckpointMessage struct {
	Protocol       string          `json:"p"`
	Operation      string          `json:"op"`
	Metadata       json.RawMessage `json:"metadata"`
	MetadataDigest *MetadataDigest `json:"metadata_digest,omitempty"`
	Memo           string          `json:"m,omitempty"`
}

type TopicMemo struct {
	IndexedFlag int
	TTLSeconds  int64
	TopicType   int
}

type CheckpointRecord struct {
	TopicID            string             `json:"topic_id"`
	Sequence           int64              `json:"sequence"`
	ConsensusTimestamp string             `json:"consensus_timestamp"`
	Payer              string             `json:"payer"`
	Message            CheckpointMessage  `json:"message"`
	EffectiveMetadata  CheckpointMetadata `json:"effective_metadata"`
}

type PublishCheckpointOptions struct {
	TransactionMemo string
	HCS1Resolver    func(ctx context.Context, hcs1Reference string) ([]byte, error)
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
	InscriberAuthURL   string
	InscriberAPIURL    string
	HederaClient       *hedera.Client
}

type PublishResult struct {
	TransactionID  string `json:"transaction_id"`
	SequenceNumber int64  `json:"sequence_number"`
}

func decodeLegacyTreeSize(raw json.RawMessage) (string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "", fmt.Errorf("treeSize is required")
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString, nil
	}

	var asUint uint64
	if err := json.Unmarshal(raw, &asUint); err == nil {
		return canonicalUint64(asUint), nil
	}

	return "", fmt.Errorf("unsupported treeSize encoding")
}
