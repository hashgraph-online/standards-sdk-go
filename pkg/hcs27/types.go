package hcs27

import (
	"context"
	"encoding/json"
)

const (
	ProtocolID    = "hcs-27"
	OperationName = "register"
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
	TreeSize    uint64 `json:"treeSize"`
	RootHashB64 string `json:"rootHashB64u"`
}

type PreviousCommitment struct {
	TreeSize    uint64 `json:"treeSize"`
	RootHashB64 string `json:"rootHashB64u"`
}

type BatchRange struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

type Signature struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Signature string `json:"b64u"`
}

type CheckpointMetadata struct {
	Type       string              `json:"type"`
	Stream     StreamID            `json:"stream"`
	Log        *LogProfile         `json:"log,omitempty"`
	Root       RootCommitment      `json:"root"`
	Previous   *PreviousCommitment `json:"prev,omitempty"`
	BatchRange BatchRange          `json:"batch_range"`
	Signature  *Signature          `json:"sig,omitempty"`
}

type MetadataDigest struct {
	Algorithm string `json:"alg"`
	DigestB64 string `json:"b64u"`
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
}

type PublishResult struct {
	TransactionID  string `json:"transaction_id"`
	SequenceNumber int64  `json:"sequence_number"`
}
