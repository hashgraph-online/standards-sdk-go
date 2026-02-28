package hcs2

import "testing"

func TestBuildAndParseTopicMemo(t *testing.T) {
	memo := BuildTopicMemo(RegistryTypeIndexed, 86400)
	if memo != "hcs-2:0:86400" {
		t.Fatalf("unexpected topic memo: %s", memo)
	}

	parsed, ok := ParseTopicMemo(memo)
	if !ok {
		t.Fatalf("expected memo to parse")
	}
	if parsed.RegistryType != RegistryTypeIndexed {
		t.Fatalf("unexpected registry type: %d", parsed.RegistryType)
	}
	if parsed.TTL != 86400 {
		t.Fatalf("unexpected ttl: %d", parsed.TTL)
	}
}

func TestBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo(OperationDelete, RegistryTypeIndexed)
	if memo != "hcs-2:op:2:0" {
		t.Fatalf("unexpected transaction memo: %s", memo)
	}
}
