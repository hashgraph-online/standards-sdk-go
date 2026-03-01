package hcs2

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestBuildTopicMemoIndexed(t *testing.T) {
	memo := BuildTopicMemo(RegistryTypeIndexed, 86400)
	if memo != "hcs-2:0:86400" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildTopicMemoNonIndexed(t *testing.T) {
	memo := BuildTopicMemo(RegistryTypeNonIndexed, 3600)
	if memo != "hcs-2:1:3600" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseTopicMemoSuccess(t *testing.T) {
	parsed, ok := ParseTopicMemo("hcs-2:0:86400")
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if parsed.RegistryType != RegistryTypeIndexed {
		t.Fatalf("unexpected registry type: %d", parsed.RegistryType)
	}
	if parsed.TTL != 86400 {
		t.Fatalf("unexpected TTL: %d", parsed.TTL)
	}
}

func TestParseTopicMemoNonIndexed(t *testing.T) {
	parsed, ok := ParseTopicMemo("hcs-2:1:3600")
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if parsed.RegistryType != RegistryTypeNonIndexed {
		t.Fatalf("expected non-indexed, got %d", parsed.RegistryType)
	}
}

func TestParseTopicMemoInvalidCases(t *testing.T) {
	cases := []string{
		"", "hcs-2", "hcs-2:0", "hcs-2:0:86400:extra",
		"bad:0:86400", "hcs-2:x:86400", "hcs-2:0:abc",
		"hcs-2:5:86400",
	}
	for _, tc := range cases {
		_, ok := ParseTopicMemo(tc)
		if ok {
			t.Fatalf("expected parse to fail for %q", tc)
		}
	}
}

func TestCoverageBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo(OperationRegister, RegistryTypeIndexed)
	if memo != "hcs-2:op:0:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildTransactionMemoUpdate(t *testing.T) {
	memo := BuildTransactionMemo(OperationUpdate, RegistryTypeIndexed)
	if memo != "hcs-2:op:1:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildTransactionMemoDelete(t *testing.T) {
	memo := BuildTransactionMemo(OperationDelete, RegistryTypeIndexed)
	if memo != "hcs-2:op:2:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildTransactionMemoMigrate(t *testing.T) {
	memo := BuildTransactionMemo(OperationMigrate, RegistryTypeNonIndexed)
	if memo != "hcs-2:op:3:1" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestValidateMessageRegisterSuccess(t *testing.T) {
	msg := Message{
		P:       "hcs-2",
		Op:      OperationRegister,
		TopicID: "0.0.12345",
	}
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMessageUpdateSuccess(t *testing.T) {
	msg := Message{
		P:       "hcs-2",
		Op:      OperationUpdate,
		UID:     "uid-123",
		TopicID: "0.0.12345",
	}
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMessageDeleteSuccess(t *testing.T) {
	msg := Message{
		P:   "hcs-2",
		Op:  OperationDelete,
		UID: "uid-123",
	}
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMessageMigrateSuccess(t *testing.T) {
	msg := Message{
		P:       "hcs-2",
		Op:      OperationMigrate,
		TopicID: "0.0.99999",
	}
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateMessageBadProtocol(t *testing.T) {
	msg := Message{P: "bad", Op: OperationRegister, TopicID: "0.0.1"}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for bad protocol")
	}
}

func TestValidateMessageBadOperation(t *testing.T) {
	msg := Message{P: "hcs-2", Op: "bad", TopicID: "0.0.1"}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for bad operation")
	}
}

func TestValidateMessageMemoTooLong(t *testing.T) {
	msg := Message{
		P:       "hcs-2",
		Op:      OperationRegister,
		TopicID: "0.0.1",
		Memo:    strings.Repeat("x", 501),
	}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for memo too long")
	}
}

func TestValidateMessageRegisterMissingTopicID(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationRegister}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for register missing t_id")
	}
}

func TestValidateMessageUpdateMissingUID(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationUpdate, TopicID: "0.0.1"}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for update missing uid")
	}
}

func TestValidateMessageUpdateMissingTopicID(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationUpdate, UID: "uid-1"}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for update missing t_id")
	}
}

func TestValidateMessageDeleteMissingUID(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationDelete}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for delete missing uid")
	}
}

func TestValidateMessageMigrateMissingTopicID(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationMigrate}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for migrate missing t_id")
	}
}

