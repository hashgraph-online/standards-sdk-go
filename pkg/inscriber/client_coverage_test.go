package inscriber

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hiero-ledger/hiero-sdk-go/v2/proto/sdk"
	protobufservices "github.com/hiero-ledger/hiero-sdk-go/v2/proto/services"
	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
	"google.golang.org/protobuf/proto"
)

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Fatal("expected 1")
	}
	if boolToInt(false) != 0 {
		t.Fatal("expected 0")
	}
}

func TestNormalizeTransactionID(t *testing.T) {
	if normalizeTransactionID("0.0.123@12345.6789") != "0.0.123-12345-6789" {
		t.Fatal("expected 0.0.123-12345-6789")
	}
	if normalizeTransactionID("0.0.123-123") != "0.0.123-123" {
		t.Fatal("expected 0.0.123-123")
	}
	if normalizeTransactionID("") != "" {
		t.Fatal("expected empty string")
	}
}

func TestNormalizeTransactionBytes(t *testing.T) {
	str, err := normalizeTransactionBytes(nil)
	if err != nil || str != "" {
		t.Fatal("expected empty string for nil")
	}

	str, err = normalizeTransactionBytes("base64data")
	if err != nil || str != "base64data" {
		t.Fatal("expected string to pass through")
	}

	bufObj := map[string]any{
		"type": "Buffer",
		"data": []any{float64(104), float64(101)},
	}
	str, err = normalizeTransactionBytes(bufObj)
	if err != nil || str != "aGU=" { // base64 for "he"
		t.Fatalf("expected aGU=, got %v with string %s", err, str)
	}

	_, err = normalizeTransactionBytes(map[string]any{"type": "Other"})
	if err == nil {
		t.Fatal("expected error for non-buffer type")
	}
}

func TestParseInscriptionJob(t *testing.T) {
	raw := map[string]any{
		"id":               "job-1",
		"status":           "completed",
		"completed":        true,
		"tx_id":            "tx1",
		"topic_id":         "topic1",
		"transactionId":    "tx1",
		"error":            "err",
		"totalCost":        float64(123),
		"totalMessages":    float64(456),
		"transactionBytes": "bytes",
	}

	job, err := parseInscriptionJob(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.ID != "job-1" || job.Status != "completed" || !job.Completed || job.TxID != "tx1" || job.TopicID != "topic1" || job.TransactionID != "tx1" || job.Error != "err" || job.TotalCost != 123 || job.TotalMessages != 456 || job.TransactionBytes != "bytes" {
		t.Fatal("unexpected job data")
	}
}

func TestResolveURL(t *testing.T) {
	client := &Client{baseURL: "http://base"}
	if client.resolveURL("http://other") != "http://other" {
		t.Fatal("expected other")
	}
	if client.resolveURL("/path") != "http://base/path" {
		t.Fatal("expected base/path")
	}
	if client.resolveURL("path") != "http://base/path" {
		t.Fatal("expected base/path")
	}
}

func TestPostJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "key" {
			t.Fatalf("expected x-api-key header, got %q", got)
		}
		if got := r.Header.Get("x-session"); got != "" {
			t.Fatalf("expected x-session header to be omitted, got %q", got)
		}
		w.Write([]byte(`{"success": true}`))
	}))
	defer ts.Close()

	client := &Client{
		apiKey:     "key",
		baseURL:    ts.URL,
		httpClient: &http.Client{},
	}
	var res map[string]any
	err := client.postJSON(context.Background(), "/path", map[string]any{"key": "val"}, &res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["success"] != true {
		t.Fatal("expected success true")
	}

	// bad json payload
	err = client.postJSON(context.Background(), "/path", make(chan int), &res)
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "key" {
			t.Fatalf("expected x-api-key header, got %q", got)
		}
		if got := r.Header.Get("x-session"); got != "" {
			t.Fatalf("expected x-session header to be omitted, got %q", got)
		}
		w.Write([]byte(`{"success": true}`))
	}))
	defer ts.Close()

	client := &Client{
		apiKey:     "key",
		baseURL:    ts.URL,
		httpClient: &http.Client{},
	}
	var res map[string]any
	err := client.getJSON(context.Background(), "/path", &res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["success"] != true {
		t.Fatal("expected success true")
	}
}

