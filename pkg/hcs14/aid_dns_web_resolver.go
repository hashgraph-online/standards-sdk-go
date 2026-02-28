package hcs14

import (
	"context"
	"net/url"
	"sort"
	"strings"
)

type AIDDNSWebResolver struct {
	dnsLookup        DNSLookupFunc
	supportedSchemes map[string]struct{}
	metadataVerifier func(ctx context.Context, input AIDDNSVerificationInput) (bool, error)
	cryptoVerifier   func(ctx context.Context, input AIDDNSVerificationInput) (bool, error)
}

type AIDDNSWebResolverOptions struct {
	DNSLookup        DNSLookupFunc
	SupportedSchemes []string
	MetadataVerifier func(ctx context.Context, input AIDDNSVerificationInput) (bool, error)
	CryptoVerifier   func(ctx context.Context, input AIDDNSVerificationInput) (bool, error)
}

type aidDNSRecord struct {
	version   string
	protocol  string
	endpoint  string
	publicKey string
	keyID     string
}

type AIDDNSVerificationInput struct {
	UAID     string
	Protocol string
	Endpoint string
	Record   aidDNSRecord
}

func NewAIDDNSWebResolver(options AIDDNSWebResolverOptions) *AIDDNSWebResolver {
	lookup := options.DNSLookup
	if lookup == nil {
		lookup = nodeDNSTXTLookup
	}

	supportedSchemes := map[string]struct{}{
		"https": {},
		"http":  {},
		"wss":   {},
		"ws":    {},
	}
	if len(options.SupportedSchemes) > 0 {
		supportedSchemes = map[string]struct{}{}
		for _, scheme := range options.SupportedSchemes {
			normalized := strings.ToLower(strings.TrimSpace(scheme))
			if normalized != "" {
				supportedSchemes[normalized] = struct{}{}
			}
		}
	}

	return &AIDDNSWebResolver{
		dnsLookup:        lookup,
		supportedSchemes: supportedSchemes,
		metadataVerifier: options.MetadataVerifier,
		cryptoVerifier:   options.CryptoVerifier,
	}
}

func (resolver *AIDDNSWebResolver) ProfileID() string {
	return AIDDNSWebProfileID
}

func (resolver *AIDDNSWebResolver) Supports(_ string, parsed ParsedUAID) bool {
	if parsed.Target != "aid" {
		return false
	}
	return isFQDN(parsed.Params["nativeId"])
}