func TestValidateMessageNegativeTTL(t *testing.T) {
	msg := Message{P: "hcs-2", Op: OperationRegister, TopicID: "0.0.1", TTL: -1}
	if err := ValidateMessage(msg); err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestValidateMessageCustomProtocol(t *testing.T) {
	msg := Message{P: "hcs-11", Op: OperationRegister, TopicID: "0.0.1"}
	if err := ValidateMessage(msg); err != nil {
		t.Fatalf("unexpected error for custom protocol: %v", err)
	}
}

func TestCoverageBuildHCS2CreateRegistryTx(t *testing.T) {
	tx := BuildHCS2CreateRegistryTx(CreateRegistryTxParams{
		RegistryType: RegistryTypeIndexed,
		TTL:          86400,
	})
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2CreateRegistryTxDefaults(t *testing.T) {
	tx := BuildHCS2CreateRegistryTx(CreateRegistryTxParams{
		RegistryType: RegistryType(99),
		TTL:          -1,
	})
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2CreateRegistryTxMemoOverride(t *testing.T) {
	tx := BuildHCS2CreateRegistryTx(CreateRegistryTxParams{
		RegistryType: RegistryTypeIndexed,
		TTL:          86400,
		MemoOverride: "custom-memo",
	})
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestCoverageBuildHCS2RegisterTx(t *testing.T) {
	tx, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "0.0.12345",
		TargetTopicID:   "0.0.54321",
		Metadata:        "my-meta",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2RegisterTxCustomProtocol(t *testing.T) {
	tx, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "0.0.12345",
		TargetTopicID:   "0.0.54321",
		Protocol:        "hcs-11",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2RegisterTxEmptyTopicID(t *testing.T) {
	_, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "",
		TargetTopicID:   "0.0.1",
	})
	if err == nil {
		t.Fatal("expected error for empty registry topic ID")
	}
}

func TestBuildHCS2RegisterTxInvalidTopicID(t *testing.T) {
	_, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "not-valid",
		TargetTopicID:   "0.0.1",
	})
	if err == nil {
		t.Fatal("expected error for invalid registry topic ID")
	}
}

func TestBuildHCS2UpdateTx(t *testing.T) {
	tx, err := BuildHCS2UpdateTx(UpdateTxParams{
		RegistryTopicID: "0.0.12345",
		UID:             "uid-123",
		TargetTopicID:   "0.0.54321",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2DeleteTx(t *testing.T) {
	tx, err := BuildHCS2DeleteTx(DeleteTxParams{
		RegistryTopicID: "0.0.12345",
		UID:             "uid-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2MigrateTx(t *testing.T) {
	tx, err := BuildHCS2MigrateTx(MigrateTxParams{
		RegistryTopicID: "0.0.12345",
		TargetTopicID:   "0.0.54321",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected non-nil transaction")
	}
}

func TestBuildHCS2RegisterTxWithAnalyticsMemo(t *testing.T) {
	tx, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "0.0.12345",
		TargetTopicID:   "0.0.54321",
		AnalyticsMemo:   "analytics-memo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.GetTransactionMemo() != "analytics-memo" {
		t.Fatalf("unexpected transaction memo: %s", tx.GetTransactionMemo())
	}
}

func TestNewClientHCS2(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected failure on invalid network")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if client.MirrorClient() == nil {
		t.Fatal("expected mirror client to be set")
	}
}

func TestCreateRegistryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	// Try with invalid topic ID or something to fail
	_, err := client.CreateRegistry(context.Background(), CreateRegistryOptions{
		RegistryType: RegistryTypeIndexed,
		TTL:          86400,
	})
	if err == nil {
		t.Fatal("expected execution failure")
	}
}

func TestRegisterEntryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})
	_, err := client.RegisterEntry(context.Background(), "invalid-topic-id", RegisterEntryOptions{}, "0.0.2")
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestUpdateEntryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})
	_, err := client.UpdateEntry(context.Background(), "invalid-topic-id", UpdateEntryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestDeleteEntryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})
	_, err := client.DeleteEntry(context.Background(), "invalid-topic-id", DeleteEntryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestMigrateRegistryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})
	_, err := client.MigrateRegistry(context.Background(), "invalid-topic-id", MigrateRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestGetRegistryFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.GetRegistry(context.Background(), "invalid-topic", QueryRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestGetTopicInfoFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.GetTopicInfo(context.Background(), "invalid-topic")
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestSubmitMessageFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.SubmitMessage(context.Background(), "0.0.1", Message{Op: "bad"}, "")
	// The bad operation will fail the validation step
	if err == nil {
		t.Fatal("expected fail from validation")
	}

	_, err = client.SubmitMessage(context.Background(), "0.0.1", Message{Op: OperationRegister, TopicID: "0.0.2", P: "hcs-2"}, "")
	// The payload is valid, but the user's mock testnet operator Account ID doesn't exist
	if err == nil {
		t.Fatal("expected fail from Hedera execution due to fake pk")
	}
}

func TestResolveRegistryFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.resolveRegistryType(context.Background(), "invalid-topic", nil)
	if err == nil {
		t.Fatal("expected get topic info fail in resolveRegistryType")
	}

	_, err = client.resolvePublicKey("invalid-topic", false)
	if err == nil {
		t.Fatal("expected get topic info fail in resolvePublicKey")
	}
}

func TestOverflowMessageSerialization(t *testing.T) {
	msg := OverflowMessage{
		P:             "hcs-2",
		Op:            OperationRegister,
		DataRef:       "hcs://1/0.0.99999",
		DataRefDigest: "abc123digest",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal overflow message: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, `"data_ref"`) {
		t.Fatal("expected data_ref field in serialized overflow message")
	}
	if !strings.Contains(s, `"data_ref_digest"`) {
		t.Fatal("expected data_ref_digest field in serialized overflow message")
	}
	if strings.Contains(s, `"t_id"`) {
		t.Fatal("overflow message should not contain t_id")
	}
}

func TestSubmitMessageOverflowTriggersInscription(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: pk.String(),
	})

	// Build a message with a huge metadata field that will exceed 1024 bytes
	bigMeta := strings.Repeat("x", 2000)
	msg := Message{
		P:        "hcs-2",
		Op:       OperationRegister,
		TopicID:  "0.0.12345",
		Metadata: bigMeta,
	}

	// This will fail at the inscriber auth step (testnet + fake operator), but the
	// error should indicate the overflow HCS-1 inscription path was attempted.
	_, err := client.submitMessage(context.Background(), "0.0.999999", msg, "")
	if err == nil {
		t.Fatal("expected error from inscription attempt")
	}
	// The error should come from the inscriber (auth or execution) since payload > 1024 bytes.
	if !strings.Contains(err.Error(), "overflow") && !strings.Contains(err.Error(), "inscrib") {
		t.Fatalf("expected overflow/inscriber error, got: %v", err)
	}
}