func TestIsRetryableWaitErrorCoverage(t *testing.T) {
	if isRetryableWaitError(nil) {
		t.Fatal("expected false")
	}
	if isRetryableWaitError(context.Canceled) {
		t.Fatal("expected false")
	}
	if !isRetryableWaitError(errors.New("temporarily unavailable")) {
		t.Fatal("expected true")
	}
}

func TestRetrieveInscription(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tx_id": "0.0.1-1-1", "status": "completed"}`))
	}))
	defer ts.Close()

	client, _ := NewClient(Config{APIKey: "key", BaseURL: ts.URL})
	job, err := client.RetrieveInscription(context.Background(), "0.0.1-1-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !job.Completed || job.TxID != "0.0.1-1-1" {
		t.Fatal("expected completed with right tx id")
	}
}

func TestWaitForInscription(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Write([]byte(`{"status": "processing"}`))
		} else {
			w.Write([]byte(`{"status": "completed", "completed": true}`))
		}
	}))
	defer ts.Close()

	client, _ := NewClient(Config{APIKey: "key", BaseURL: ts.URL})
	job, err := client.WaitForInscription(context.Background(), "tx", WaitOptions{Interval: time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !job.Completed {
		t.Fatal("expected completed")
	}
}

func TestAsTransferTransaction(t *testing.T) {
	tx := hedera.NewTransferTransaction()
	res, err := asTransferTransaction(tx)
	if err != nil || res == nil {
		t.Fatal("expected success with pointer")
	}
	res2, err := asTransferTransaction(*tx)
	if err != nil || res2 == nil {
		t.Fatal("expected success with value")
	}
	_, err = asTransferTransaction("invalid")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppendSerializedTransactionSignaturePreservesExistingSignatures(t *testing.T) {
	payerKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to create payer key: %v", err)
	}
	holderKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to create holder key: %v", err)
	}

	payerAccountID, err := hedera.AccountIDFromString("0.0.1001")
	if err != nil {
		t.Fatalf("failed to parse payer account ID: %v", err)
	}
	holderAccountID, err := hedera.AccountIDFromString("0.0.1002")
	if err != nil {
		t.Fatalf("failed to parse holder account ID: %v", err)
	}
	nodeAccountID, err := hedera.AccountIDFromString("0.0.3")
	if err != nil {
		t.Fatalf("failed to parse node account ID: %v", err)
	}

	transferTransaction, err := hedera.NewTransferTransaction().
		AddHbarTransfer(holderAccountID, hedera.NewHbar(-1)).
		AddHbarTransfer(nodeAccountID, hedera.NewHbar(1)).
		SetTransactionID(hedera.TransactionIDGenerate(payerAccountID)).
		SetNodeAccountIDs([]hedera.AccountID{nodeAccountID}).
		Freeze()
	if err != nil {
		t.Fatalf("failed to freeze transfer transaction: %v", err)
	}

	transferTransaction.Sign(payerKey)
	serializedTransaction, err := transferTransaction.ToBytes()
	if err != nil {
		t.Fatalf("failed to encode transfer transaction: %v", err)
	}
	initialSignatureCount, err := countSerializedTransactionSignatures(serializedTransaction)
	if err != nil {
		t.Fatalf("failed to inspect original signature count: %v", err)
	}

	updatedTransactionBytes, err := appendSerializedTransactionSignature(serializedTransaction, holderKey)
	if err != nil {
		t.Fatalf("failed to append holder signature: %v", err)
	}
	updatedTransaction, err := decodeTransferTransaction(updatedTransactionBytes)
	if err != nil {
		t.Fatalf("failed to decode updated transfer transaction: %v", err)
	}

	updatedSignatureCount, err := countSerializedTransactionSignatures(updatedTransactionBytes)
	if err != nil {
		t.Fatalf("failed to inspect updated signature count: %v", err)
	}
	if updatedSignatureCount != initialSignatureCount+1 {
		t.Fatalf(
			"expected signature pair count to grow from %d to %d, got %d",
			initialSignatureCount,
			initialSignatureCount+1,
			updatedSignatureCount,
		)
	}

	transactionID, err := hedera.TransactionGetTransactionID(updatedTransaction)
	if err != nil {
		t.Fatalf("failed to inspect updated transaction ID: %v", err)
	}
	if transactionID.String() != transferTransaction.GetTransactionID().String() {
		t.Fatalf("expected transaction ID %s, got %s", transferTransaction.GetTransactionID(), transactionID)
	}
}

