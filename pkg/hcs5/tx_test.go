package hcs5

import (
	"testing"
)

func TestBuildHCS1HRL(t *testing.T) {
	hrl := BuildHCS1HRL("0.0.123456")
	if hrl != "hcs://1/0.0.123456" {
		t.Fatalf("unexpected HRL: %s", hrl)
	}
}

func TestBuildMintWithHRLTx(t *testing.T) {
	transaction, err := BuildMintWithHRLTx("0.0.4321", "0.0.123456", "mint test")
	if err != nil {
		t.Fatalf("BuildMintWithHRLTx failed: %v", err)
	}
	if transaction.GetTokenID().String() != "0.0.4321" {
		t.Fatalf("unexpected token ID: %s", transaction.GetTokenID().String())
	}
	if transaction.GetTransactionMemo() != "mint test" {
		t.Fatalf("unexpected transaction memo: %s", transaction.GetTransactionMemo())
	}
	metadata := transaction.GetMetadatas()
	if len(metadata) != 1 {
		t.Fatalf("expected one metadata entry")
	}
	if string(metadata[0]) != "hcs://1/0.0.123456" {
		t.Fatalf("unexpected metadata: %s", string(metadata[0]))
	}
}
