package hcs12

import "testing"

func TestBuildRegistryMemoPrimary(t *testing.T) {
	memo, err := BuildRegistryMemo(RegistryTypeAction, 3600)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if memo != "hcs-12:1:3600:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseRegistryMemoPrimary(t *testing.T) {
	registryType, ttl, ok := ParseRegistryMemo("hcs-12:1:7200:2")
	if !ok {
		t.Fatalf("expected parse success")
	}
	if registryType != RegistryTypeAssembly {
		t.Fatalf("unexpected registry type: %s", registryType)
	}
	if ttl != 7200 {
		t.Fatalf("unexpected ttl: %d", ttl)
	}
}
