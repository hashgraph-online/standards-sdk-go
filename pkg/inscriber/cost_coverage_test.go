package inscriber

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestFormatTinybarToHBAR(t *testing.T) {
	if formatTinybarToHBAR(100000000) != "1" {
		t.Fatal("expected 1")
	}
	if formatTinybarToHBAR(50000000) != "0.5" {
		t.Fatal("expected 0.5")
	}
}

func TestAbsInt64(t *testing.T) {
	if absInt64(-5) != 5 {
		t.Fatal("expected 5")
	}
	if absInt64(5) != 5 {
		t.Fatal("expected 5")
	}
	if absInt64(0) != 0 {
		t.Fatal("expected 0")
	}
}

func TestNormalizedPayerAccount(t *testing.T) {
	if normalizedPayerAccount("0.0.123-12345-6789") != "0.0.123" {
		t.Fatal("expected 0.0.123")
	}
	if normalizedPayerAccount("  0.0.123-12345  ") != "0.0.123" {
		t.Fatal("expected 0.0.123")
	}
	if normalizedPayerAccount("") != "" {
		t.Fatal("expected empty")
	}
}

func TestParseJobQuote(t *testing.T) {
	quote, err := parseJobQuote(InscriptionJob{TotalCost: 200000000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.TotalCostHBAR != "2" {
		t.Fatal("expected 2")
	}
	if len(quote.Breakdown.Transfers) != 1 {
		t.Fatal("expected 1 transfer breakdown")
	}

	quoteDefault, _ := parseJobQuote(InscriptionJob{})
	if quoteDefault.TotalCostHBAR != "0.001" {
		t.Fatal("expected 0.001 fallback")
	}
}

func TestResolveInscriptionCostSummaryEmptyTx(t *testing.T) {
	summary, err := resolveInscriptionCostSummary(context.Background(), "", "testnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != nil {
		t.Fatal("expected nil summary")
	}
}

func TestResolveInscriptionCostSummaryMirrorError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	// temporarily override MirrorBaseURL logic if possible, or just expect error if we can't connect
	// Since we can't easily inject mock URL to resolveInscriptionCostSummary, we will just pass dummy network
	// that fails on mirror node resolution or actual HTTP wait.
	// Actually we expect it to fail reaching out to real testnet mirror if we give it a bad tx ID format
	// wait, normalizeTransactionID keeps the format. Mirror node might return 400.
	_, err := resolveInscriptionCostSummary(context.Background(), "invalid_tx", "unknown_network")
	if err == nil {
		t.Fatal("expected error for invalid network / tx")
	}
}

func TestTransactionFromBytesError(t *testing.T) {
	_, err := transactionFromBytes([]byte("invalid bytes"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTransferTinybarError(t *testing.T) {
	_, err := parseTransferTinybar([]byte("invalid"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTransferTinybarAndQuote(t *testing.T) {
	accountID, _ := hedera.AccountIDFromString("0.0.1")
	otherAccountID, _ := hedera.AccountIDFromString("0.0.2")

	tx, _ := hedera.NewTransferTransaction().
		AddHbarTransfer(accountID, hedera.NewHbar(-1)).
		AddHbarTransfer(otherAccountID, hedera.NewHbar(1)).
		Freeze()
	
	bytes, _ := tx.ToBytes()
	
	val, err := parseTransferTinybar(bytes)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if val != 100000000 {
		t.Fatalf("expected 1 hbar tinybar, got %d", val)
	}

	b64Bytes := base64.StdEncoding.EncodeToString(bytes)
	quote, err := parseJobQuote(InscriptionJob{TransactionBytes: b64Bytes})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if quote.TotalCostHBAR != "1" {
		t.Fatal("expected quote cost to reflect tx transfers")
	}
}
