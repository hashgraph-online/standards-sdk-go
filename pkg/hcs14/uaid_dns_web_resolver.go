package hcs14

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type UaidDNSWebResolver struct {
	dnsLookup                DNSLookupFunc
	requireFullResolution    bool
	enableFollowupResolution bool
}

type UaidDNSWebResolverOptions struct {
	DNSLookup                DNSLookupFunc
	RequireFullResolution    bool
	EnableFollowupResolution bool
}

type uaidDNSRecord struct {
	target            string
	identifier        string
	did               string
	reconstructedUAID string
}

// NewUaidDNSWebResolver creates a new UaidDNSWebResolver.
func NewUaidDNSWebResolver(options UaidDNSWebResolverOptions) *UaidDNSWebResolver {
	lookup := options.DNSLookup
	if lookup == nil {
		lookup = nodeDNSTXTLookup
	}

	return &UaidDNSWebResolver{
		dnsLookup:                lookup,
		requireFullResolution:    options.RequireFullResolution,
		enableFollowupResolution: options.EnableFollowupResolution,
	}
}

// ProfileID returns the resolver profile identifier.
func (resolver *UaidDNSWebResolver) ProfileID() string {
	return UAIDDNSWebProfileID
}

// Supports reports whether the resolver supports the provided input.
func (resolver *UaidDNSWebResolver) Supports(_ string, parsed ParsedUAID) bool {
	nativeID := parsed.Params["nativeId"]
	return (parsed.Target == "aid" || parsed.Target == "did") && isFQDN(nativeID)
}

// ResolveProfile resolves the requested identifier data.
func (resolver *UaidDNSWebResolver) ResolveProfile(
	ctx context.Context,
	uaid string,
	resolverContext UAIDProfileResolverContext,
) (*UAIDResolutionResult, error) {
	return resolver.resolve(ctx, uaid, func(followupCtx context.Context, followupProfileID string, followupUAID string) (*UAIDResolutionResult, error) {
		if resolverContext.ResolveUaidProfileByID == nil {
			return nil, nil
		}
		return resolverContext.ResolveUaidProfileByID(followupCtx, followupProfileID, followupUAID)
	})
}

// Resolve resolves the requested identifier data.
func (resolver *UaidDNSWebResolver) Resolve(
	ctx context.Context,
	uaid string,
	followupResolver FollowupResolverFunc,
) (*UAIDResolutionResult, error) {
	return resolver.resolve(ctx, uaid, func(followupCtx context.Context, _ string, followupUAID string) (*UAIDResolutionResult, error) {
		if followupResolver == nil {
			return nil, nil
		}
		return followupResolver(followupCtx, followupUAID)
	})
}

func (resolver *UaidDNSWebResolver) resolve(
	ctx context.Context,
	uaid string,
	followupResolver func(ctx context.Context, profileID string, uaid string) (*UAIDResolutionResult, error),
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

	nativeID := parsed.Params["nativeId"]
	if !isFQDN(nativeID) {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_NOT_APPLICABLE",
				Message: "UAID DNS profile requires nativeId as an FQDN",
			},
		}, nil
	}

	normalizedNativeID := normalizeDomain(nativeID)
	dnsName := "_uaid." + normalizedNativeID

	txtRecords, err := resolver.dnsLookup(ctx, dnsName)
	if err != nil {
		return nil, err
	}
	if len(txtRecords) == 0 {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_NO_DNS_RECORD",
				Message: "no _uaid TXT record found",
				Details: map[string]any{"dnsName": dnsName},
			},
		}, nil
	}

	inputCanonical := BuildCanonicalUAID(parsed.Target, parsed.ID, canonicalizeNativeDomainParams(parsed.Params))
	validRecords := make([]uaidDNSRecord, 0)
	for _, txtRecord := range txtRecords {
		fields := parseSemicolonFields(txtRecord)
		record := validateUAIDDNSRecord(fields, normalizedNativeID)
		if record != nil {
			validRecords = append(validRecords, *record)
		}
	}

	if len(validRecords) == 0 {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_INVALID_UAID_DNS_RECORD",
				Message: "invalid _uaid TXT payload",
				Details: map[string]any{"dnsName": dnsName},
			},
		}, nil
	}

	matchingRecords := make([]uaidDNSRecord, 0)
	for _, record := range validRecords {
		if record.target == parsed.Target && record.identifier == parsed.ID && record.reconstructedUAID == inputCanonical {
			matchingRecords = append(matchingRecords, record)
		}
	}
	if len(matchingRecords) == 0 {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_UAID_MISMATCH",
				Message: "TXT fields do not match input UAID after canonical reconstruction",
				Details: map[string]any{
					"dnsName":        dnsName,
					"inputCanonical": inputCanonical,
					"candidateCount": len(validRecords),
				},
			},
		}, nil
	}

	sort.SliceStable(matchingRecords, func(left int, right int) bool {
		return matchingRecords[left].reconstructedUAID < matchingRecords[right].reconstructedUAID
	})
	selectedRecord := matchingRecords[0]

	if resolver.enableFollowupResolution && followupResolver != nil {
		followupProfiles := selectFollowupProfiles(parsed)
		failedProfiles := make([]string, 0)
		for _, profileID := range followupProfiles {
			followupResult, followupErr := followupResolver(ctx, profileID, uaid)
			if followupErr != nil {
				return nil, followupErr
			}
			if followupResult == nil {
				continue
			}
			if followupResult.Error != nil || !followupResult.Metadata.Resolved {
				failedProfiles = append(failedProfiles, profileID)
				continue
			}

			followupResult.Metadata.Profile = UAIDDNSWebProfileID
			followupResult.Metadata.ResolutionMode = "full-resolution"
			followupResult.Metadata.SelectedFollowupProfile = profileID
			followupResult.Metadata.ReconstructedUAID = selectedRecord.reconstructedUAID
			if followupResult.DID == "" {
				followupResult.DID = selectedRecord.did
			}
			return followupResult, nil
		}
		if len(failedProfiles) > 0 {
			return &UAIDResolutionResult{
				ID: uaid,
				Metadata: UAIDMetadata{
					Profile:  UAIDDNSWebProfileID,
					Resolved: false,
				},
				Error: &UAIDResolutionError{
					Code:    "ERR_FOLLOWUP_RESOLUTION_FAILED",
					Message: "follow-up profile resolution failed",
					Details: map[string]any{
						"followupProfileId":       profileIDLast(failedProfiles),
						"attemptedFailedProfiles": failedProfiles,
					},
				},
			}, nil
		}
	}

	if resolver.requireFullResolution {
		return &UAIDResolutionResult{
			ID: uaid,
			Metadata: UAIDMetadata{
				Profile:  UAIDDNSWebProfileID,
				Resolved: false,
			},
			Error: &UAIDResolutionError{
				Code:    "ERR_NO_FOLLOWUP_PROFILE",
				Message: "full resolution required but no follow-up profile is available",
			},
		}, nil
	}

	return &UAIDResolutionResult{
		ID:  uaid,
		DID: selectedRecord.did,
		Metadata: UAIDMetadata{
			Profile:           UAIDDNSWebProfileID,
			Resolved:          true,
			VerificationLevel: "dns-binding",
			ResolutionMode:    "dns-binding-only",
			ReconstructedUAID: selectedRecord.reconstructedUAID,
		},
	}, nil
}

