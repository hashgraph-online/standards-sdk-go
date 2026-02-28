package hcs2

import (
	"strings"
	"testing"
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
