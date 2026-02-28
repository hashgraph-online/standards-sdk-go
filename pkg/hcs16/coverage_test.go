package hcs16

import (
	"testing"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestBuildCreateFloraTopicTxSuccess(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	tx, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID: "0.0.12345",
		TopicType:      FloraTopicTypeCommunication,
		AdminKey:       privKey.PublicKey(),
		SubmitKey:      privKey.PublicKey(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildCreateFloraTopicTxEmptyAccountID(t *testing.T) {
	_, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID: "",
		TopicType:      FloraTopicTypeCommunication,
	})
	if err == nil {
		t.Fatal("expected error for empty account ID")
	}
}

func TestBuildCreateFloraTopicTxInvalidType(t *testing.T) {
	_, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID: "0.0.12345",
		TopicType:      FloraTopicType(99),
	})
	if err == nil {
		t.Fatal("expected error for invalid topic type")
	}
}

func TestBuildCreateFloraTopicTxAutoRenew(t *testing.T) {
	tx, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID:   "0.0.12345",
		TopicType:        FloraTopicTypeTransaction,
		AutoRenewAccount: "0.0.99999",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildCreateFloraTopicTxAutoRenewInvalid(t *testing.T) {
	_, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID:   "0.0.12345",
		TopicType:        FloraTopicTypeState,
		AutoRenewAccount: "invalid-account",
	})
	if err == nil {
		t.Fatal("expected error for invalid auto renew account")
	}
}

