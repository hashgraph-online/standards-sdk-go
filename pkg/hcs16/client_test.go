package hcs16

import (
	"testing"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestParseTopicMemo(t *testing.T) {
	client := &Client{}

	parsed := client.ParseTopicMemo("hcs-16:0.0.12345:2")
	if parsed == nil {
		t.Fatalf("expected topic memo to parse")
	}
	if parsed.Protocol != "hcs-16" {
		t.Fatalf("unexpected protocol: %s", parsed.Protocol)
	}
	if parsed.FloraAccountID != "0.0.12345" {
		t.Fatalf("unexpected flora account ID: %s", parsed.FloraAccountID)
	}
	if parsed.TopicType != FloraTopicTypeState {
		t.Fatalf("unexpected topic type: %d", parsed.TopicType)
	}

	if client.ParseTopicMemo("hcs-16:invalid") != nil {
		t.Fatalf("expected invalid memo to return nil")
	}
}

func TestBuildFloraTopicCreateTxs(t *testing.T) {
	keyList := hedera.KeyListWithThreshold(1)
	submitList := hedera.KeyListWithThreshold(1)
	client := &Client{}

	transactions, err := client.BuildFloraTopicCreateTxs("0.0.100", keyList, submitList, "")
	if err != nil {
		t.Fatalf("BuildFloraTopicCreateTxs failed: %v", err)
	}
	if len(transactions) != 3 {
		t.Fatalf("expected 3 topic create transactions, got %d", len(transactions))
	}
}

func TestExtractMirrorKeyCandidate(t *testing.T) {
	raw := map[string]any{
		"key": map[string]any{
			"ECDSA_secp256k1": "302e020100300506032b657004220420abc",
		},
	}
	key := extractMirrorKeyString(raw)
	if key != "302e020100300506032b657004220420abc" {
		t.Fatalf("unexpected mirror key extraction result: %s", key)
	}
}
