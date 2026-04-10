package hcs15

import hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

const (
	HCS15BaseAccountCreateTransactionMemo  = "hcs-15:op:base_create"
	HCS15PetalAccountCreateTransactionMemo = "hcs-15:op:petal_create"
)

type ClientConfig struct {
	OperatorAccountID  string
	OperatorPrivateKey string
	Network            string
	MirrorBaseURL      string
	MirrorAPIKey       string
	HederaClient       *hedera.Client
}

type BaseAccountCreateOptions struct {
	InitialBalanceHbar            float64
	MaxAutomaticTokenAssociations *int32
	AccountMemo                   string
	TransactionMemo               string
}

type PetalAccountCreateOptions struct {
	BasePrivateKey                string
	InitialBalanceHbar            float64
	MaxAutomaticTokenAssociations *int32
	AccountMemo                   string
	TransactionMemo               string
}

type BaseAccountCreateResult struct {
	AccountID     string
	PrivateKey    hedera.PrivateKey
	PrivateKeyRaw string
	PublicKey     hedera.PublicKey
	EVMAddress    string
	Receipt       hedera.TransactionReceipt
}

type PetalAccountCreateResult struct {
	AccountID string
	Receipt   hedera.TransactionReceipt
}
