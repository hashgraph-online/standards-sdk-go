package hcs15

import (
	"testing"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestBuildBaseAccountCreateTx(t *testing.T) {
	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	maxAssociations := int32(10)
	transaction, err := BuildBaseAccountCreateTx(BaseAccountCreateTxParams{
		PublicKey:                     privateKey.PublicKey(),
		InitialBalanceHbar:            1,
		MaxAutomaticTokenAssociations: &maxAssociations,
		AccountMemo:                   "base-account",
	})
	if err != nil {
		t.Fatalf("BuildBaseAccountCreateTx failed: %v", err)
	}

	if len(transaction.GetAlias()) == 0 {
		t.Fatalf("expected base account transaction alias to be set")
	}
	if transaction.GetInitialBalance().AsTinybar() <= 0 {
		t.Fatalf("expected initial balance to be positive")
	}
	if transaction.GetMaxAutomaticTokenAssociations() != maxAssociations {
		t.Fatalf("unexpected max automatic token associations: %d", transaction.GetMaxAutomaticTokenAssociations())
	}
	if transaction.GetTransactionMemo() != HCS15BaseAccountCreateTransactionMemo {
		t.Fatalf("unexpected transaction memo: %s", transaction.GetTransactionMemo())
	}
}

func TestBuildPetalAccountCreateTx(t *testing.T) {
	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	maxAssociations := int32(100)
	transaction, err := BuildPetalAccountCreateTx(PetalAccountCreateTxParams{
		PublicKey:                     privateKey.PublicKey(),
		InitialBalanceHbar:            1,
		MaxAutomaticTokenAssociations: &maxAssociations,
		AccountMemo:                   "petal-account",
	})
	if err != nil {
		t.Fatalf("BuildPetalAccountCreateTx failed: %v", err)
	}

	if len(transaction.GetAlias()) != 0 {
		t.Fatalf("expected petal account transaction alias to be unset")
	}
	if transaction.GetInitialBalance().AsTinybar() <= 0 {
		t.Fatalf("expected initial balance to be positive")
	}
	if transaction.GetMaxAutomaticTokenAssociations() != maxAssociations {
		t.Fatalf("unexpected max automatic token associations: %d", transaction.GetMaxAutomaticTokenAssociations())
	}
	if transaction.GetTransactionMemo() != HCS15PetalAccountCreateTransactionMemo {
		t.Fatalf("unexpected transaction memo: %s", transaction.GetTransactionMemo())
	}
}

func TestBuildBaseAccountCreateTx_WithMemoOverride(t *testing.T) {
	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	transaction, err := BuildBaseAccountCreateTx(BaseAccountCreateTxParams{
		PublicKey:       privateKey.PublicKey(),
		TransactionMemo: "custom-memo",
	})
	if err != nil {
		t.Fatalf("BuildBaseAccountCreateTx failed: %v", err)
	}
	if transaction.GetTransactionMemo() != "custom-memo" {
		t.Fatalf("expected memo override, got %s", transaction.GetTransactionMemo())
	}
}

func TestBuildPetalAccountCreateTx_WithMemoOverride(t *testing.T) {
	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	transaction, err := BuildPetalAccountCreateTx(PetalAccountCreateTxParams{
		PublicKey:       privateKey.PublicKey(),
		TransactionMemo: "petal-custom",
	})
	if err != nil {
		t.Fatalf("BuildPetalAccountCreateTx failed: %v", err)
	}
	if transaction.GetTransactionMemo() != "petal-custom" {
		t.Fatalf("expected memo override, got %s", transaction.GetTransactionMemo())
	}
}