func validateUAIDDNSRecord(fields map[string]string, queriedNativeID string) *uaidDNSRecord {
	target := strings.TrimSpace(fields["target"])
	identifier := strings.TrimSpace(fields["id"])
	uid := strings.TrimSpace(fields["uid"])
	proto := strings.TrimSpace(fields["proto"])
	nativeID := strings.TrimSpace(fields["nativeId"])

	if target != "aid" && target != "did" {
		return nil
	}
	if identifier == "" || uid == "" || proto == "" || nativeID == "" {
		return nil
	}
	if normalizeDomain(nativeID) != queriedNativeID {
		return nil
	}
	if registry, exists := fields["registry"]; exists && strings.TrimSpace(registry) == "" {
		return nil
	}

	did := strings.TrimSpace(fields["did"])
	if did != "" {
		if target != "did" || !strings.HasPrefix(did, "did:") {
			return nil
		}
	}

	params := map[string]string{
		"uid":      uid,
		"proto":    proto,
		"nativeId": nativeID,
	}
	copyIfPresent(fields, "registry", params)
	copyIfPresent(fields, "domain", params)
	copyIfPresent(fields, "src", params)
	copyIfPresent(fields, "version", params)

	return &uaidDNSRecord{
		target:            target,
		identifier:        identifier,
		did:               did,
		reconstructedUAID: BuildCanonicalUAID(target, identifier, canonicalizeNativeDomainParams(params)),
	}
}

func copyIfPresent(from map[string]string, key string, target map[string]string) {
	value := strings.TrimSpace(from[key])
	if value != "" {
		target[key] = value
	}
}

func canonicalizeNativeDomainParams(input map[string]string) map[string]string {
	output := map[string]string{}
	for key, value := range input {
		output[key] = strings.TrimSpace(value)
	}
	if nativeID := output["nativeId"]; isFQDN(nativeID) {
		output["nativeId"] = normalizeDomain(nativeID)
	}
	if domain := output["domain"]; isFQDN(domain) {
		output["domain"] = normalizeDomain(domain)
	}
	return output
}

func selectFollowupProfiles(parsed ParsedUAID) []string {
	if parsed.Target == "aid" {
		if strings.EqualFold(parsed.Params["registry"], "ans") {
			return []string{ANSDNSWebProfileID, AIDDNSWebProfileID}
		}
		return []string{AIDDNSWebProfileID}
	}
	return []string{UAIDDidResolutionProfileID}
}

func notApplicableError(uaid string, message string) *UAIDResolutionResult {
	return &UAIDResolutionResult{
		ID: uaid,
		Metadata: UAIDMetadata{
			Profile:  UAIDDNSWebProfileID,
			Resolved: false,
		},
		Error: &UAIDResolutionError{
			Code:    "ERR_NOT_APPLICABLE",
			Message: message,
		},
	}
}

// String returns the string representation.
func (resolver *UaidDNSWebResolver) String() string {
	return fmt.Sprintf("UaidDNSWebResolver(profile=%s)", UAIDDNSWebProfileID)
}

func profileIDLast(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}