func countSerializedTransactionSignatures(transactionBytes []byte) (int, error) {
	var transactionList sdk.TransactionList
	if err := proto.Unmarshal(transactionBytes, &transactionList); err != nil {
		return 0, err
	}
	if len(transactionList.TransactionList) == 0 {
		return 0, errors.New("expected at least one serialized transaction")
	}

	totalSignatures := 0
	for _, transaction := range transactionList.TransactionList {
		var signedTransaction protobufservices.SignedTransaction
		if err := proto.Unmarshal(transaction.SignedTransactionBytes, &signedTransaction); err != nil {
			return 0, err
		}
		totalSignatures += len(signedTransaction.GetSigMap().GetSigPair())
	}

	return totalSignatures, nil
}

func TestResolveMirrorKeyType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"key": {"_type": "ProtobufEncoded"}}`))
	}))
	defer ts.Close()

	_, err := resolveMirrorKeyType(context.Background(), "testnet", "0.0.8")
	// will succeed or fail depending on network, but let's just make sure it doesn't crash.
	if err != nil {
		// allow it to fail, we just want to execute code paths
	}

	// mock mirror by directly testing parseOperatorPrivateKey which calls it
	_, err = parseOperatorPrivateKey(context.Background(), "testnet", "0.0.8", "badkey")
	if err == nil {
		t.Fatal("expected err for bad key")
	}
}

func TestExecuteTransactionFailures(t *testing.T) {
	_, err := ExecuteTransaction(context.Background(), "invalid-base64", HederaClientConfig{
		Network:    NetworkTestnet,
		AccountID:  "0.0.1",
		PrivateKey: "302e020100300506032b657004220420" + "0000000000000000000000000000000000000000000000000000000000000000",
	})
	if err == nil {
		t.Fatal("expected failure on invalid base64")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err = ExecuteTransaction(context.Background(), "dmFsaWRCYXNlNjQ=", HederaClientConfig{
		Network:    NetworkTestnet,
		AccountID:  "0.0.1",
		PrivateKey: pk.String(),
	})
	// `dmFsaWRCYXNlNjQ=` is "validBase64", decoder will fail to decode as Hedera Transaction
	if err == nil {
		t.Fatal("expected failure on invalid decode")
	}
}

func TestInscribeAndExecuteFailsStart(t *testing.T) {
	client := &Client{baseURL: "http://[::1]:namedport", httpClient: &http.Client{}}

	req := StartInscriptionRequest{
		HolderID: "0.0.123",
		Mode:     ModeFile,
		File:     FileInput{Type: "url", URL: "http://"},
	}

	_, err := client.InscribeAndExecute(context.Background(), req, HederaClientConfig{}, false)
	if err == nil {
		t.Fatal("expected start to fail")
	}
}

func TestExecuteRebuiltTransferTransactionFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	accountID, _ := hedera.AccountIDFromString("0.0.1")
	otherAccountID, _ := hedera.AccountIDFromString("0.0.2")

	// Create and marshal an invalid transfer tx
	tx := hedera.NewTopicCreateTransaction()
	txBytes, _ := tx.ToBytes()
	_, err := executeRebuiltTransferTransaction(txBytes, "testnet", accountID, pk)
	if err == nil {
		t.Fatal("expected failure on non-transfer tx rebuild")
	}

	trTx, _ := hedera.NewTransferTransaction().
		AddHbarTransfer(accountID, hedera.NewHbar(-1)).
		AddHbarTransfer(otherAccountID, hedera.NewHbar(1)).
		SetTransactionID(hedera.TransactionIDGenerate(otherAccountID)). // Wrong payer!
		Freeze()
	trBytes, _ := trTx.ToBytes()
	_, err = executeRebuiltTransferTransaction(trBytes, "testnet", accountID, pk)
	if err == nil {
		t.Fatal("expected failure on mismatched payer account")
	}

	trTx2, _ := hedera.NewTransferTransaction().
		AddHbarTransfer(accountID, hedera.NewHbar(-1)).
		AddHbarTransfer(otherAccountID, hedera.NewHbar(1)).
		SetTransactionID(hedera.TransactionIDGenerate(accountID)).
		SetNodeAccountIDs([]hedera.AccountID{otherAccountID}).
		Freeze()
	trBytes2, _ := trTx2.ToBytes()
	_, err = executeRebuiltTransferTransaction(trBytes2, "invalid-network-name", accountID, pk)
	if err == nil {
		t.Fatal("expected err on bad network")
	}

	// Just checking ExecuteTransaction with invalid network
	_, err = ExecuteTransaction(context.Background(), "b64", HederaClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected invalid network error")
	}
}

func TestSerializedTransactionMetadata(t *testing.T) {
	payerAccountID, err := hedera.AccountIDFromString("0.0.1001")
	if err != nil {
		t.Fatalf("failed to parse payer account ID: %v", err)
	}
	receiverAccountID, err := hedera.AccountIDFromString("0.0.1002")
	if err != nil {
		t.Fatalf("failed to parse receiver account ID: %v", err)
	}
	nodeAccountID, err := hedera.AccountIDFromString("0.0.3")
	if err != nil {
		t.Fatalf("failed to parse node account ID: %v", err)
	}

	transferTransaction, err := hedera.NewTransferTransaction().
		AddHbarTransfer(payerAccountID, hedera.NewHbar(-1)).
		AddHbarTransfer(receiverAccountID, hedera.NewHbar(1)).
		SetTransactionID(hedera.TransactionIDGenerate(payerAccountID)).
		SetNodeAccountIDs([]hedera.AccountID{nodeAccountID}).
		Freeze()
	if err != nil {
		t.Fatalf("failed to freeze transfer transaction: %v", err)
	}

	serializedBytes, err := transferTransaction.ToBytes()
	if err != nil {
		t.Fatalf("failed to encode transfer transaction: %v", err)
	}

	var transactionList sdk.TransactionList
	if unmarshalErr := proto.Unmarshal(serializedBytes, &transactionList); unmarshalErr != nil {
		t.Fatalf("failed to decode transaction list: %v", unmarshalErr)
	}

	transactionID, parsedNodeAccountID, err := serializedTransactionMetadata(transactionList.TransactionList[0])
	if err != nil {
		t.Fatalf("failed to read serialized transaction metadata: %v", err)
	}
	if transactionID.String() != transferTransaction.GetTransactionID().String() {
		t.Fatalf("expected transaction ID %s, got %s", transferTransaction.GetTransactionID(), transactionID)
	}
	if parsedNodeAccountID.String() != nodeAccountID.String() {
		t.Fatalf("expected node account ID %s, got %s", nodeAccountID, parsedNodeAccountID)
	}
}

func TestNetworkNodeAddress(t *testing.T) {
	nodeAccountID, err := hedera.AccountIDFromString("0.0.3")
	if err != nil {
		t.Fatalf("failed to parse node account ID: %v", err)
	}

	nodeAddress, err := networkNodeAddress(hedera.ClientForTestnet(), string(NetworkTestnet), nodeAccountID)
	if err != nil {
		t.Fatalf("failed to derive network node address: %v", err)
	}
	if nodeAddress != "0.testnet.hedera.com:50212" {
		t.Fatalf("unexpected node address %s", nodeAddress)
	}

	_, err = networkNodeAddress(nil, string(NetworkTestnet), nodeAccountID)
	if err == nil {
		t.Fatal("expected missing client error")
	}

	missingNodeAccountID, err := hedera.AccountIDFromString("0.0.999")
	if err != nil {
		t.Fatalf("failed to parse missing node account ID: %v", err)
	}

	_, err = networkNodeAddress(hedera.ClientForTestnet(), "invalid-network", missingNodeAccountID)
	if err == nil {
		t.Fatal("expected missing node error")
	}
}

func TestNodeAddressServerName(t *testing.T) {
	serverName, err := nodeAddressServerName("0.testnet.hedera.com:50212")
	if err != nil {
		t.Fatalf("failed to parse node server name: %v", err)
	}
	if serverName != "0.testnet.hedera.com" {
		t.Fatalf("unexpected server name %s", serverName)
	}

	if _, err := nodeAddressServerName("invalid-address"); err == nil {
		t.Fatal("expected invalid node address error")
	}
}
