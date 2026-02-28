package hcs14

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ANSDNSWebResolver struct {
	dnsLookup        DNSLookupFunc
	httpClient       *http.Client
	supportedSchemes map[string]struct{}
}

type ANSDNSWebResolverOptions struct {
	DNSLookup        DNSLookupFunc
	HTTPClient       *http.Client
	SupportedSchemes []string
}

type ansDNSRecord struct {
	version string
	url     string
}

type endpointCandidate struct {
	key      string
	endpoint string
	parsed   *url.URL
}

type parsedAgentCard struct {
	ansName   string
	endpoints map[string]string
}

var semverPattern = regexp.MustCompile(`^(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

// NewANSDNSWebResolver creates a new ANSDNSWebResolver.
func NewANSDNSWebResolver(options ANSDNSWebResolverOptions) *ANSDNSWebResolver {
	lookup := options.DNSLookup
	if lookup == nil {
		lookup = nodeDNSTXTLookup
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	supportedSchemes := map[string]struct{}{"https": {}, "wss": {}}
	if len(options.SupportedSchemes) > 0 {
		supportedSchemes = map[string]struct{}{}
		for _, scheme := range options.SupportedSchemes {
			normalized := strings.ToLower(strings.TrimSpace(scheme))
			if normalized != "" {
				supportedSchemes[normalized] = struct{}{}
			}
		}
	}

	return &ANSDNSWebResolver{
		dnsLookup:        lookup,
		httpClient:       httpClient,
		supportedSchemes: supportedSchemes,
	}
}

// ProfileID returns the resolver profile identifier.
func (resolver *ANSDNSWebResolver) ProfileID() string {
	return ANSDNSWebProfileID
}

// Supports reports whether the resolver supports the provided input.
func (resolver *ANSDNSWebResolver) Supports(_ string, parsed ParsedUAID) bool {
	if parsed.Target != "aid" {
		return false
	}
	if strings.TrimSpace(parsed.Params["registry"]) != "ans" {
		return false
	}
	return isFQDN(parsed.Params["nativeId"])
}

// ResolveProfile resolves the requested identifier data.
func (resolver *ANSDNSWebResolver) ResolveProfile(
	ctx context.Context,
	uaid string,
	_ UAIDProfileResolverContext,
) (*UAIDResolutionResult, error) {
	return resolver.Resolve(ctx, uaid)
}

// Resolve resolves the requested identifier data.
func (resolver *ANSDNSWebResolver) Resolve(
	ctx context.Context,
	uaid string,
) (*UAIDResolutionResult, error) {
	parsed, err := ParseUAID(uaid)
	if err != nil {
		return resolver.buildError(uaid, "ERR_INVALID_UAID", err.Error(), nil), nil
	}
	if parsed.Target != "aid" {
		return resolver.buildError(uaid, "ERR_NOT_APPLICABLE", "ANS profile applies only to uaid:aid identifiers", nil), nil
	}
	if strings.TrimSpace(parsed.Params["registry"]) != "ans" {
		return resolver.buildError(uaid, "ERR_NOT_APPLICABLE", "ANS profile requires registry=ans", nil), nil
	}

	nativeID := strings.TrimSpace(parsed.Params["nativeId"])
	if !isFQDN(nativeID) {
		return resolver.buildError(uaid, "ERR_NOT_APPLICABLE", "ANS profile requires nativeId as FQDN", nil), nil
	}

	uid := strings.TrimSpace(parsed.Params["uid"])
	if uid == "" || uid == "0" {
		return resolver.buildError(uaid, "ERR_NOT_APPLICABLE", "ANS profile requires non-zero uid parameter", nil), nil
	}

	protocol := strings.TrimSpace(parsed.Params["proto"])
	if protocol == "" || protocol == "0" {
		return resolver.buildError(uaid, "ERR_PROTOCOL_UNSPECIFIED", "ANS profile requires usable proto parameter", nil), nil
	}

	uaidVersion := normalizeVersion(parsed.Params["version"])
	if rawVersion := strings.TrimSpace(parsed.Params["version"]); rawVersion != "" && uaidVersion == "" {
		return resolver.buildError(uaid, "ERR_NOT_APPLICABLE", "invalid version parameter in UAID", nil), nil
	}

	normalizedNativeID := normalizeDomain(nativeID)
	dnsName := "_ans." + normalizedNativeID
	txtRecords, err := resolver.dnsLookup(ctx, dnsName)
	if err != nil {
		return nil, err
	}
	if len(txtRecords) == 0 {
		return resolver.buildError(uaid, "ERR_NO_DNS_RECORD", "no _ans TXT record found", map[string]any{"dnsName": dnsName}), nil
	}

	validRecords := make([]ansDNSRecord, 0)
	for _, txtRecord := range txtRecords {
		record := parseANSDNSRecord(txtRecord)
		if record != nil {
			validRecords = append(validRecords, *record)
		}
	}
	if len(validRecords) == 0 {
		return resolver.buildError(uaid, "ERR_INVALID_ANS_RECORD", "invalid _ans TXT payload", map[string]any{"dnsName": dnsName}), nil
	}

	matchingRecords := validRecords
	if uaidVersion != "" {
		versionedRecords := make([]ansDNSRecord, 0)
		for _, record := range validRecords {
			if record.version != "" {
				versionedRecords = append(versionedRecords, record)
			}
		}
		if len(versionedRecords) == 0 {
			return resolver.buildError(
				uaid,
				"ERR_VERSION_MISMATCH",
				"UAID specifies version but DNS record has no version fields",
				map[string]any{"dnsName": dnsName, "uaidVersion": uaidVersion},
			), nil
		}

		matchingRecords = make([]ansDNSRecord, 0)
		for _, record := range versionedRecords {
			if record.version == uaidVersion {
				matchingRecords = append(matchingRecords, record)
			}
		}
		if len(matchingRecords) == 0 {
			return resolver.buildError(
				uaid,
				"ERR_VERSION_MISMATCH",
				"UAID version does not match ANS DNS TXT record",
				map[string]any{"dnsName": dnsName, "uaidVersion": uaidVersion},
			), nil
		}
	}

	sort.SliceStable(matchingRecords, func(left int, right int) bool {
		return matchingRecords[left].url < matchingRecords[right].url
	})
	selectedRecord := matchingRecords[0]

	agentCardPayload, err := resolver.fetchJSON(ctx, selectedRecord.url)
	if err != nil {
		return resolver.buildError(
			uaid,
			"ERR_AGENT_CARD_INVALID",
			"agent card retrieval failed",
			map[string]any{"stage": "fetch", "agentCardUrl": selectedRecord.url, "reason": err.Error()},
		), nil
	}

	card, err := parseAgentCard(agentCardPayload)
	if err != nil {
		return resolver.buildError(
			uaid,
			"ERR_AGENT_CARD_INVALID",
			"agent card is missing required fields",
			map[string]any{"stage": "validate", "agentCardUrl": selectedRecord.url, "reason": err.Error()},
		), nil
	}

	if card.ansName != "" {
		if card.ansName != uid {
			return resolver.buildError(
				uaid,
				"ERR_AGENT_CARD_INVALID",
				"agent card ansName does not match UAID uid",
				map[string]any{"expectedUid": uid, "actualAnsName": card.ansName},
			), nil
		}
	} else if selectedRecord.version != "" {
		expectedUID := "ans://v" + selectedRecord.version + "." + normalizedNativeID
		if uid != expectedUID {
			return resolver.buildError(
				uaid,
				"ERR_AGENT_CARD_INVALID",
				"UID does not match versioned ANS format from DNS record",
				map[string]any{"expectedUid": expectedUID, "actualUid": uid},
			), nil
		}
	}

	endpointCandidates := resolver.extractEndpointCandidates(card.endpoints)
	anchoredCandidates := make([]endpointCandidate, 0)
	for _, candidate := range endpointCandidates {
		if normalizeDomain(candidate.parsed.Hostname()) == normalizedNativeID {
			anchoredCandidates = append(anchoredCandidates, candidate)
		}
	}
	if len(anchoredCandidates) == 0 {
		return resolver.buildError(
			uaid,
			"ERR_ENDPOINT_NOT_ANCHORED",
			"no endpoint URL is anchored to nativeId host",
			map[string]any{"nativeId": normalizedNativeID},
		), nil
	}

	selectedEndpoint := selectPreferredEndpoint(anchoredCandidates, protocol)
	if selectedEndpoint == nil {
		return resolver.buildError(uaid, "ERR_ENDPOINT_NOT_FOUND", "no usable endpoint found", nil), nil
	}

	return &UAIDResolutionResult{
		ID: uaid,
		Service: []ServiceEndpoint{
			{
				ID:              uaid + "#ans-endpoint",
				Type:            "ANSService",
				ServiceEndpoint: selectedEndpoint.endpoint,
			},
		},
		Metadata: UAIDMetadata{
			Profile:            ANSDNSWebProfileID,
			Resolved:           true,
			VerificationLevel:  "metadata",
			VerificationMethod: "metadata-match",
			PrecedenceSource:   "dns",
			Protocol:           protocol,
			Endpoint:           selectedEndpoint.endpoint,
			AgentCardURL:       selectedRecord.url,
		},
	}, nil
}

func (resolver *ANSDNSWebResolver) fetchJSON(ctx context.Context, rawURL string) (map[string]any, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("accept", "application/json")

	response, err := resolver.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status %d", response.StatusCode)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func parseANSDNSRecord(record string) *ansDNSRecord {
	fields := parseSemicolonFields(record)
	if strings.ToLower(strings.TrimSpace(fields["v"])) != "ans1" {
		return nil
	}

	rawURL := strings.TrimSpace(fields["url"])
	if rawURL == "" {
		return nil
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	if strings.ToLower(parsedURL.Scheme) != "https" {
		return nil
	}

	return &ansDNSRecord{
		version: normalizeVersion(fields["version"]),
		url:     parsedURL.String(),
	}
}

func parseAgentCard(payload map[string]any) (*parsedAgentCard, error) {
	endpoints := map[string]string{}
	ansName := strings.TrimSpace(asString(payload["ansName"]))

	if rawEndpoints, ok := payload["endpoints"].(map[string]any); ok {
		for key, value := range rawEndpoints {
			if objectValue, objectOK := value.(map[string]any); objectOK {
				endpoint := strings.TrimSpace(asString(objectValue["url"]))
				if endpoint != "" {
					endpoints[key] = endpoint
				}
			}
		}
	}

	if fallbackURL := strings.TrimSpace(asString(payload["url"])); fallbackURL != "" {
		endpoints["primary"] = fallbackURL
	}

	if interfaces, ok := payload["additionalInterfaces"].([]any); ok {
		for index, item := range interfaces {
			interfaceObject, objectOK := item.(map[string]any)
			if !objectOK {
				continue
			}
			interfaceURL := strings.TrimSpace(asString(interfaceObject["url"]))
			if interfaceURL == "" {
				continue
			}
			endpoints[fmt.Sprintf("interface-%d", index)] = interfaceURL
		}
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("missing endpoints")
	}

	return &parsedAgentCard{
		ansName:   ansName,
		endpoints: endpoints,
	}, nil
}

func (resolver *ANSDNSWebResolver) extractEndpointCandidates(endpoints map[string]string) []endpointCandidate {
	candidates := make([]endpointCandidate, 0)
	for key, endpoint := range endpoints {
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			continue
		}

		scheme := strings.ToLower(strings.TrimSuffix(parsedURL.Scheme, ":"))
		if _, ok := resolver.supportedSchemes[scheme]; !ok {
			continue
		}
		candidates = append(candidates, endpointCandidate{
			key:      key,
			endpoint: parsedURL.String(),
			parsed:   parsedURL,
		})
	}
	return candidates
}

func selectPreferredEndpoint(candidates []endpointCandidate, protocol string) *endpointCandidate {
	if len(candidates) == 0 {
		return nil
	}

	sort.SliceStable(candidates, func(left int, right int) bool {
		return candidates[left].key < candidates[right].key
	})

	for _, candidate := range candidates {
		if hasPathSegment(candidate.parsed.Path, protocol) {
			choice := candidate
			return &choice
		}
	}

	choice := candidates[0]
	return &choice
}

func hasPathSegment(path string, protocol string) bool {
	normalizedProtocol := strings.ToLower(strings.TrimSpace(protocol))
	if normalizedProtocol == "" {
		return false
	}

	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if strings.ToLower(strings.TrimSpace(segment)) == normalizedProtocol {
			return true
		}
	}
	return false
}

func normalizeVersion(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "v") || strings.HasPrefix(trimmed, "V") {
		trimmed = trimmed[1:]
	}
	if !semverPattern.MatchString(trimmed) {
		return ""
	}
	return trimmed
}

func asString(value any) string {
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	return stringValue
}

func (resolver *ANSDNSWebResolver) buildError(
	uaid string,
	code string,
	message string,
	details map[string]any,
) *UAIDResolutionResult {
	return &UAIDResolutionResult{
		ID: uaid,
		Metadata: UAIDMetadata{
			Profile:  ANSDNSWebProfileID,
			Resolved: false,
		},
		Error: &UAIDResolutionError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
