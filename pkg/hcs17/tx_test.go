package hcs17

import (
	"encoding/json"
	"testing"
)

func TestBuildCreateStateTopicTx(t *testing.T) {
	transaction := BuildCreateStateTopicTx(CreateTopicOptions{
		TTLSeconds: 3600,
	})
	if transaction.GetTopicMemo() != "hcs-17:0:3600" {
		t.Fatalf("unexpected topic memo: %s", transaction.GetTopicMemo())
	}
}

func TestBuildStateHashMessageTx(t *testing.T) {
	message := StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "0xabc",
		Topics:    []string{"0.0.1"},
		AccountID: "0.0.2",
	}
	transaction, err := BuildStateHashMessageTx("0.0.100", message, "")
	if err != nil {
		t.Fatalf("BuildStateHashMessageTx failed: %v", err)
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
		t.Fatalf("unexpected state hash: %+v", payload["state_hash"])
	}
}
