package hcs2

type Operation string

const (
	OperationRegister Operation = "register"
	OperationUpdate   Operation = "update"
	OperationDelete   Operation = "delete"
	OperationMigrate  Operation = "migrate"
)

type RegistryType int

const (
	RegistryTypeIndexed    RegistryType = 0
	RegistryTypeNonIndexed RegistryType = 1
)

type Message struct {
	P        string    `json:"p"`
	Op       Operation `json:"op"`
	TopicID  string    `json:"t_id,omitempty"`
	UID      string    `json:"uid,omitempty"`
	Metadata string    `json:"metadata,omitempty"`
	Memo     string    `json:"m,omitempty"`
	TTL      int64     `json:"ttl,omitempty"`
}

type CreateRegistryOptions struct {
	RegistryType        RegistryType
	TTL                 int64
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
}

type RegisterEntryOptions struct {
	TargetTopicID string
	Metadata      string
	Memo          string
	AnalyticsMemo string
	RegistryType  *RegistryType
}

type UpdateEntryOptions struct {
	TargetTopicID string
	UID           string
	Metadata      string
	Memo          string
	AnalyticsMemo string
	RegistryType  *RegistryType
}

type DeleteEntryOptions struct {
	UID           string
	Memo          string
	AnalyticsMemo string
	RegistryType  *RegistryType
}

type MigrateRegistryOptions struct {
	TargetTopicID string
	Metadata      string
	Memo          string
	AnalyticsMemo string
	RegistryType  *RegistryType
}

type QueryRegistryOptions struct {
	Limit int
	Order string
	Skip  int64
}

type RegistryEntry struct {
	TopicID            string       `json:"topic_id"`
	Sequence           int64        `json:"sequence"`
	Timestamp          string       `json:"timestamp"`
	Payer              string       `json:"payer"`
	Message            Message      `json:"message"`
	ConsensusTimestamp string       `json:"consensus_timestamp"`
	RegistryType       RegistryType `json:"registry_type"`
}

type TopicRegistry struct {
	TopicID      string          `json:"topic_id"`
	RegistryType RegistryType    `json:"registry_type"`
	TTL          int64           `json:"ttl"`
	Entries      []RegistryEntry `json:"entries"`
	LatestEntry  *RegistryEntry  `json:"latest_entry,omitempty"`
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

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}
