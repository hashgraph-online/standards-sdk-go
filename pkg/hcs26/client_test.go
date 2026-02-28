package hcs26

import "testing"

func TestNewClientInvalidNetwork(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Network: "invalid",
	})
	if err == nil {
		t.Fatalf("expected network validation error")
	}
}

func TestParseDiscoveryRegisterNewShape(t *testing.T) {
	register, ok, err := parseDiscoveryRegister(map[string]any{
		"p":          "hcs-26",
		"op":         "register",
		"t_id":       "0.0.1001",
		"account_id": "0.0.2002",
		"metadata": map[string]any{
			"name":        "skill",
			"description": "desc",
			"author":      "author",
			"license":     "Apache-2.0",
		},
	}, 42)
	if err != nil || !ok {
		t.Fatalf("expected parse success, err=%v", err)
	}
	if register.VersionRegistry != "0.0.1001" {
		t.Fatalf("unexpected version registry: %s", register.VersionRegistry)
	}
	if register.SequenceNumber != 42 {
		t.Fatalf("unexpected sequence number: %d", register.SequenceNumber)
	}
}

func TestValidateManifest(t *testing.T) {
	err := validateManifest(SkillManifest{
		Name:        "skill",
		Description: "desc",
		Version:     "1.0.0",
		License:     "Apache-2.0",
		Author:      "author",
		Files: []ManifestFile{
			{
				Path:   "SKILL.md",
				HRL:    "hcs://1/0.0.123",
				SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Mime:   "text/markdown",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected valid manifest, got %v", err)
	}
}
