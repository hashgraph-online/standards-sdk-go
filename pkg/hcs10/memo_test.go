package hcs10

import "testing"

func TestBuildInboundMemo(t *testing.T) {
	memo := BuildInboundMemo(60, "0.0.123")
	if memo != "hcs-10:0:60:0:0.0.123" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseMemoType(t *testing.T) {
	topicType, ok := ParseMemoType("hcs-10:0:60:1")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if topicType != TopicTypeOutbound {
		t.Fatalf("unexpected topic type: %d", topicType)
	}
}

