package main

import "testing"

func TestCanonicalLedgerNetwork(t *testing.T) {
	if value := canonicalLedgerNetwork("mainnet"); value != "hedera:mainnet" {
		t.Fatalf("expected hedera:mainnet, got %s", value)
	}
	if value := canonicalLedgerNetwork("testnet"); value != "hedera:testnet" {
		t.Fatalf("expected hedera:testnet, got %s", value)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	value := firstNonEmpty("", " ", "ok")
	if value != "ok" {
		t.Fatalf("expected ok, got %s", value)
	}
}
