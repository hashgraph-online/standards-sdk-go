package hcs16

import hedera "github.com/hashgraph/hedera-sdk-go/v2"

type FloraTopicType int

const (
	FloraTopicTypeCommunication FloraTopicType = 0
	FloraTopicTypeTransaction   FloraTopicType = 1
	FloraTopicTypeState         FloraTopicType = 2
)

type FloraOperation string

const (
	FloraOperationFloraCreated FloraOperation = "flora_created"
	FloraOperationTransaction  FloraOperation = "transaction"
	FloraOperationStateUpdate  FloraOperation = "state_update"
	FloraOperationJoinRequest  FloraOperation = "flora_join_request"
	FloraOperationJoinVote     FloraOperation = "flora_join_vote"
	FloraOperationJoinAccepted FloraOperation = "flora_join_accepted"
)

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	KeyType            string
	MirrorBaseURL      string
	MirrorAPIKey       string
	InscriberAuthURL   string
	InscriberAPIURL    string
}

type FloraMember struct {
	AccountID string `json:"accountId"`
	PublicKey string `json:"publicKey,omitempty"`
	Weight    int    `json:"weight,omitempty"`
}

type FloraTopics struct {
	Communication string `json:"communication"`
	Transaction   string `json:"transaction"`
	State         string `json:"state"`
}

type FloraMessage map[string]any

type FloraMessageRecord struct {
	Message            FloraMessage
	ConsensusTimestamp string
	SequenceNumber     int64
	Payer              string
}

type CreateFloraTopicOptions struct {
	FloraAccountID   string
	TopicType        FloraTopicType
	AdminKey         hedera.Key
	SubmitKey        hedera.Key
	AutoRenewAccount string
	SignerKeys       []hedera.PrivateKey
	TransactionMemo  string
}

type CreateFloraAccountOptions struct {
	KeyList                       *hedera.KeyList
	InitialBalanceHbar            float64
	MaxAutomaticTokenAssociations int32
}

type CreateFloraAccountWithTopicsOptions struct {
	Members            []string
	Threshold          int
	InitialBalanceHbar float64
	AutoRenewAccountID string
}

type TransactionTopicFee struct {
	Amount                int64
	FeeCollectorAccountID string
	DenominatingTokenID   string
}

type TransactionTopicConfig struct {
	Memo           string
	AdminKey       hedera.Key
	SubmitKey      hedera.Key
	FeeScheduleKey hedera.Key
	CustomFees     []TransactionTopicFee
	FeeExemptKeys  []hedera.Key
}

type CreateFloraProfileOptions struct {
	FloraAccountID  string
	DisplayName     string
	Members         []FloraMember
	Threshold       int
	Topics          FloraTopics
	InboundTopicID  string
	OutboundTopicID string
	Bio             string
	Metadata        map[string]any
	Policies        map[string]any
}

type CreateFloraProfileResult struct {
	ProfileTopicID string
	TransactionID  string
}

type ParseTopicMemoResult struct {
	Protocol       string
	FloraAccountID string
	TopicType      FloraTopicType
}

type CreateFloraAccountWithTopicsResult struct {
	FloraAccountID string
	Topics         FloraTopics
}

const (
	HCS16FloraAccountCreateTransactionMemo = "hcs-16:op:0:0"
	HCS16AccountKeyUpdateTransactionMemo   = "hcs-16:op:1:1"
	HCS16TopicKeyUpdateTransactionMemo     = "hcs-16:op:1:1"
	HCS17StateHashTransactionMemo          = "hcs-17:op:6:2"
)
