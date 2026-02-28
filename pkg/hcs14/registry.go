package hcs14

import (
	"context"
	"strings"
)

type ResolveUaidProfileOptions struct {
	ProfileID string
}

type ResolverRegistry struct {
	didResolvers         []DIDResolver
	didProfileResolvers  []DIDProfileResolver
	uaidProfileResolvers []UAIDProfileResolver
}

func NewResolverRegistry() *ResolverRegistry {
	return &ResolverRegistry{
		didResolvers:         make([]DIDResolver, 0),
		didProfileResolvers:  make([]DIDProfileResolver, 0),
		uaidProfileResolvers: make([]UAIDProfileResolver, 0),
	}
}

func (registry *ResolverRegistry) RegisterDIDResolver(resolver DIDResolver) {
	registry.didResolvers = append(registry.didResolvers, resolver)
}

func (registry *ResolverRegistry) RegisterDIDProfileResolver(resolver DIDProfileResolver) {
	registry.didProfileResolvers = append(registry.didProfileResolvers, resolver)
}

func (registry *ResolverRegistry) RegisterUAIDProfileResolver(resolver UAIDProfileResolver) {
	registry.uaidProfileResolvers = append(registry.uaidProfileResolvers, resolver)
}

func (registry *ResolverRegistry) ResolveDID(ctx context.Context, did string) (*DIDDocument, error) {
	for _, resolver := range registry.didResolvers {
		if !resolver.Supports(did) {
			continue
		}
		document, err := resolver.Resolve(ctx, did)
		if err != nil {
			return nil, err
		}
		if document != nil {
			return document, nil
		}
	}
	return nil, nil
}

func (registry *ResolverRegistry) ResolveUAIDProfile(
	ctx context.Context,
	uaid string,
	options ResolveUaidProfileOptions,
) (*UAIDResolutionResult, error) {
	parsed, err := ParseUAID(uaid)
	if err != nil {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_INVALID_UAID",
				Message: err.Error(),
			},
		}, nil
	}

	profile, resolveErr := registry.resolveUaidProfileInternal(ctx, uaid, parsed, options.ProfileID, "")
	if resolveErr != nil {
		return nil, resolveErr
	}
	if profile != nil {
		return profile, nil
	}

	if options.ProfileID != "" {
		return nil, nil
	}

	derivedDID := deriveDIDFromParsedUAID(parsed)
	if derivedDID == "" {
		if parsed.Target == "aid" {
			return &UAIDResolutionResult{ID: uaid}, nil
		}
		return nil, nil
	}

	didDocument, didResolveErr := registry.ResolveDID(ctx, derivedDID)
	if didResolveErr != nil {
		return nil, didResolveErr
	}
	return registry.buildFallbackProfile(uaid, derivedDID, didDocument), nil
}

func (registry *ResolverRegistry) resolveUaidProfileInternal(
	ctx context.Context,
	uaid string,
	parsed ParsedUAID,
	profileID string,
	excludeProfileID string,
) (*UAIDResolutionResult, error) {
	derivedDID := deriveDIDFromParsedUAID(parsed)
	var didDocument *DIDDocument
	var err error
	if derivedDID != "" {
		didDocument, err = registry.ResolveDID(ctx, derivedDID)
		if err != nil {
			return nil, err
		}
	}

	fallback := registry.buildFallbackProfile(uaid, derivedDID, didDocument)
	resolvers := registry.uaidProfileResolvers
	if profileID != "" {
		resolvers = make([]UAIDProfileResolver, 0)
		for _, resolver := range registry.uaidProfileResolvers {
			if resolver.ProfileID() == profileID {
				resolvers = append(resolvers, resolver)
			}
		}
	}

	for _, resolver := range resolvers {
		if excludeProfileID != "" && resolver.ProfileID() == excludeProfileID {
			continue
		}
		if profileID == "" && !resolver.Supports(uaid, parsed) {
			continue
		}

		context := UAIDProfileResolverContext{
			ParsedUAID:  parsed,
			DID:         derivedDID,
			DIDDocument: didDocument,
			ResolveDID: func(resolveCtx context.Context, did string) (*DIDDocument, error) {
				return registry.ResolveDID(resolveCtx, did)
			},
			ResolveDIDProfile: func(resolveCtx context.Context, did string, resolverContext DIDProfileResolverContext) (*UAIDResolutionResult, error) {
				return registry.resolveDIDProfile(resolveCtx, did, resolverContext)
			},
			ResolveUaidProfileByID: func(resolveCtx context.Context, requestedProfileID string, requestedUAID string) (*UAIDResolutionResult, error) {
				parsedUAID, parseErr := ParseUAID(requestedUAID)
				if parseErr != nil {
					return nil, nil
				}
				return registry.resolveUaidProfileInternal(
					resolveCtx,
					requestedUAID,
					parsedUAID,
					requestedProfileID,
					resolver.ProfileID(),
				)
			},
		}

		resolved, resolveErr := resolver.ResolveProfile(ctx, uaid, context)
		if resolveErr != nil {
			return nil, resolveErr
		}
		if resolved == nil {
			continue
		}

		isErrorProfile := resolved.Error != nil || !resolved.Metadata.Resolved
		if isErrorProfile {
			if profileID == "" {
				continue
			}
			return resolved, nil
		}
		return mergeResolutionResult(fallback, resolved), nil
	}

	return nil, nil
}

