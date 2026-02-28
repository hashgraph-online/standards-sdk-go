package hcs26

import "testing"

func TestBuildTopicMemo(t *testing.T) {
	memo := BuildTopicMemo(true, 3600, TopicTypeDiscovery)
	if memo != "hcs-26:0:3600:0" {
		t.Fatalf("unexpected topic memo: %s", memo)
	}
}

func TestParseTopicMemo(t *testing.T) {
	parsed, ok := ParseTopicMemo("hcs-26:1:86400:1")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if parsed.Indexed {
		t.Fatalf("expected non-indexed memo")
	}
	if parsed.TopicType != TopicTypeVersion {
		t.Fatalf("unexpected topic type: %d", parsed.TopicType)
	}
}

func TestBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo(OperationRegister, TopicTypeDiscovery)
	if memo != "hcs-26:op:0:0" {
		t.Fatalf("unexpected transaction memo: %s", memo)
	}
}
