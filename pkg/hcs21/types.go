package hcs21

import "regexp"

const (
	Protocol             = "hcs-21"
	MaxMessageBytes      = 1024
	SafeMessageBytes     = 1000
	DefaultTopicTTL int64 = 86400
)

type Operation string

const (
	OperationRegister Operation = "register"
	OperationUpdate   Operation = "update"
	OperationDelete   Operation = "delete"
)

type TopicType int

const (
	TopicTypeAdapterRegistry      TopicType = 0
	TopicTypeRegistryOfRegistries TopicType = 1
	TopicTypeAdapterCategory      TopicType = 2
)

type AdapterPackage struct {
	Registry  string `json:"registry"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
}

type AdapterDeclaration struct {
	P                string         `json:"p"`
	Op               Operation      `json:"op"`
	AdapterID        string         `json:"adapter_id"`
	Entity           string         `json:"entity"`
	Package          AdapterPackage `json:"package"`
	Manifest         string         `json:"manifest"`
	ManifestSequence int64          `json:"manifest_sequence,omitempty"`
	Config           map[string]any `json:"config"`
	StateModel       string         `json:"state_model,omitempty"`
	Signature        string         `json:"signature,omitempty"`
}

type AdapterDeclarationEnvelope struct {
	Declaration        AdapterDeclaration `json:"declaration"`
	ConsensusTimestamp string             `json:"consensus_timestamp,omitempty"`
	SequenceNumber     int64              `json:"sequence_number"`
	Payer              string             `json:"payer,omitempty"`
}

type AdapterCategoryEntry struct {
	AdapterID          string `json:"adapter_id"`
	AdapterTopicID     string `json:"adapter_topic_id"`
	Metadata           string `json:"metadata,omitempty"`
	Memo               string `json:"memo,omitempty"`
	Payer              string `json:"payer,omitempty"`
	SequenceNumber     int64  `json:"sequence_number"`
	ConsensusTimestamp string `json:"consensus_timestamp,omitempty"`
}

type ManifestPointer struct {
	Pointer          string `json:"pointer"`
	TopicID          string `json:"topic_id"`
	SequenceNumber   int64  `json:"sequence_number"`
	ManifestSequence int64  `json:"manifest_sequence,omitempty"`
	JobID            string `json:"job_id,omitempty"`
	TransactionID    string `json:"transaction_id,omitempty"`
	TotalCostHbar    string `json:"total_cost_hbar,omitempty"`
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateRegistryTopicOptions struct {
	TTL                 int64
	Indexed             bool
	Type                TopicType
	MetaTopicID         string
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
	TransactionMemo     string
}

type PublishDeclarationOptions struct {
	TopicID         string
	Declaration     AdapterDeclaration
	TransactionMemo string
}

type FetchDeclarationsOptions struct {
	Limit int
	Order string
}

type BuildDeclarationParams struct {
	Op               Operation
	AdapterID        string
	Entity           string
	Package          AdapterPackage
	Manifest         string
	ManifestSequence int64
	Config           map[string]any
	StateModel       string
	Signature        string
}

var (
	manifestPointerPattern = regexp.MustCompile(`^(?:hcs:\/\/1\/0\.0\.\d+|ipfs:\/\/\S+|ar:\/\/\S+|oci:\/\/\S+|https?:\/\/\S+)$`)
	metaPointerPattern     = regexp.MustCompile(`^(?:0\.0\.\d+|hcs:\/\/1\/0\.0\.\d+(?:\/\d+)?|ipfs:\/\/\S+|ar:\/\/\S+|oci:\/\/\S+|https?:\/\/\S+)$`)
)

