package hcs14

import (
	"context"
	"fmt"
	"strings"
)

type UAIDDidResolutionResolver struct{}

func NewUAIDDidResolutionResolver() *UAIDDidResolutionResolver {
	return &UAIDDidResolutionResolver{}
}

func (resolver *UAIDDidResolutionResolver) ProfileID() string {
	return UAIDDidResolutionProfileID
}

func (resolver *UAIDDidResolutionResolver) Supports(_ string, parsed ParsedUAID) bool {
	return parsed.Target == "did"
}

func (resolver *UAIDDidResolutionResolver) ResolveProfile(
	ctx context.Context,
	uaid string,
	context UAIDProfileResolverContext,
) (*UAIDResolutionResult, error) {
	parsed := context.ParsedUAID
	if parsed.Target != "did" {
		return uaidDidProfileError(
			uaid,
			"ERR_INVALID_UAID",
			"identifier is not uaid:did and cannot be resolved by this profile",
			nil,
		), nil
	}

	baseDID := strings.TrimSpace(context.DID)
	if src := strings.TrimSpace(parsed.Params["src"]); src != "" {
		decodedDID, decodeErr := decodeSrcDID(src)
		if decodeErr == nil {
			baseDID = decodedDID
		}
	}
	if baseDID == "" {
		return uaidDidProfileError(
			uaid,
			"ERR_BASE_DID_UNDETERMINED",
			"unable to determine base DID; provide src parameter or resolvable method mapping",
			map[string]any{"uaid": uaid},
		), nil
	}

	var didDocument *DIDDocument
	if context.DIDDocument != nil && context.DIDDocument.ID == baseDID {
		didDocument = context.DIDDocument
	} else if context.ResolveDID != nil {
		resolvedDocument, resolveErr := context.ResolveDID(ctx, baseDID)
		if resolveErr != nil {
			return nil, resolveErr
		}
		didDocument = resolvedDocument
	}

	if didDocument == nil {
		return uaidDidProfileError(
			uaid,
			"ERR_DID_RESOLUTION_FAILED",
			"base DID resolution failed",
			map[string]any{"uaid": uaid, "baseDid": baseDID},
		), nil
	}
	if strings.TrimSpace(didDocument.ID) == "" {
		return uaidDidProfileError(
			uaid,
			"ERR_DID_DOCUMENT_INVALID",
			"resolved DID document is malformed",
			map[string]any{"uaid": uaid, "baseDid": baseDID},
		), nil
	}

	services := cloneServices(didDocument.Service)
	includeHintedService := len(services) == 0
	if includeHintedService {
		if hintedService := buildHintedService(uaid, parsed.Params); hintedService != nil {
			services = append(services, *hintedService)
		}
	}

	return &UAIDResolutionResult{
		ID:                 uaid,
		DID:                baseDID,
		AlsoKnownAs:        dedupeStrings(append([]string{baseDID}, didDocument.AlsoKnownAs...)),
		VerificationMethod: cloneVerificationMethods(didDocument.VerificationMethod),
		Authentication:     append([]string{}, didDocument.Authentication...),
		AssertionMethod:    append([]string{}, didDocument.AssertionMethod...),
		Service:            services,
		Metadata: UAIDMetadata{
			Profile:            UAIDDidResolutionProfileID,
			Resolved:           true,
			BaseDID:            baseDID,
			BaseDIDResolved:    true,
			VerificationMethod: "did-resolution",
			ResolutionMode:     "full-resolution",
		},
	}, nil
}

func uaidDidProfileError(
	uaid string,
	code string,
	message string,
	details map[string]any,
) *UAIDResolutionResult {
	return &UAIDResolutionResult{
		ID: uaid,
		Metadata: UAIDMetadata{
			Profile:  UAIDDidResolutionProfileID,
			Resolved: false,
		},
		Error: &UAIDResolutionError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func decodeSrcDID(src string) (string, error) {
	decoded, err := decodeMultibaseB58btc(src)
	if err != nil {
		return "", err
	}
	did := strings.TrimSpace(string(decoded))
	if did == "" {
		return "", fmt.Errorf("decoded DID is empty")
	}
	return did, nil
}

func buildHintedService(uaid string, params map[string]string) *ServiceEndpoint {
	proto := strings.TrimSpace(params["proto"])
	nativeID := strings.TrimSpace(params["nativeId"])
	domain := strings.TrimSpace(params["domain"])
	if proto == "" && nativeID == "" && domain == "" {
		return nil
	}

	parts := make([]string, 0)
	if proto != "" {
		parts = append(parts, "proto="+proto)
	}
	if nativeID != "" {
		parts = append(parts, "nativeId="+nativeID)
	}
	if domain != "" {
		parts = append(parts, "domain="+domain)
	}

	return &ServiceEndpoint{
		ID:              uaid + "#hcs14-hinted-service-1",
		Type:            "HintedService",
		ServiceEndpoint: strings.Join(parts, ";"),
		Source:          "uaid-parameters",
	}
}

func cloneServices(input []ServiceEndpoint) []ServiceEndpoint {
	if len(input) == 0 {
		return nil
	}
	output := make([]ServiceEndpoint, len(input))
	copy(output, input)
	return output
}

func cloneVerificationMethods(input []DIDVerificationMethod) []DIDVerificationMethod {
	if len(input) == 0 {
		return nil
	}
	output := make([]DIDVerificationMethod, len(input))
	copy(output, input)
	return output
}

func dedupeStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	output := make([]string, 0, len(input))
	for _, value := range input {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		output = append(output, trimmed)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}
