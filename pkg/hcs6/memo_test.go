package hcs6

import "testing"

func TestBuildTopicMemo(t *testing.T) {
	memo := BuildTopicMemo(7200)
	if memo != "hcs-6:1:7200" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseTopicMemo(t *testing.T) {
	parsed, ok := ParseTopicMemo("hcs-6:1:86400")
	if !ok {
		t.Fatalf("expected memo parse success")
	}
	if parsed.RegistryType != RegistryTypeNonIndexed {
		t.Fatalf("unexpected registry type: %d", parsed.RegistryType)
	}
	if parsed.TTL != 86400 {
		t.Fatalf("unexpected ttl: %d", parsed.TTL)
	}
}
