package hcs21

import "testing"

func TestBuildRegistryMemo(t *testing.T) {
	memo, err := BuildRegistryMemo(3600, true, TopicTypeAdapterRegistry, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if memo != "hcs-21:0:3600:0" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestBuildDeclarationMessageTx(t *testing.T) {
	transaction, err := BuildDeclarationMessageTx("0.0.1001", AdapterDeclaration{
		P:         "hcs-21",
		Op:        OperationRegister,
		AdapterID: "adapter-1",
		Entity:    "service",
		Package: AdapterPackage{
			Registry:  "npm",
			Name:      "adapter",
			Version:   "1.0.0",
			Integrity: "sha384-abc",
		},
		Manifest: "hcs://1/0.0.123",
		Config: map[string]any{
			"type": "state",
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transaction == nil {
		t.Fatalf("expected transaction")
	}
}

