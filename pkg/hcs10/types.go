package hcs10

import hedera "github.com/hashgraph/hedera-sdk-go/v2"

type TopicType int

const (
	TopicTypeInbound TopicType = iota
	TopicTypeOutbound
	TopicTypeConnection
	TopicTypeRegistry
)

type Operation string

const (
	OperationRegister          Operation = "register"
	OperationDelete            Operation = "delete"
	OperationConnectionRequest Operation = "connection_request"
	OperationConnectionCreated Operation = "connection_created"
	OperationMessage           Operation = "message"
	OperationCloseConnection   Operation = "close_connection"
	OperationTransaction       Operation = "transaction"
)

type Message struct {
	P                        string    `json:"p"`
	Op                       Operation `json:"op"`
	AccountID                string    `json:"account_id,omitempty"`
	InboundTopicID           string    `json:"inbound_topic_id,omitempty"`
	OutboundTopicID          string    `json:"outbound_topic_id,omitempty"`
	RequestorOutboundTopicID string    `json:"requestor_outbound_topic_id,omitempty"`
	ConnectionTopicID        string    `json:"connection_topic_id,omitempty"`
	ConnectedAccountID       string    `json:"connected_account_id,omitempty"`
	ConnectionID             int64     `json:"connection_id,omitempty"`
	ConnectionRequestID      int64     `json:"connection_request_id,omitempty"`
	ConfirmedRequestID       int64     `json:"confirmed_request_id,omitempty"`
	OperatorID               string    `json:"operator_id,omitempty"`
	Data                     string    `json:"data,omitempty"`
	UID                      string    `json:"uid,omitempty"`
	Memo                     string    `json:"m,omitempty"`
	ScheduleID               string    `json:"schedule_id,omitempty"`
	TransactionID            string    `json:"tx_id,omitempty"`
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
}

type CreateTopicOptions struct {
	TTL                 int64
	AccountID           string
	InboundTopicID      string
	ConnectionID        int64
	MetadataTopicID     string
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
	AdminKey            string
	SubmitKey           string
	MemoOverride        string
	TransactionMemo     string
}

type CreateRegistryTopicResult struct {
	Success       bool   `json:"success"`
	TopicID       string `json:"topic_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

type TopicRecord struct {
	TopicID string
	Memo    string
}

type MessageRecord struct {
	Message            Message `json:"message"`
	ConsensusTimestamp string  `json:"consensus_timestamp"`
	SequenceNumber     int64   `json:"sequence_number"`
	Payer              string  `json:"payer"`
}

type SubmitResult struct {
	Success        bool   `json:"success"`
	TransactionID  string `json:"transaction_id,omitempty"`
	Error          string `json:"error,omitempty"`
	SequenceNumber int64  `json:"sequence_number,omitempty"`
}

type CreateTopicTxParams struct {
	TopicType       TopicType
	TTL             int64
	AccountID       string
	InboundTopicID  string
	ConnectionID    int64
	MetadataTopicID string
	AdminKey        hedera.Key
	SubmitKey       hedera.Key
	MemoOverride    string
}
