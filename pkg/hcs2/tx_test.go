package hcs2

import (
	"encoding/json"
	"testing"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestBuildHCS2CreateRegistryTx(t *testing.T) {
	transaction := BuildHCS2CreateRegistryTx(CreateRegistryTxParams{
		RegistryType: RegistryTypeIndexed,
		TTL:          7200,
		MemoOverride: "",
	})

	if transaction.GetTopicMemo() != "hcs-2:0:7200" {
		t.Fatalf("unexpected topic memo: %s", transaction.GetTopicMemo())
	}
}

func TestBuildHCS2CreateRegistryTxWithMemoOverrideAndKeys(t *testing.T) {
	privateKey, err := hedera.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	transaction := BuildHCS2CreateRegistryTx(CreateRegistryTxParams{
		RegistryType: RegistryTypeIndexed,
		TTL:          7200,
		AdminKey:     privateKey.PublicKey(),
		SubmitKey:    privateKey.PublicKey(),
		MemoOverride: "hcs-2:custom",
	})

	if transaction.GetTopicMemo() != "hcs-2:custom" {
		t.Fatalf("unexpected topic memo: %s", transaction.GetTopicMemo())
	}

	adminKey, adminErr := transaction.GetAdminKey()
	if adminErr != nil || adminKey == nil {
		t.Fatalf("expected admin key to be set: %v", adminErr)
	}
	submitKey, submitErr := transaction.GetSubmitKey()
	if submitErr != nil || submitKey == nil {
		t.Fatalf("expected submit key to be set: %v", submitErr)
	}
}

func TestBuildHCS2RegisterTx(t *testing.T) {
	transaction, err := BuildHCS2RegisterTx(RegisterTxParams{
		RegistryTopicID: "0.0.100",
		TargetTopicID:   "0.0.200",
		Metadata:        "hcs://1/0.0.200",
		Memo:            "register",
		AnalyticsMemo:   "hcs-2:op:0:0",
	})
	if err != nil {
		t.Fatalf("BuildHCS2RegisterTx failed: %v", err)
	}

	if transaction.GetTopicID().String() != "0.0.100" {
		t.Fatalf("unexpected topic ID: %s", transaction.GetTopicID().String())
	}
	if transaction.GetTransactionMemo() != "hcs-2:op:0:0" {
		t.Fatalf("unexpected transaction memo: %s", transaction.GetTransactionMemo())
	}

	message := decodeMessageFromTx(t, transaction)
	if message.Op != OperationRegister {
		t.Fatalf("unexpected operation: %s", message.Op)
	}
	if message.TopicID != "0.0.200" {
		t.Fatalf("unexpected t_id: %s", message.TopicID)
	}
}

func TestBuildHCS2UpdateDeleteMigrateTx(t *testing.T) {
	updateTx, updateErr := BuildHCS2UpdateTx(UpdateTxParams{
		RegistryTopicID: "0.0.100",
		UID:             "2",
		TargetTopicID:   "0.0.201",
		Metadata:        "hcs://1/0.0.201",
		Memo:            "update",
		AnalyticsMemo:   "hcs-2:op:1:0",
	})
	if updateErr != nil {
		t.Fatalf("BuildHCS2UpdateTx failed: %v", updateErr)
	}
	updateMessage := decodeMessageFromTx(t, updateTx)
	if updateMessage.Op != OperationUpdate || updateMessage.UID != "2" || updateMessage.TopicID != "0.0.201" {
		t.Fatalf("unexpected update message: %+v", updateMessage)
	}

	deleteTx, deleteErr := BuildHCS2DeleteTx(DeleteTxParams{
		RegistryTopicID: "0.0.100",
		UID:             "2",
		Memo:            "delete",
		AnalyticsMemo:   "hcs-2:op:2:0",
	})
	if deleteErr != nil {
		t.Fatalf("BuildHCS2DeleteTx failed: %v", deleteErr)
	}
	deleteMessage := decodeMessageFromTx(t, deleteTx)
	if deleteMessage.Op != OperationDelete || deleteMessage.UID != "2" {
		t.Fatalf("unexpected delete message: %+v", deleteMessage)
	}

	migrateTx, migrateErr := BuildHCS2MigrateTx(MigrateTxParams{
		RegistryTopicID: "0.0.100",
		TargetTopicID:   "0.0.300",
		Metadata:        "hcs://1/0.0.300",
		Memo:            "migrate",
		AnalyticsMemo:   "hcs-2:op:3:0",
	})
	if migrateErr != nil {
		t.Fatalf("BuildHCS2MigrateTx failed: %v", migrateErr)
	}
	migrateMessage := decodeMessageFromTx(t, migrateTx)
	if migrateMessage.Op != OperationMigrate || migrateMessage.TopicID != "0.0.300" {
		t.Fatalf("unexpected migrate message: %+v", migrateMessage)
	}
}

func decodeMessageFromTx(t *testing.T, transaction *hedera.TopicMessageSubmitTransaction) Message {
	t.Helper()

	var message Message
	if err := json.Unmarshal(transaction.GetMessage(), &message); err != nil {
		t.Fatalf("failed to decode tx message payload: %v", err)
	}
	return message
}
