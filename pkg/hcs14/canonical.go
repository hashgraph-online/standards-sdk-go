package hcs14

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func canonicalizeAgentData(input CanonicalAgentData) (CanonicalAgentData, string, error) {
	normalized := CanonicalAgentData{
		Registry: strings.ToLower(strings.TrimSpace(input.Registry)),
		Name:     strings.TrimSpace(input.Name),
		Version:  strings.TrimSpace(input.Version),
		Protocol: strings.ToLower(strings.TrimSpace(input.Protocol)),
		NativeID: strings.TrimSpace(input.NativeID),
		Skills:   slices.Clone(input.Skills),
	}

	if normalized.Registry == "" {
		return CanonicalAgentData{}, "", fmt.Errorf("registry is required")
	}
	if normalized.Name == "" {
		return CanonicalAgentData{}, "", fmt.Errorf("name is required")
	}
	if normalized.Version == "" {
		return CanonicalAgentData{}, "", fmt.Errorf("version is required")
	}
	if normalized.Protocol == "" {
		return CanonicalAgentData{}, "", fmt.Errorf("protocol is required")
	}
	if normalized.NativeID == "" {
		return CanonicalAgentData{}, "", fmt.Errorf("nativeId is required")
	}
	if normalized.Protocol == "hcs-10" && !isHederaCAIP10(normalized.NativeID) {
		return CanonicalAgentData{}, "", fmt.Errorf("for protocol hcs-10, nativeId must be Hedera CAIP-10")
	}
	if normalized.Protocol == "acp-virtuals" && !isEIP155CAIP10(normalized.NativeID) {
		return CanonicalAgentData{}, "", fmt.Errorf("for protocol acp-virtuals, nativeId must be EIP-155 CAIP-10")
	}

	slices.Sort(normalized.Skills)

	canonicalObject := map[string]any{
		"skills":   normalized.Skills,
		"name":     normalized.Name,
		"nativeId": normalized.NativeID,
		"protocol": normalized.Protocol,
		"registry": normalized.Registry,
		"version":  normalized.Version,
	}
	canonicalJSONBytes, err := json.Marshal(canonicalObject)
	if err != nil {
		return CanonicalAgentData{}, "", err
	}

	return normalized, string(canonicalJSONBytes), nil
}
