package hcs15

import (
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

// BuildBaseAccountCreateTx builds and returns the configured value.
func BuildBaseAccountCreateTx(params BaseAccountCreateTxParams) (*hedera.AccountCreateTransaction, error) {
	if params.PublicKey.String() == "" {
		return nil, fmt.Errorf("public key is required")
	}

	initialBalance := params.InitialBalanceHbar
	if initialBalance <= 0 {
		initialBalance = 1
	}

	transaction := hedera.NewAccountCreateTransaction().
		SetKey(params.PublicKey).
		SetAlias(params.PublicKey.ToEvmAddress()).
		SetInitialBalance(hedera.NewHbar(initialBalance)).
		SetTransactionMemo(normalizeMemo(params.TransactionMemo, HCS15BaseAccountCreateTransactionMemo))

	if params.MaxAutomaticTokenAssociations != nil {
		transaction.SetMaxAutomaticTokenAssociations(*params.MaxAutomaticTokenAssociations)
	}
	if strings.TrimSpace(params.AccountMemo) != "" {
		transaction.SetAccountMemo(strings.TrimSpace(params.AccountMemo))
	}

	return transaction, nil
}

// BuildPetalAccountCreateTx builds and returns the configured value.
func BuildPetalAccountCreateTx(params PetalAccountCreateTxParams) (*hedera.AccountCreateTransaction, error) {
	if params.PublicKey.String() == "" {
		return nil, fmt.Errorf("public key is required")
	}

	initialBalance := params.InitialBalanceHbar
	if initialBalance <= 0 {
		initialBalance = 1
	}

	transaction := hedera.NewAccountCreateTransaction().
		SetKey(params.PublicKey).
		SetInitialBalance(hedera.NewHbar(initialBalance)).
		SetTransactionMemo(normalizeMemo(params.TransactionMemo, HCS15PetalAccountCreateTransactionMemo))

	if params.MaxAutomaticTokenAssociations != nil {
		transaction.SetMaxAutomaticTokenAssociations(*params.MaxAutomaticTokenAssociations)
	}
	if strings.TrimSpace(params.AccountMemo) != "" {
		transaction.SetAccountMemo(strings.TrimSpace(params.AccountMemo))
	}

	return transaction, nil
}

type BaseAccountCreateTxParams struct {
	PublicKey                     hedera.PublicKey
	InitialBalanceHbar            float64
	MaxAutomaticTokenAssociations *int32
	AccountMemo                   string
	TransactionMemo               string
}

type PetalAccountCreateTxParams struct {
	PublicKey                     hedera.PublicKey
	InitialBalanceHbar            float64
	MaxAutomaticTokenAssociations *int32
	AccountMemo                   string
	TransactionMemo               string
}

func normalizeMemo(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
