package hcs16

import (
	"encoding/json"
	"testing"
)

func TestBuildFloraCreatedTx(t *testing.T) {
	transaction, err := BuildFloraCreatedTx(
		"0.0.100",
		"0.0.operator@0.0.flora",
		"0.0.flora",
		FloraTopics{
			Communication: "0.0.101",
			Transaction:   "0.0.102",
			State:         "0.0.103",
		},
	)
	if err != nil {
		t.Fatalf("BuildFloraCreatedTx failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(transaction.GetMessage(), &payload); err != nil {
		t.Fatalf("failed to decode message payload: %v", err)
	}
	if payload["p"] != "hcs-16" {
		t.Fatalf("unexpected protocol: %+v", payload["p"])
	}
	if payload["op"] != string(FloraOperationFloraCreated) {
		t.Fatalf("unexpected operation: %+v", payload["op"])
	}
}

func TestBuildTransactionTx(t *testing.T) {
	transaction, err := BuildTransactionTx(
		"0.0.100",
		"0.0.operator@0.0.flora",
		"0.0.200",
		"desc",
	)
	if err != nil {
		t.Fatalf("BuildTransactionTx failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(transaction.GetMessage(), &payload); err != nil {
		t.Fatalf("failed to decode message payload: %v", err)
	}
	if payload["op"] != string(FloraOperationTransaction) {
		t.Fatalf("unexpected operation: %+v", payload["op"])
	}
	if payload["schedule_id"] != "0.0.200" {
		t.Fatalf("unexpected schedule_id: %+v", payload["schedule_id"])
	}
}

func TestBuildStateUpdateTx(t *testing.T) {
	epoch := int64(42)
	transaction, err := BuildStateUpdateTx(
		"0.0.100",
		"0.0.operator@0.0.flora",
		"0xabc",
		&epoch,
		"",
		[]string{"0.0.500"},
		"memo",
		"",
	)
	if err != nil {
		t.Fatalf("BuildStateUpdateTx failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(transaction.GetMessage(), &payload); err != nil {
		t.Fatalf("failed to decode message payload: %v", err)
	}
	if payload["p"] != "hcs-17" {
		t.Fatalf("unexpected protocol: %+v", payload["p"])
	}
	if payload["op"] != "state_hash" {
		t.Fatalf("unexpected operation: %+v", payload["op"])
	}
	if payload["state_hash"] != "0xabc" {
		t.Fatalf("unexpected state_hash: %+v", payload["state_hash"])
	}
}

func TestBuildCreateFloraTopicTx(t *testing.T) {
	transaction, err := BuildCreateFloraTopicTx(CreateFloraTopicOptions{
		FloraAccountID: "0.0.12345",
		TopicType:      FloraTopicTypeState,
	})
	if err != nil {
		t.Fatalf("BuildCreateFloraTopicTx failed: %v", err)
	}
	if transaction.GetTopicMemo() != "hcs-16:0.0.12345:2" {
		t.Fatalf("unexpected topic memo: %s", transaction.GetTopicMemo())
	}
	if transaction.GetTransactionMemo() != "hcs-16:op:0:2" {
		t.Fatalf("unexpected transaction memo: %s", transaction.GetTransactionMemo())
	}
}
