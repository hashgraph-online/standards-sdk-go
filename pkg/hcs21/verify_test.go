package hcs21

import (
	"crypto/sha384"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestCanonicalize(t *testing.T) {
	value := map[string]any{
		"b": "two",
		"a": "one",
	}
	canonical := Canonicalize(value)
	if canonical != "{\"a\":\"one\",\"b\":\"two\"}" {
		t.Fatalf("unexpected canonical form: %s", canonical)
	}
}

func TestVerifyArtifactDigest(t *testing.T) {
	payload := []byte("hello")
	sum := sha384.Sum384(payload)
	hexDigest := hex.EncodeToString(sum[:])
	base64Digest := base64.StdEncoding.EncodeToString(sum[:])

	if !VerifyArtifactDigest(payload, "sha384:"+hexDigest) {
		t.Fatalf("expected hex digest verification success")
	}
	if !VerifyArtifactDigest(payload, "sha384-"+base64Digest) {
		t.Fatalf("expected base64 digest verification success")
	}
}

