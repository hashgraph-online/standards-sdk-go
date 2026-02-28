package hcs18

import "testing"

func TestBuildDiscoveryMemo(t *testing.T) {
	memo := BuildDiscoveryMemo(7200, "")
	if memo != "hcs-18:0:7200" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseDiscoveryMemo(t *testing.T) {
	parsed, ok := ParseDiscoveryMemo("hcs-18:0:86400")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if parsed.TTL != 86400 {
		t.Fatalf("unexpected ttl: %d", parsed.TTL)
	}
}
