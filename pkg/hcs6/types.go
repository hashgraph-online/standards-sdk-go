package hcs6

type Operation string

const (
	OperationRegister Operation = "register"
)

type RegistryType int

const (
	RegistryTypeNonIndexed RegistryType = 1
)

type Message struct {
	P       string    `json:"p"`
	Op      Operation `json:"op"`
	TopicID string    `json:"t_id,omitempty"`
	Memo    string    `json:"m,omitempty"`
}

type TopicMemo struct {
	Protocol     string
	RegistryType RegistryType
	TTL          int64
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateRegistryOptions struct {
	TTL                 int64
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
}

type RegisterEntryOptions struct {
	TargetTopicID string
	Memo          string
	AnalyticsMemo string
}

type QueryRegistryOptions struct {
	Limit int
	Order string
	Skip  int64
}

type RegistryEntry struct {
	TopicID            string      `json:"topic_id"`
	Sequence           int64       `json:"sequence"`
	Timestamp          string      `json:"timestamp"`
	Payer              string      `json:"payer"`
	Message            Message     `json:"message"`
	ConsensusTimestamp string      `json:"consensus_timestamp"`
	RegistryType       RegistryType `json:"registry_type"`
}

type TopicRegistry struct {
	TopicID      string        `json:"topic_id"`
	RegistryType RegistryType  `json:"registry_type"`
	TTL          int64         `json:"ttl"`
	Entries      []RegistryEntry `json:"entries"`
	LatestEntry  *RegistryEntry `json:"latest_entry,omitempty"`
}

type CreateRegistryResult struct {
	Success       bool   `json:"success"`
	TopicID       string `json:"topic_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type OperationResult struct {
	Success        bool   `json:"success"`
	TransactionID  string `json:"transaction_id,omitempty"`
	Error          string `json:"error,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

