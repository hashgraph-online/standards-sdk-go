package hcs12

import hedera "github.com/hashgraph/hedera-sdk-go/v2"

type RegistryType string

const (
	RegistryTypeAction    RegistryType = "action"
	RegistryTypeAssembly  RegistryType = "assembly"
	RegistryTypeHashlinks RegistryType = "hashlinks"
)

type AssemblyOperation string

const (
	OperationRegister  AssemblyOperation = "register"
	OperationAddAction AssemblyOperation = "add-action"
	OperationAddBlock  AssemblyOperation = "add-block"
	OperationUpdate    AssemblyOperation = "update"
)

type ActionRegistration struct {
	P           string   `json:"p"`
	Op          string   `json:"op"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	TID         string   `json:"t_id,omitempty"`
}

type AssemblyRegistration struct {
	P           string   `json:"p"`
	Op          string   `json:"op"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type AssemblyAddAction struct {
	P     string `json:"p"`
	Op    string `json:"op"`
	TID   string `json:"t_id"`
	Alias string `json:"alias,omitempty"`
}

type AssemblyAddBlock struct {
	P       string         `json:"p"`
	Op      string         `json:"op"`
	BlockID string         `json:"block_t_id"`
	Data    map[string]any `json:"data,omitempty"`
}

type AssemblyUpdate struct {
	P    string         `json:"p"`
	Op   string         `json:"op"`
	Data map[string]any `json:"data,omitempty"`
}

type HashLinksRegistration struct {
	P           string   `json:"p"`
	Op          string   `json:"op"`
	TID         string   `json:"t_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category,omitempty"`
	Featured    bool     `json:"featured,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Author      string   `json:"author,omitempty"`
	Website     string   `json:"website,omitempty"`
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateRegistryTopicOptions struct {
	RegistryType        RegistryType
	TTL                 int64
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
	MemoOverride        string
	TransactionMemo     string
}

type SubmitMessageResult struct {
	Success        bool   `json:"success"`
	TransactionID  string `json:"transaction_id,omitempty"`
	Error          string `json:"error,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type CreateTopicResult struct {
	Success       bool   `json:"success"`
	TopicID       string `json:"topic_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type RegistryEntry struct {
	SequenceNumber     int64          `json:"sequence_number"`
	ConsensusTimestamp string         `json:"consensus_timestamp"`
	Payer              string         `json:"payer"`
	Payload            map[string]any `json:"payload"`
}

type QueryOptions struct {
	SequenceNumber string
	Limit          int
	Order          string
}

type CreateRegistryTopicTxParams struct {
	RegistryType RegistryType
	TTL          int64
	AdminKey     hedera.Key
	SubmitKey    hedera.Key
	MemoOverride string
}
