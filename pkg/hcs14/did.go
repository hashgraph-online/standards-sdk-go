package hcs14

import (
	"crypto/sha512"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var uaidParamOrder = []string{
	"uid",
	"registry",
	"proto",
	"nativeId",
	"domain",
	"src",
	"version",
}

func CreateUAIDAID(
	input CanonicalAgentData,
	params RoutingParams,
	includeParams bool,
) (string, error) {
	normalized, canonicalJSON, err := canonicalizeAgentData(input)
	if err != nil {
		return "", err
	}

	sum := sha512.Sum384([]byte(canonicalJSON))
	identifier := base58Encode(sum[:])

	if !includeParams {
		return "uaid:aid:" + identifier, nil
	}

	defaulted := params
	if strings.TrimSpace(defaulted.Registry) == "" {
		defaulted.Registry = normalized.Registry
	}
	if strings.TrimSpace(defaulted.NativeID) == "" {
		defaulted.NativeID = normalized.NativeID
	}
	if strings.TrimSpace(defaulted.UID) == "" {
		defaulted.UID = "0"
	}

	return BuildCanonicalUAID("aid", identifier, routingParamsToMap(defaulted)), nil
}

func CreateUAIDFromDID(existingDID string, params RoutingParams) (string, error) {
	trimmed := strings.TrimSpace(existingDID)
	if trimmed == "" {
		return "", fmt.Errorf("existing DID is required")
	}

	var method string
	var idPart string
	if strings.HasPrefix(trimmed, "uaid:aid:") {
		method = "aid"
		idPart = strings.TrimPrefix(trimmed, "uaid:aid:")
	} else if strings.HasPrefix(trimmed, "did:") {
		parts := strings.SplitN(trimmed, ":", 3)
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid DID format")
		}
		method = parts[1]
		idPart = parts[2]
	} else {
		return "", fmt.Errorf("invalid DID format")
	}

	sanitized, hadSuffix := sanitizeDidSpecificID(idPart)
	finalID := sanitized

	if method == "hedera" {
		for _, networkPrefix := range []string{"mainnet:", "testnet:", "previewnet:", "devnet:"} {
			if strings.HasPrefix(finalID, networkPrefix) {
				finalID = strings.TrimPrefix(finalID, networkPrefix)
				break
			}
		}
	}

	paramMap := routingParamsToMap(params)
	if hadSuffix {
		if _, exists := paramMap["src"]; !exists {
			paramMap["src"] = "z" + base58Encode([]byte(trimmed))
		}
	}

	return BuildCanonicalUAID("did", finalID, paramMap), nil
}

func ParseUAID(value string) (ParsedUAID, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "uaid:") {
		return ParsedUAID{}, fmt.Errorf("invalid UAID")
	}

	remainder := strings.TrimPrefix(trimmed, "uaid:")
	target := ""
	switch {
	case strings.HasPrefix(remainder, "aid:"):
		target = "aid"
		remainder = strings.TrimPrefix(remainder, "aid:")
	case strings.HasPrefix(remainder, "did:"):
		target = "did"
		remainder = strings.TrimPrefix(remainder, "did:")
	default:
		return ParsedUAID{}, fmt.Errorf("invalid UAID target")
	}

	identifier := remainder
	paramSection := ""
	if separatorIndex := strings.Index(remainder, ";"); separatorIndex >= 0 {
		identifier = remainder[:separatorIndex]
		paramSection = remainder[separatorIndex+1:]
	}
	if strings.TrimSpace(identifier) == "" {
		return ParsedUAID{}, fmt.Errorf("UAID identifier is required")
	}

	params := parseSemicolonFields(paramSection)
	for key, value := range params {
		decoded, decodeErr := url.QueryUnescape(value)
		if decodeErr == nil {
			params[key] = decoded
		}
	}
	return ParsedUAID{
		Target: target,
		ID:     identifier,
		Params: params,
	}, nil
}

func BuildCanonicalUAID(target string, identifier string, params map[string]string) string {
	entries := make([]string, 0)
	usedKeys := map[string]struct{}{}

	for _, key := range uaidParamOrder {
		value := strings.TrimSpace(params[key])
		if value == "" {
			continue
		}
		encodedValue := url.QueryEscape(value)
		encodedValue = strings.ReplaceAll(encodedValue, "+", "%20")
		entries = append(entries, key+"="+encodedValue)
		usedKeys[key] = struct{}{}
	}

	extraKeys := make([]string, 0)
	for key, value := range params {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, exists := usedKeys[key]; exists {
			continue
		}
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)
	for _, key := range extraKeys {
		value := strings.TrimSpace(params[key])
		encodedValue := url.QueryEscape(value)
		encodedValue = strings.ReplaceAll(encodedValue, "+", "%20")
		entries = append(entries, key+"="+encodedValue)
	}

	if len(entries) == 0 {
		return "uaid:" + target + ":" + identifier
	}
	return "uaid:" + target + ":" + identifier + ";" + strings.Join(entries, ";")
}

func sanitizeDidSpecificID(idPart string) (string, bool) {
	index := strings.IndexAny(idPart, ";?#")
	if index == -1 {
		return idPart, false
	}
	return idPart[:index], true
}

func routingParamsToMap(params RoutingParams) map[string]string {
	output := map[string]string{}

	if value := strings.TrimSpace(params.UID); value != "" {
		output["uid"] = value
	}
	if value := strings.TrimSpace(params.Registry); value != "" {
		output["registry"] = value
	}
	if value := strings.TrimSpace(params.Proto); value != "" {
		output["proto"] = value
	}
	if value := strings.TrimSpace(params.NativeID); value != "" {
		output["nativeId"] = value
	}
	if value := strings.TrimSpace(params.Domain); value != "" {
		output["domain"] = value
	}
	if value := strings.TrimSpace(params.Src); value != "" {
		output["src"] = value
	}
	if value := strings.TrimSpace(params.Version); value != "" {
		output["version"] = value
	}

	return output
}