func TestBuildCreateTransactionTopicTx(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	tx, err := BuildCreateTransactionTopicTx(TransactionTopicConfig{
		Memo:      "test-memo",
		AdminKey:  privKey.PublicKey(),
		SubmitKey: privKey.PublicKey(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildCreateTransactionTopicTxEmptyMemo(t *testing.T) {
	_, err := BuildCreateTransactionTopicTx(TransactionTopicConfig{Memo: ""})
	if err == nil {
		t.Fatal("expected error for empty memo")
	}
}

func TestBuildCreateTransactionTopicTxFeeConfig(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	_, err := BuildCreateTransactionTopicTx(TransactionTopicConfig{
		Memo:            "test",
		FeeScheduleKey:  privKey.PublicKey(),
	})
	if err == nil {
		t.Fatal("expected error for unsupported HIP-991")
	}
}

func TestBuildCreateFloraAccountTx(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	keyList := hedera.KeyListWithThreshold(1).Add(privKey.PublicKey())
	tx, err := BuildCreateFloraAccountTx(keyList, 5.0, 10, "test-memo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildCreateFloraAccountTxNilKeyList(t *testing.T) {
	_, err := BuildCreateFloraAccountTx(nil, 5.0, 10, "test-memo")
	if err == nil {
		t.Fatal("expected error for nil key list")
	}
}

func TestBuildCreateFloraAccountTxDefaults(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	keyList := hedera.KeyListWithThreshold(1).Add(privKey.PublicKey())
	tx, err := BuildCreateFloraAccountTx(keyList, -1, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildScheduleAccountKeyUpdateTx(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	keyList := hedera.KeyListWithThreshold(1).Add(privKey.PublicKey())
	tx, err := BuildScheduleAccountKeyUpdateTx("0.0.12345", keyList, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildScheduleAccountKeyUpdateTxEmptyID(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	keyList := hedera.KeyListWithThreshold(1).Add(privKey.PublicKey())
	_, err := BuildScheduleAccountKeyUpdateTx("", keyList, "")
	if err == nil {
		t.Fatal("expected error for empty account ID")
	}
}

func TestBuildScheduleAccountKeyUpdateTxNilKeyList(t *testing.T) {
	_, err := BuildScheduleAccountKeyUpdateTx("0.0.12345", nil, "")
	if err == nil {
		t.Fatal("expected error for nil key list")
	}
}

func TestBuildScheduleAccountKeyUpdateTxInvalidID(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	keyList := hedera.KeyListWithThreshold(1).Add(privKey.PublicKey())
	_, err := BuildScheduleAccountKeyUpdateTx("invalid", keyList, "")
	if err == nil {
		t.Fatal("expected error for invalid account ID")
	}
}

func TestBuildScheduleTopicKeyUpdateTx(t *testing.T) {
	privKey, _ := hedera.PrivateKeyGenerateEd25519()
	tx, err := BuildScheduleTopicKeyUpdateTx("0.0.12345", privKey.PublicKey(), privKey.PublicKey(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildScheduleTopicKeyUpdateTxEmptyTopicID(t *testing.T) {
	_, err := BuildScheduleTopicKeyUpdateTx("", nil, nil, "")
	if err == nil {
		t.Fatal("expected error for empty topic ID")
	}
}

func TestBuildScheduleTopicKeyUpdateTxInvalidTopicID(t *testing.T) {
	_, err := BuildScheduleTopicKeyUpdateTx("invalid", nil, nil, "")
	if err == nil {
		t.Fatal("expected error for invalid topic ID")
	}
}

func TestBuildMessageTx(t *testing.T) {
	tx, err := BuildMessageTx("0.0.12345", "0.0.1", FloraOperationFloraCreated, map[string]any{"key": "val"}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildMessageTxEmptyTopicID(t *testing.T) {
	_, err := BuildMessageTx("", "0.0.1", FloraOperationFloraCreated, nil, "")
	if err == nil {
		t.Fatal("expected error for empty topic ID")
	}
}

func TestBuildMessageTxEmptyOperatorID(t *testing.T) {
	_, err := BuildMessageTx("0.0.1", "", FloraOperationFloraCreated, nil, "")
	if err == nil {
		t.Fatal("expected error for empty operator ID")
	}
}

func TestBuildMessageTxInvalidOperation(t *testing.T) {
	_, err := BuildMessageTx("0.0.1", "0.0.2", "bad-op", nil, "")
	if err == nil {
		t.Fatal("expected error for invalid operation")
	}
}

func TestBuildMessageTxInvalidTopicID(t *testing.T) {
	_, err := BuildMessageTx("invalid", "0.0.1", FloraOperationFloraCreated, nil, "")
	if err == nil {
		t.Fatal("expected error for invalid topic ID format")
	}
}

func TestBuildFloraCreatedTxCoverage(t *testing.T) {
	tx, err := BuildFloraCreatedTx("0.0.10", "0.0.1", "0.0.100", FloraTopics{
		Communication: "0.0.11",
		Transaction:   "0.0.12",
		State:         "0.0.13",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildTransactionTxCoverage(t *testing.T) {
	tx, err := BuildTransactionTx("0.0.10", "0.0.1", "schedule-id", "test-data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildStateUpdateTxCoverage(t *testing.T) {
	epoch := int64(1)
	tx, err := BuildStateUpdateTx("0.0.10", "0.0.1", "hash123", &epoch, "0.0.5", []string{"0.0.11"}, "memo", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildStateUpdateTxEmptyTopicID(t *testing.T) {
	_, err := BuildStateUpdateTx("", "0.0.1", "hash", nil, "", nil, "", "")
	if err == nil {
		t.Fatal("expected error for empty topic ID")
	}
}

func TestBuildStateUpdateTxEmptyOperatorID(t *testing.T) {
	_, err := BuildStateUpdateTx("0.0.1", "", "hash", nil, "", nil, "", "")
	if err == nil {
		t.Fatal("expected error for empty operator ID")
	}
}

func TestBuildStateUpdateTxInvalidTopicID(t *testing.T) {
	_, err := BuildStateUpdateTx("invalid", "0.0.1", "hash", nil, "", nil, "", "")
	if err == nil {
		t.Fatal("expected error for invalid topic ID")
	}
}

func TestBuildStateUpdateTxDefaultAccountID(t *testing.T) {
	tx, err := BuildStateUpdateTx("0.0.10", "0.0.1", "hash123", nil, "", nil, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildFloraJoinRequestTx(t *testing.T) {
	tx, err := BuildFloraJoinRequestTx("0.0.10", "0.0.1", "0.0.5", 1, "0.0.20", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildFloraJoinVoteTx(t *testing.T) {
	tx, err := BuildFloraJoinVoteTx("0.0.10", "0.0.1", "0.0.5", true, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildFloraJoinAcceptedTx(t *testing.T) {
	epoch := int64(2)
	tx, err := BuildFloraJoinAcceptedTx("0.0.10", "0.0.1", []string{"0.0.5", "0.0.6"}, &epoch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildFloraJoinAcceptedTxNoEpoch(t *testing.T) {
	tx, err := BuildFloraJoinAcceptedTx("0.0.10", "0.0.1", []string{"0.0.5"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestNormalizeMemo(t *testing.T) {
	if normalizeMemo("custom", "default") != "custom" {
		t.Fatal("expected custom memo")
	}
	if normalizeMemo("", "default") != "default" {
		t.Fatal("expected default memo")
	}
	if normalizeMemo("   ", "default") != "default" {
		t.Fatal("expected default for whitespace")
	}
}

func TestEncodeFloraTopicMemo(t *testing.T) {
	memo := encodeFloraTopicMemo("0.0.12345", FloraTopicTypeCommunication)
	if memo != "hcs-16:0.0.12345:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestIsValidFloraTopicType(t *testing.T) {
	if !isValidFloraTopicType(FloraTopicTypeCommunication) {
		t.Fatal("expected communication to be valid")
	}
	if !isValidFloraTopicType(FloraTopicTypeTransaction) {
		t.Fatal("expected transaction to be valid")
	}
	if !isValidFloraTopicType(FloraTopicTypeState) {
		t.Fatal("expected state to be valid")
	}
	if isValidFloraTopicType(FloraTopicType(99)) {
		t.Fatal("expected 99 to be invalid")
	}
}

func TestIsValidFloraOperation(t *testing.T) {
	ops := []FloraOperation{
		FloraOperationFloraCreated, FloraOperationTransaction, FloraOperationStateUpdate,
		FloraOperationJoinRequest, FloraOperationJoinVote, FloraOperationJoinAccepted,
	}
	for _, op := range ops {
		if !isValidFloraOperation(op) {
			t.Fatalf("expected %q to be valid", op)
		}
	}
	if isValidFloraOperation("bad") {
		t.Fatal("expected 'bad' to be invalid")
	}
}

func TestExtractMirrorKeyString(t *testing.T) {
	result := extractMirrorKeyString(nil)
	if result != "" {
		t.Fatalf("expected empty string for nil, got %q", result)
	}
}

func TestExtractMirrorKeyCandidateCoverage(t *testing.T) {
	result := extractMirrorKeyCandidate("abc123")
	if result != "abc123" {
		t.Fatalf("expected 'abc123', got %q", result)
	}

	mapResult := extractMirrorKeyCandidate(map[string]any{"key": "mykey"})
	if mapResult != "mykey" {
		t.Fatalf("expected 'mykey', got %q", mapResult)
	}

	ecdsaResult := extractMirrorKeyCandidate(map[string]any{"ECDSA_secp256k1": "eckey"})
	if ecdsaResult != "eckey" {
		t.Fatalf("expected 'eckey', got %q", ecdsaResult)
	}

	ed25519Result := extractMirrorKeyCandidate(map[string]any{"ed25519": "edkey"})
	if ed25519Result != "edkey" {
		t.Fatalf("expected 'edkey', got %q", ed25519Result)
	}

	nested := extractMirrorKeyCandidate(map[string]any{"nested": map[string]any{"key": "deepkey"}})
	if nested != "deepkey" {
		t.Fatalf("expected 'deepkey', got %q", nested)
	}

	emptyResult := extractMirrorKeyCandidate(42)
	if emptyResult != "" {
		t.Fatalf("expected empty for unsupported type, got %q", emptyResult)
	}
}