func (registry *ResolverRegistry) resolveDIDProfile(
	ctx context.Context,
	did string,
	resolverContext DIDProfileResolverContext,
) (*UAIDResolutionResult, error) {
	var didDocument *DIDDocument
	var err error
	if resolverContext.DIDDocument != nil {
		didDocument = resolverContext.DIDDocument
	} else {
		didDocument, err = registry.ResolveDID(ctx, did)
		if err != nil {
			return nil, err
		}
	}

	fallbackID := did
	if resolverContext.UAID != "" {
		fallbackID = resolverContext.UAID
	}
	fallback := registry.buildFallbackProfile(fallbackID, did, didDocument)
	for _, resolver := range registry.didProfileResolvers {
		resolved, resolveErr := resolver.ResolveProfile(ctx, did, DIDProfileResolverContext{
			UAID:        resolverContext.UAID,
			ParsedUAID:  resolverContext.ParsedUAID,
			DIDDocument: didDocument,
		})
		if resolveErr != nil {
			return nil, resolveErr
		}
		if resolved != nil {
			return mergeResolutionResult(fallback, resolved), nil
		}
	}
	return fallback, nil
}

func (registry *ResolverRegistry) buildFallbackProfile(
	id string,
	did string,
	didDocument *DIDDocument,
) *UAIDResolutionResult {
	alsoKnownAs := make([]string, 0)
	if did != "" && did != id {
		alsoKnownAs = append(alsoKnownAs, did)
	}
	if didDocument != nil {
		alsoKnownAs = append(alsoKnownAs, didDocument.AlsoKnownAs...)
	}

	result := &UAIDResolutionResult{
		ID:          id,
		DID:         did,
		AlsoKnownAs: dedupeStrings(alsoKnownAs),
	}

	if didDocument != nil {
		result.VerificationMethod = cloneVerificationMethods(didDocument.VerificationMethod)
		result.Authentication = append([]string{}, didDocument.Authentication...)
		result.AssertionMethod = append([]string{}, didDocument.AssertionMethod...)
		result.Service = cloneServices(didDocument.Service)
	}

	return result
}

func mergeResolutionResult(
	fallback *UAIDResolutionResult,
	resolved *UAIDResolutionResult,
) *UAIDResolutionResult {
	if fallback == nil {
		return resolved
	}
	if resolved == nil {
		return fallback
	}

	merged := *fallback
	if strings.TrimSpace(resolved.ID) != "" {
		merged.ID = resolved.ID
	}
	if strings.TrimSpace(resolved.DID) != "" {
		merged.DID = resolved.DID
	}
	if len(resolved.AlsoKnownAs) > 0 {
		merged.AlsoKnownAs = dedupeStrings(append(merged.AlsoKnownAs, resolved.AlsoKnownAs...))
	}
	if len(resolved.VerificationMethod) > 0 {
		merged.VerificationMethod = cloneVerificationMethods(resolved.VerificationMethod)
	}
	if len(resolved.Authentication) > 0 {
		merged.Authentication = append([]string{}, resolved.Authentication...)
	}
	if len(resolved.AssertionMethod) > 0 {
		merged.AssertionMethod = append([]string{}, resolved.AssertionMethod...)
	}
	if len(resolved.Service) > 0 {
		merged.Service = cloneServices(resolved.Service)
	}
	if resolved.Metadata != (UAIDMetadata{}) {
		merged.Metadata = resolved.Metadata
	}
	if resolved.Error != nil {
		merged.Error = resolved.Error
	}
	return &merged
}

func deriveDIDFromParsedUAID(parsed ParsedUAID) string {
	if parsed.Target == "aid" {
		proto := strings.TrimSpace(parsed.Params["proto"])
		nativeID := strings.TrimSpace(parsed.Params["nativeId"])
		if proto == "hcs-10" {
			network, accountID, ok := parseHederaCAIP10(nativeID)
			if ok {
				return "did:hedera:" + network + ":" + accountID
			}
		}
		return ""
	}

	src := strings.TrimSpace(parsed.Params["src"])
	if src != "" {
		decoded, err := decodeMultibaseB58btc(src)
		if err == nil {
			decodedDID := strings.TrimSpace(string(decoded))
			if strings.HasPrefix(decodedDID, "did:") {
				return decodedDID
			}
		}
	}

	id := strings.TrimSpace(parsed.ID)
	if strings.HasPrefix(id, "mainnet:") ||
		strings.HasPrefix(id, "testnet:") ||
		strings.HasPrefix(id, "previewnet:") ||
		strings.HasPrefix(id, "devnet:") {
		return "did:hedera:" + id
	}

	proto := strings.TrimSpace(parsed.Params["proto"])
	nativeID := strings.TrimSpace(parsed.Params["nativeId"])
	if proto == "hcs-10" {
		network, _, ok := parseHederaCAIP10(nativeID)
		if ok {
			return "did:hedera:" + network + ":" + id
		}
	}

	return ""
}

func parseHederaCAIP10(value string) (string, string, bool) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "hedera:") {
		return "", "", false
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) != 3 {
		return "", "", false
	}

	network := strings.TrimSpace(parts[1])
	accountID := strings.TrimSpace(parts[2])
	if network == "" || accountID == "" {
		return "", "", false
	}
	if !isHederaCAIP10(trimmed) {
		return "", "", false
	}
	return network, accountID, true
}
