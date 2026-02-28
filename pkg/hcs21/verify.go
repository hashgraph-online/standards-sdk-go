package hcs21

import (
	"crypto/sha384"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type jsonValue = any

func sortJSON(value jsonValue) jsonValue {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		sorted := make(map[string]any, len(keys))
		for _, key := range keys {
			sorted[key] = sortJSON(typed[key])
		}
		return sorted
	case []any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, sortJSON(item))
		}
		return items
	default:
		return value
	}
}

// Canonicalize performs the requested operation.
func Canonicalize(value any) string {
	sorted := sortJSON(value)
	encoded, err := json.Marshal(sorted)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// VerifyDeclarationSignature performs the requested operation.
func VerifyDeclarationSignature(declaration AdapterDeclaration, publisherPublicKey string) bool {
	if strings.TrimSpace(declaration.Signature) == "" {
		return false
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(declaration.Signature)
	if err != nil {
		return false
	}
	publicKey, err := hedera.PublicKeyFromString(publisherPublicKey)
	if err != nil {
		return false
	}

	unsigned := map[string]any{
		"p":         declaration.P,
		"op":        declaration.Op,
		"adapter_id": declaration.AdapterID,
		"entity":    declaration.Entity,
		"package":   declaration.Package,
		"manifest":  declaration.Manifest,
		"config":    declaration.Config,
	}
	if declaration.ManifestSequence > 0 {
		unsigned["manifest_sequence"] = declaration.ManifestSequence
	}
	if strings.TrimSpace(declaration.StateModel) != "" {
		unsigned["state_model"] = declaration.StateModel
	}

	payload := Canonicalize(unsigned)
	if payload == "" {
		return false
	}

	return publicKey.Verify([]byte(payload), signatureBytes)
}

// VerifyManifestSignature performs the requested operation.
func VerifyManifestSignature(manifest any, signatureBase64 string, publisherPublicKey string) bool {
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}
	publicKey, err := hedera.PublicKeyFromString(publisherPublicKey)
	if err != nil {
		return false
	}

	payload := Canonicalize(manifest)
	if payload == "" {
		return false
	}
	return publicKey.Verify([]byte(payload), signatureBytes)
}

func normalizeDigest(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(trimmed), "sha384-"), "sha384:")
	return trimmed
}

// VerifyArtifactDigest performs the requested operation.
func VerifyArtifactDigest(artifact []byte, expectedDigest string) bool {
	sum := sha384.Sum384(artifact)
	hexDigest := strings.ToLower(hex.EncodeToString(sum[:]))
	base64Digest := strings.ToLower(base64.StdEncoding.EncodeToString(sum[:]))
	expected := normalizeDigest(expectedDigest)
	return expected == hexDigest || expected == base64Digest
}