func (resolver *AIDDNSWebResolver) ResolveProfile(
	ctx context.Context,
	uaid string,
	context UAIDProfileResolverContext,
) (*UAIDResolutionResult, error) {
	parsed := context.ParsedUAID
	if parsed.Target != "aid" {
		return aidProfileError(uaid, "ERR_NOT_APPLICABLE", "AID DNS/Web profile applies only to uaid:aid identifiers", nil), nil
	}

	nativeID := strings.TrimSpace(parsed.Params["nativeId"])
	if !isFQDN(nativeID) {
		return aidProfileError(uaid, "ERR_NOT_APPLICABLE", "AID DNS/Web profile requires nativeId as FQDN", nil), nil
	}

	dnsName := "_agent." + normalizeDomain(nativeID)
	txtRecords, err := resolver.dnsLookup(ctx, dnsName)
	if err != nil {
		return nil, err
	}
	if len(txtRecords) == 0 {
		return aidProfileError(uaid, "ERR_NO_DNS_RECORD", "no _agent TXT record found", map[string]any{"dnsName": dnsName}), nil
	}

	validRecords := make([]aidDNSRecord, 0)
	endpointInvalid := false
	for _, txtRecord := range txtRecords {
		record, parseErr := resolver.parseRecord(txtRecord)
		if parseErr != "" {
			if parseErr == "ERR_ENDPOINT_INVALID" {
				endpointInvalid = true
			}
			continue
		}
		validRecords = append(validRecords, record)
	}
	if len(validRecords) == 0 {
		errorCode := "ERR_INVALID_AID_RECORD"
		errorMessage := "AID DNS TXT payload is malformed or unsupported"
		if endpointInvalid {
			errorCode = "ERR_ENDPOINT_INVALID"
			errorMessage = "AID DNS record endpoint URI is invalid or unsupported"
		}
		return aidProfileError(uaid, errorCode, errorMessage, map[string]any{"dnsName": dnsName}), nil
	}

	sort.SliceStable(validRecords, func(left int, right int) bool {
		leftKey := validRecords[left].protocol + "|" + validRecords[left].endpoint + "|" + validRecords[left].publicKey + "|" + validRecords[left].keyID
		rightKey := validRecords[right].protocol + "|" + validRecords[right].endpoint + "|" + validRecords[right].publicKey + "|" + validRecords[right].keyID
		return leftKey < rightKey
	})
	selectedRecord := validRecords[0]

	verificationLevel := "none"
	verificationMethod := ""
	verificationInput := AIDDNSVerificationInput{
		UAID:     uaid,
		Protocol: selectedRecord.protocol,
		Endpoint: selectedRecord.endpoint,
		Record:   selectedRecord,
	}

	if resolver.metadataVerifier != nil {
		verified, verifyErr := resolver.metadataVerifier(ctx, verificationInput)
		if verifyErr != nil {
			return nil, verifyErr
		}
		if !verified {
			return aidProfileError(uaid, "ERR_VERIFICATION_FAILED", "AID metadata verification failed", map[string]any{"dnsName": dnsName}), nil
		}
		verificationLevel = "metadata"
		verificationMethod = "metadata-match"
	}

	if selectedRecord.publicKey != "" && resolver.cryptoVerifier != nil {
		verified, verifyErr := resolver.cryptoVerifier(ctx, verificationInput)
		if verifyErr != nil {
			return nil, verifyErr
		}
		if !verified {
			return aidProfileError(uaid, "ERR_VERIFICATION_FAILED", "AID cryptographic verification failed", map[string]any{"dnsName": dnsName}), nil
		}
		verificationLevel = "cryptographic"
		verificationMethod = "aid-pka"
	}

	result := &UAIDResolutionResult{
		ID:  uaid,
		DID: context.DID,
		Service: []ServiceEndpoint{
			{
				ID:              uaid + "#aid-endpoint",
				Type:            "AIDService",
				ServiceEndpoint: selectedRecord.endpoint,
			},
		},
		Metadata: UAIDMetadata{
			Profile:            AIDDNSWebProfileID,
			Resolved:           true,
			VerificationLevel:  verificationLevel,
			VerificationMethod: verificationMethod,
			PrecedenceSource:   "dns",
			Protocol:           selectedRecord.protocol,
			Endpoint:           selectedRecord.endpoint,
		},
	}
	return result, nil
}

func (resolver *AIDDNSWebResolver) parseRecord(rawRecord string) (aidDNSRecord, string) {
	fields := parseSemicolonFields(rawRecord)
	version := strings.TrimSpace(fields["v"])
	protocol := strings.TrimSpace(fields["p"])
	if protocol == "" {
		protocol = strings.TrimSpace(fields["proto"])
	}
	endpoint := strings.TrimSpace(fields["u"])

	if version == "" || protocol == "" || endpoint == "" {
		return aidDNSRecord{}, "ERR_INVALID_AID_RECORD"
	}
	if !strings.HasPrefix(strings.ToLower(version), "aid") {
		return aidDNSRecord{}, "ERR_INVALID_AID_RECORD"
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return aidDNSRecord{}, "ERR_ENDPOINT_INVALID"
	}
	scheme := strings.ToLower(strings.TrimSuffix(parsedURL.Scheme, ":"))
	if _, ok := resolver.supportedSchemes[scheme]; !ok {
		return aidDNSRecord{}, "ERR_ENDPOINT_INVALID"
	}

	return aidDNSRecord{
		version:   version,
		protocol:  protocol,
		endpoint:  parsedURL.String(),
		publicKey: strings.TrimSpace(fields["k"]),
		keyID:     strings.TrimSpace(fields["i"]),
	}, ""
}

func aidProfileError(uaid string, code string, message string, details map[string]any) *UAIDResolutionResult {
	return &UAIDResolutionResult{
		ID: uaid,
		Metadata: UAIDMetadata{
			Profile:  AIDDNSWebProfileID,
			Resolved: false,
		},
		Error: &UAIDResolutionError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
