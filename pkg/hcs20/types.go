package hcs20

import (
	"regexp"
	"time"
)

const (
	ProtocolID = "hcs-20"

	OperationDeploy   = "deploy"
	OperationMint     = "mint"
	OperationBurn     = "burn"
	OperationTransfer = "transfer"
	OperationRegister = "register"

	DefaultPublicTopicID   = "0.0.4350190"
	DefaultRegistryTopicID = "0.0.4362300"

	MaxNumberLength   = 18
	MaxNameLength     = 100
	MaxMetadataLength = 100
	MaxMemoLength     = 500
)

var (
	hederaEntityRegex = regexp.MustCompile(`^(0|(?:[1-9]\d*))\.(0|(?:[1-9]\d*))\.(0|(?:[1-9]\d*))(?:-([a-z]{5}))?$`)
	numberRegex       = regexp.MustCompile(`^\d+$`)
)

type Message struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`

	Name     string `json:"name,omitempty"`
	Tick     string `json:"tick,omitempty"`
	Max      string `json:"max,omitempty"`
	Limit    string `json:"lim,omitempty"`
	Metadata string `json:"metadata,omitempty"`
	Memo     string `json:"m,omitempty"`

	Amount string `json:"amt,omitempty"`
	To     string `json:"to,omitempty"`
	From   string `json:"from,omitempty"`

	Private *bool  `json:"private,omitempty"`
	TopicID string `json:"t_id,omitempty"`
}

type DeployMessage struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Name      string `json:"name"`
	Tick      string `json:"tick"`
	Max       string `json:"max"`
	Limit     string `json:"lim,omitempty"`
	Metadata  string `json:"metadata,omitempty"`
	Memo      string `json:"m,omitempty"`
}

type MintMessage struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Tick      string `json:"tick"`
	Amount    string `json:"amt"`
	To        string `json:"to"`
	Memo      string `json:"m,omitempty"`
}

type BurnMessage struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Tick      string `json:"tick"`
	Amount    string `json:"amt"`
	From      string `json:"from"`
	Memo      string `json:"m,omitempty"`
}

type TransferMessage struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Tick      string `json:"tick"`
	Amount    string `json:"amt"`
	From      string `json:"from"`
	To        string `json:"to"`
	Memo      string `json:"m,omitempty"`
}

type RegisterMessage struct {
	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Name      string `json:"name"`
	Metadata  string `json:"metadata,omitempty"`
	Private   bool   `json:"private"`
	TopicID   string `json:"t_id"`
	Memo      string `json:"m,omitempty"`
}

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
	PublicTopicID      string
	RegistryTopicID    string
}

type CreateTopicOptions struct {
	Memo                string
	AdminKey            string
	SubmitKey           string
	UseOperatorAsAdmin  bool
	UseOperatorAsSubmit bool
}

type DeployPointsProgress struct {
	Stage      string
	Percentage int
	TopicID    string
	DeployTxID string
	Error      string
}

type DeployPointsOptions struct {
	Name               string
	Tick               string
	Max                string
	LimitPerMint       string
	Metadata           string
	Memo               string
	TopicMemo          string
	UsePrivateTopic    bool
	DisableMirrorCheck bool
	ProgressCallback   func(DeployPointsProgress)
}

type MintPointsProgress struct {
	Stage      string
	Percentage int
	MintTxID   string
	Error      string
}

type MintPointsOptions struct {
	Tick               string
	Amount             string
	To                 string
	Memo               string
	TopicID            string
	DisableMirrorCheck bool
	ProgressCallback   func(MintPointsProgress)
}

type TransferPointsProgress struct {
	Stage        string
	Percentage   int
	TransferTxID string
	Error        string
}

type TransferPointsOptions struct {
	Tick               string
	Amount             string
	From               string
	To                 string
	Memo               string
	TopicID            string
	DisableMirrorCheck bool
	ProgressCallback   func(TransferPointsProgress)
}

type BurnPointsProgress struct {
	Stage      string
	Percentage int
	BurnTxID   string
	Error      string
}

type BurnPointsOptions struct {
	Tick               string
	Amount             string
	From               string
	Memo               string
	TopicID            string
	DisableMirrorCheck bool
	ProgressCallback   func(BurnPointsProgress)
}

type RegisterTopicProgress struct {
	Stage        string
	Percentage   int
	RegisterTxID string
	Error        string
}

type RegisterTopicOptions struct {
	TopicID            string
	Name               string
	Metadata           string
	IsPrivate          bool
	Memo               string
	DisableMirrorCheck bool
	ProgressCallback   func(RegisterTopicProgress)
}

type PointsInfo struct {
	Name                string `json:"name"`
	Tick                string `json:"tick"`
	MaxSupply           string `json:"maxSupply"`
	LimitPerMint        string `json:"limitPerMint,omitempty"`
	Metadata            string `json:"metadata,omitempty"`
	TopicID             string `json:"topicId"`
	DeployerAccountID   string `json:"deployerAccountId"`
	CurrentSupply       string `json:"currentSupply"`
	DeploymentTimestamp string `json:"deploymentTimestamp"`
	IsPrivate           bool   `json:"isPrivate"`
}

type PointsBalance struct {
	Tick        string `json:"tick"`
	AccountID   string `json:"accountId"`
	Balance     string `json:"balance"`
	LastUpdated string `json:"lastUpdated"`
}

type PointsTransaction struct {
	ID             string `json:"id"`
	Operation      string `json:"operation"`
	Tick           string `json:"tick"`
	Amount         string `json:"amount,omitempty"`
	From           string `json:"from,omitempty"`
	To             string `json:"to,omitempty"`
	Timestamp      string `json:"timestamp"`
	SequenceNumber int64  `json:"sequenceNumber"`
	TopicID        string `json:"topicId"`
	TransactionID  string `json:"transactionId"`
	Memo           string `json:"memo,omitempty"`
}

type OperationResult struct {
	TopicID        string
	TransactionID  string
	SequenceNumber int64
	ConsensusAt    time.Time
}

type PointsState struct {
	DeployedPoints         map[string]PointsInfo
	Balances               map[string]map[string]PointsBalance
	Transactions           []PointsTransaction
	LastProcessedSequence  int64
	LastProcessedTimestamp string
}

type IndexerConfig struct {
	Network       string
	MirrorBaseURL string
	MirrorAPIKey  string
}

type IndexOptions struct {
	PublicTopicID        string
	RegistryTopicID      string
	IncludePublicTopic   bool
	IncludeRegistryTopic bool
	PrivateTopics        []string
}
