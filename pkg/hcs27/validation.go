package hcs27

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type HCS1ResolverFunc func(ctx context.Context, hcs1Reference string) ([]byte, error)

func ValidateCheckpointMessage(
	ctx context.Context,
	message CheckpointMessage,
	resolver HCS1ResolverFunc,
) (CheckpointMetadata, error) {
	var metadata CheckpointMetadata

	if message.Protocol != ProtocolID {
		return metadata, fmt.Errorf("invalid protocol %q", message.Protocol)
	}
	if message.Operation != OperationName {
		return metadata, fmt.Errorf("invalid operation %q", message.Operation)
	}
	if len(message.Memo) >= 300 {
		return metadata, fmt.Errorf("message memo must be less than 300 characters")
	}
	if len(message.Metadata) == 0 {
		return metadata, fmt.Errorf("metadata is required")
	}

	trimmed := strings.TrimSpace(string(message.Metadata))
	if trimmed == "" {
		return metadata, fmt.Errorf("metadata is required")
	}

	var resolvedPayload []byte
	if strings.HasPrefix(trimmed, "\"") {
		var reference string
		if err := json.Unmarshal(message.Metadata, &reference); err != nil {
			return metadata, fmt.Errorf("metadata reference must be a JSON string: %w", err)
		}
		if !strings.HasPrefix(reference, "hcs://1/") {
			return metadata, fmt.Errorf("metadata reference must be an hcs://1 URI")
		}
		if resolver == nil {
			return metadata, fmt.Errorf("metadata reference provided but no HCS-1 resolver configured")
		}

		payload, err := resolver(ctx, reference)
		if err != nil {
			return metadata, fmt.Errorf("failed to resolve metadata reference %q: %w", reference, err)
		}
		resolvedPayload = payload

		if err := json.Unmarshal(payload, &metadata); err != nil {
			return metadata, fmt.Errorf("resolved metadata is not valid JSON: %w", err)
		}
	} else {
		if err := json.Unmarshal(message.Metadata, &metadata); err != nil {
			return metadata, fmt.Errorf("metadata object is invalid JSON: %w", err)
		}
	}

	if err := validateMetadata(metadata); err != nil {
		return metadata, err
	}

	if message.MetadataDigest != nil {
		if strings.TrimSpace(message.MetadataDigest.Algorithm) != "sha-256" {
			return metadata, fmt.Errorf("metadata_digest.alg must be sha-256")
		}
		if len(resolvedPayload) == 0 {
			return metadata, fmt.Errorf("metadata_digest requires metadata reference resolution")
		}

		sum := sha256.Sum256(resolvedPayload)
		expected := base64.RawURLEncoding.EncodeToString(sum[:])
		if expected != strings.TrimSpace(message.MetadataDigest.DigestB64) {
			return metadata, fmt.Errorf("metadata digest does not match resolved payload")
		}
	}

	return metadata, nil
}

func validateMetadata(metadata CheckpointMetadata) error {
	if strings.TrimSpace(metadata.Type) != "ans-checkpoint-v1" {
		return fmt.Errorf("metadata.type must be ans-checkpoint-v1")
	}
	if strings.TrimSpace(metadata.Stream.Registry) == "" {
		return fmt.Errorf("metadata.stream.registry is required")
	}
	if strings.TrimSpace(metadata.Stream.LogID) == "" {
		return fmt.Errorf("metadata.stream.log_id is required")
	}
	if metadata.Log == nil {
		return fmt.Errorf("metadata.log is required")
	}
	if strings.TrimSpace(metadata.Log.Algorithm) != "sha-256" {
		return fmt.Errorf("metadata.log.alg must be sha-256")
	}
	if strings.TrimSpace(metadata.Log.Leaf) == "" {
		return fmt.Errorf("metadata.log.leaf is required")
	}
	if strings.TrimSpace(metadata.Log.Merkle) == "" {
		return fmt.Errorf("metadata.log.merkle is required")
	}
	if metadata.Root.TreeSize == 0 && metadata.BatchRange.End > 0 {
		return fmt.Errorf("metadata.root.treeSize must be >= batch_range.end")
	}
	if _, err := base64.RawURLEncoding.DecodeString(metadata.Root.RootHashB64); err != nil {
		return fmt.Errorf("metadata.root.rootHashB64u must be base64url: %w", err)
	}
	if metadata.Previous != nil {
		if _, err := base64.RawURLEncoding.DecodeString(metadata.Previous.RootHashB64); err != nil {
			return fmt.Errorf("metadata.prev.rootHashB64u must be base64url: %w", err)
		}
		if metadata.Previous.TreeSize > metadata.Root.TreeSize {
			return fmt.Errorf("metadata.prev.treeSize must be <= metadata.root.treeSize")
		}
	}

	if metadata.BatchRange.End < metadata.BatchRange.Start {
		return fmt.Errorf("metadata.batch_range.end must be >= start")
	}
	if metadata.BatchRange.End > metadata.Root.TreeSize {
		return fmt.Errorf("metadata.batch_range.end must be <= metadata.root.treeSize")
	}

	if metadata.Signature != nil {
		if strings.TrimSpace(metadata.Signature.Algorithm) == "" {
			return fmt.Errorf("metadata.sig.alg is required when metadata.sig is present")
		}
		if strings.TrimSpace(metadata.Signature.KeyID) == "" {
			return fmt.Errorf("metadata.sig.kid is required when metadata.sig is present")
		}
		if strings.TrimSpace(metadata.Signature.Signature) == "" {
			return fmt.Errorf("metadata.sig.b64u is required when metadata.sig is present")
		}
		if _, err := base64.RawURLEncoding.DecodeString(metadata.Signature.Signature); err != nil {
			return fmt.Errorf("metadata.sig.b64u must be base64url: %w", err)
		}
	}

	return nil
}
