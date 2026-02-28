package hcs7

import hedera "github.com/hashgraph/hedera-sdk-go/v2"

type Operation string

const (
	OperationRegisterConfig Operation = "register-config"
	OperationRegister       Operation = "register"
)

type ConfigType string

const (
	ConfigTypeEVM  ConfigType = "evm"
	ConfigTypeWASM ConfigType = "wasm"
)

type StateValueType string

const (
	StateValueTypeNumber StateValueType = "number"
	StateValueTypeString StateValueType = "string"
	StateValueTypeBool   StateValueType = "bool"
)

type AbiIO struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

type AbiDefinition struct {
	Name            string  `json:"name"`
	Inputs          []AbiIO `json:"inputs"`
	Outputs         []AbiIO `json:"outputs"`
	StateMutability string  `json:"stateMutability"`
	Type            string  `json:"type"`
}

type EvmConfigPayload struct {
	ContractAddress string        `json:"contractAddress"`
	Abi             AbiDefinition `json:"abi"`
}

type WasmInputType struct {
	StateData map[string]StateValueType `json:"stateData"`
}

type WasmOutputType struct {
	Type   string `json:"type"`
	Format string `json:"format"`
}

type WasmConfigPayload struct {
	WasmTopicID string         `json:"wasmTopicId"`
	InputType   WasmInputType  `json:"inputType"`
	OutputType  WasmOutputType `json:"outputType"`
}

type MetadataData struct {
	Weight int64          `json:"weight"`
	Tags   []string       `json:"tags"`
	Extra  map[string]any `json:"-"`
}

type Message struct {
	P       string         `json:"p"`
	Op      Operation      `json:"op"`
	Type    ConfigType     `json:"t,omitempty"`
	Config  any            `json:"c,omitempty"`
	TopicID string         `json:"t_id,omitempty"`
	Data    map[string]any `json:"d,omitempty"`
	Memo    string         `json:"m,omitempty"`
}

type TopicMemo struct {
	Protocol string
	Indexed  bool
	TTL      int64
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

type RegisterConfigOptions struct {
	RegistryTopicID string
	Type            ConfigType
	EVM             *EvmConfigPayload
	WASM            *WasmConfigPayload
	Memo            string
	AnalyticsMemo   string
	SubmitKey       string
}

type RegisterMetadataOptions struct {
	RegistryTopicID string
	MetadataTopicID string
	Weight          int64
	Tags            []string
	Data            map[string]any
	Memo            string
	AnalyticsMemo   string
	SubmitKey       string
}

type RegistryOperationResult struct {
	Success        bool   `json:"success"`
	TransactionID  string `json:"transaction_id,omitempty"`
	Error          string `json:"error,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type RegistryEntry struct {
	SequenceNumber int64   `json:"sequence_number"`
	Timestamp      string  `json:"timestamp"`
	Payer          string  `json:"payer"`
	Message        Message `json:"message"`
}

type RegistryTopic struct {
	TopicID string          `json:"topic_id"`
	TTL     int64           `json:"ttl,omitempty"`
	Entries []RegistryEntry `json:"entries"`
}

type QueryRegistryOptions struct {
	Limit int
	Order string
	Skip  int64
}

type CreateRegistryResult struct {
	Success       bool   `json:"success"`
	TopicID       string `json:"topic_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type CreateRegistryTxParams struct {
	TTL       int64
	AdminKey  hedera.Key
	SubmitKey hedera.Key
}
