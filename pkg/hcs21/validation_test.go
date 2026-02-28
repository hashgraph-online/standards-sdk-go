package hcs21

import "testing"

func TestValidateDeclaration(t *testing.T) {
	err := ValidateDeclaration(AdapterDeclaration{
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
	})
	if err != nil {
		t.Fatalf("expected valid declaration, got %v", err)
	}
}

func TestValidateDeclarationMissingConfig(t *testing.T) {
	err := ValidateDeclaration(AdapterDeclaration{
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
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
