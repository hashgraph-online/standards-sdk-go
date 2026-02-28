package hcs27

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestEmptyRootVector(t *testing.T) {
	rootHex := hex.EncodeToString(EmptyRoot())
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if rootHex != expected {
		t.Fatalf("unexpected empty root: %s", rootHex)
	}
}

func TestSingleEntryLeafVector(t *testing.T) {
	entry := map[string]any{
		"event":     "register",
		"issued_at": "2026-01-01T00:00:00Z",
		"log_id":    "default",
		"payload": map[string]any{
			"hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			"uri":  "hcs://1/0.0.123",
		},
		"record_id": "registry-native-id",
		"registry":  "example",
	}

	leafHex, err := LeafHashHexFromEntry(entry)
	if err != nil {
		t.Fatalf("failed to hash entry: %v", err)
	}

	expected := "a12882925d08570166fe748ebdc16670fc0c69428e2b60ed388b35b52c91d6e2"
	if leafHex != expected {
		t.Fatalf("unexpected leaf hash: %s", leafHex)
	}

	ok, err := VerifyInclusionProof(
		0,
		1,
		leafHex,
		[]string{},
		base64.StdEncoding.EncodeToString(mustHex(leafHex)),
	)
	if err != nil {
		t.Fatalf("inclusion verification returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected inclusion proof to verify for single-entry tree")
	}
}

func TestConsistencyProof_EmptyToAny(t *testing.T) {
	ok, err := VerifyConsistencyProof(
		0,
		10,
		"",
		"ignored",
		nil,
	)
	if err != nil {
		t.Fatalf("consistency verification returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected consistency proof to verify for oldTreeSize=0")
	}
}

func mustHex(value string) []byte {
	decoded, _ := hex.DecodeString(value)
	return decoded
}
