package shared

import (
	"testing"
)

func TestParsePrivateKeyEdge(t *testing.T) {
	_, err := ParsePrivateKey("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}

	_, err = ParsePrivateKey("0xinvalidhex")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestLoadDotEnvIfPresent(t *testing.T) {
	// Just call it; it should not panic even without .env
	loadDotEnvIfPresent()
}
