package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

	"github.com/hashgraph-online/standards-sdk-go/pkg/registrybroker"
)

const (
	defaultBaseURL       = "https://hol.org/registry/api/v1"
	domainProofPrefix    = "hol-skill-verification="
	defaultWaitSeconds   = 180
	defaultCloudflareTTL = 120
)

type cloudflareEnvelope[T any] struct {
	Success bool `json:"success"`
	Result  T    `json:"result"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type cloudflareRecord struct {
	ID string `json:"id"`
}

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
	}
}

func run() error {
	cfg, err := resolveConfigFromEnv()
	if err != nil {
		return err
	}

	client, err := registrybroker.NewRegistryBrokerClient(registrybroker.RegistryBrokerClientOptions{
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return err
	}

	if err := authenticateWithLedger(context.Background(), client, cfg.AccountID, cfg.PrivateKey, cfg.Network); err != nil {
		return err
	}
	client.SetDefaultHeader("x-account-id", cfg.AccountID)
	fmt.Printf("authenticated account %s on %s\n", cfg.AccountID, cfg.Network)

	version, err := resolveSkillVersion(context.Background(), client, cfg.SkillName, cfg.SkillVersion)
	if err != nil {
		return err
	}

	beforeTotal, beforeDomain, err := readSkillTrust(context.Background(), client, cfg.SkillName, version)
	if err != nil {
		return err
	}
	fmt.Printf("before: trust.total=%.2f verified.domainProof=%.2f\n", beforeTotal, beforeDomain)

	challengeRaw, err := client.CreateSkillDomainProofChallenge(
		context.Background(),
		registrybroker.SkillVerificationDomainProofChallengeRequest{
			Name:    cfg.SkillName,
			Version: version,
			Domain:  cfg.Domain,
		},
	)
	if err != nil {
		return err
	}

	challenge, err := parseChallenge(challengeRaw)
	if err != nil {
		return err
	}
	fmt.Printf("challenge domain=%s\n", challenge.Domain)
	fmt.Printf("txt name=%s\n", challenge.TxtRecordName)
	fmt.Printf("txt value=%s\n", challenge.TxtRecordValue)
	fmt.Printf("expires=%s\n", challenge.ExpiresAt)

	if cfg.AutoDNS && cfg.CloudflareToken != "" && cfg.CloudflareZoneID != "" {
		if err := upsertCloudflareTXT(
			cfg.CloudflareToken,
			cfg.CloudflareZoneID,
			challenge.TxtRecordName,
			challenge.TxtRecordValue,
		); err != nil {
			return err
		}
		fmt.Println("updated TXT record via Cloudflare API")
	} else {
		fmt.Println("automatic DNS update skipped; set TXT record manually before timeout")
	}

	challengeToken := challenge.TxtRecordValue
	if strings.HasPrefix(challengeToken, domainProofPrefix) {
		challengeToken = strings.TrimPrefix(challengeToken, domainProofPrefix)
	}

	if challengeToken == "" {
		return fmt.Errorf("empty challenge token")
	}

	preDNSVerifyRaw, err := client.VerifySkillDomainProof(
		context.Background(),
		registrybroker.SkillVerificationDomainProofVerifyRequest{
			Name:           cfg.SkillName,
			Version:        version,
			Domain:         challenge.Domain,
			ChallengeToken: challengeToken,
		},
	)
	if err != nil {
		return err
	}
	preDNSOk, preDNSReason := parseVerifySignal(preDNSVerifyRaw)
	fmt.Printf("pre-dns verify: ok=%t reason=%s\n", preDNSOk, preDNSReason)

	baselineTotal, baselineDomain, err := readSkillTrust(context.Background(), client, cfg.SkillName, version)
	if err != nil {
		return err
	}
	fmt.Printf("baseline: trust.total=%.2f verified.domainProof=%.2f\n", baselineTotal, baselineDomain)

	verifyDeadline := time.Now().Add(time.Duration(cfg.WaitDNSSeconds) * time.Second)
	for {
		verifyRaw, verifyErr := client.VerifySkillDomainProof(
			context.Background(),
			registrybroker.SkillVerificationDomainProofVerifyRequest{
				Name:           cfg.SkillName,
				Version:        version,
				Domain:         challenge.Domain,
				ChallengeToken: challengeToken,
			},
		)
		if verifyErr != nil {
			return verifyErr
		}

		ok, reason := parseVerifySignal(verifyRaw)
		if ok {
			fmt.Println("domain proof verified")
			break
		}

		if time.Now().After(verifyDeadline) {
			return fmt.Errorf("domain proof did not verify before timeout (last reason: %s)", reason)
		}

		fmt.Printf("dns not ready yet (%s), retrying in 10s\n", reason)
		time.Sleep(10 * time.Second)
	}

	afterTotal, afterDomain, err := readSkillTrust(context.Background(), client, cfg.SkillName, version)
	if err != nil {
		return err
	}
	fmt.Printf("after: trust.total=%.2f verified.domainProof=%.2f\n", afterTotal, afterDomain)
	fmt.Printf("delta: trust.total=%.2f verified.domainProof=%.2f\n", afterTotal-baselineTotal, afterDomain-baselineDomain)

	return nil
}

type demoConfig struct {
	BaseURL          string
	SkillName        string
	SkillVersion     string
	Domain           string
	WaitDNSSeconds   int
	AutoDNS          bool
	CloudflareToken  string
	CloudflareZoneID string
	AccountID        string
	PrivateKey       string
	Network          string
}

func resolveConfigFromEnv() (demoConfig, error) {
	baseURL := strings.TrimSpace(getEnvOrDefault("REGISTRY_BROKER_BASE_URL", defaultBaseURL))
	skillName := strings.TrimSpace(getEnvOrDefault("SKILL_NAME", ""))
	if skillName == "" {
		return demoConfig{}, fmt.Errorf("SKILL_NAME is required")
	}

	skillVersion := strings.TrimSpace(os.Getenv("SKILL_VERSION"))
	domain := strings.TrimSpace(os.Getenv("SKILL_DOMAIN_PROOF_DOMAIN"))
	waitRaw := strings.TrimSpace(getEnvOrDefault("SKILL_DOMAIN_PROOF_WAIT_DNS_SECONDS", strconv.Itoa(defaultWaitSeconds)))
	waitSeconds, err := strconv.Atoi(waitRaw)
	if err != nil || waitSeconds <= 0 {
		waitSeconds = defaultWaitSeconds
	}

	autoDNS := true
	if raw := strings.TrimSpace(os.Getenv("SKILL_DOMAIN_PROOF_AUTO_DNS")); raw != "" {
		parsed, parseErr := strconv.ParseBool(raw)
		if parseErr == nil {
			autoDNS = parsed
		}
	}

	accountID := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TESTNET_HEDERA_ACCOUNT_ID"),
		os.Getenv("HEDERA_ACCOUNT_ID"),
	))
	privateKey := strings.TrimSpace(firstNonEmpty(
		os.Getenv("TESTNET_HEDERA_PRIVATE_KEY"),
		os.Getenv("HEDERA_PRIVATE_KEY"),
	))
	network := strings.TrimSpace(firstNonEmpty(
		os.Getenv("LEDGER_NETWORK"),
		os.Getenv("HEDERA_NETWORK"),
		"testnet",
	))

	if accountID == "" || privateKey == "" {
		return demoConfig{}, fmt.Errorf("TESTNET_HEDERA_ACCOUNT_ID and TESTNET_HEDERA_PRIVATE_KEY (or HEDERA_* equivalents) are required")
	}

	return demoConfig{
		BaseURL:          baseURL,
		SkillName:        skillName,
		SkillVersion:     skillVersion,
		Domain:           domain,
		WaitDNSSeconds:   waitSeconds,
		AutoDNS:          autoDNS,
		CloudflareToken:  strings.TrimSpace(os.Getenv("CLOUDFLARE_API_TOKEN")),
		CloudflareZoneID: strings.TrimSpace(os.Getenv("CLOUDFLARE_ZONE_ID")),
		AccountID:        accountID,
		PrivateKey:       privateKey,
		Network:          network,
	}, nil
}

func authenticateWithLedger(
	ctx context.Context,
	client *registrybroker.RegistryBrokerClient,
	accountID string,
	privateKeyRaw string,
	network string,
) error {
	privateKey, err := hedera.PrivateKeyFromString(privateKeyRaw)
	if err != nil {
		return err
	}

	_, err = client.AuthenticateWithLedger(ctx, registrybroker.LedgerAuthenticationOptions{
		AccountID: accountID,
		Network:   canonicalLedgerNetwork(network),
		Sign: func(message string) (registrybroker.LedgerAuthenticationSignerResult, error) {
			signature := privateKey.Sign([]byte(message))
			return registrybroker.LedgerAuthenticationSignerResult{
				Signature:     hex.EncodeToString(signature),
				SignatureKind: "raw",
			}, nil
		},
	})
	return err
}

func resolveSkillVersion(
	ctx context.Context,
	client *registrybroker.RegistryBrokerClient,
	name string,
	requestedVersion string,
) (string, error) {
	if strings.TrimSpace(requestedVersion) != "" {
		return strings.TrimSpace(requestedVersion), nil
	}

	result, err := client.ListSkills(ctx, registrybroker.ListSkillsOptions{
		Name:  name,
		Limit: intPtr(1),
	})
	if err != nil {
		return "", err
	}

	items := getObjectList(result, "items")
	if len(items) == 0 {
		return "", fmt.Errorf("skill %s not found", name)
	}
	version := strings.TrimSpace(getString(items[0], "version"))
	if version == "" {
		return "", fmt.Errorf("skill %s has no version in response", name)
	}
	return version, nil
}

func readSkillTrust(
	ctx context.Context,
	client *registrybroker.RegistryBrokerClient,
	name string,
	version string,
) (float64, float64, error) {
	result, err := client.ListSkills(ctx, registrybroker.ListSkillsOptions{
		Name:    name,
		Version: version,
		Limit:   intPtr(1),
	})
	if err != nil {
		return 0, 0, err
	}

	items := getObjectList(result, "items")
	if len(items) == 0 {
		return 0, 0, fmt.Errorf("skill %s@%s not found", name, version)
	}
	item := items[0]

	trustTotal := getNumber(item, "trustScore")
	if trustScoresRaw, ok := item["trustScores"].(map[string]any); ok {
		if totalCandidate := getNumber(trustScoresRaw, "total"); totalCandidate > 0 {
			trustTotal = totalCandidate
		}
	}

	domainScore := 0.0
	if trustScoresRaw, ok := item["trustScores"].(map[string]any); ok {
		domainScore = getNumber(trustScoresRaw, "verified.domainProof")
	}

	return trustTotal, domainScore, nil
}

type domainChallenge struct {
	Domain         string
	TxtRecordName  string
	TxtRecordValue string
	ExpiresAt      string
}

func parseChallenge(payload registrybroker.JSONObject) (domainChallenge, error) {
	domain := strings.TrimSpace(getString(payload, "domain"))
	txtName := strings.TrimSpace(getString(payload, "txtRecordName"))
	txtValue := strings.TrimSpace(getString(payload, "txtRecordValue"))
	expiresAt := strings.TrimSpace(getString(payload, "expiresAt"))
	if domain == "" || txtName == "" || txtValue == "" {
		return domainChallenge{}, fmt.Errorf("invalid challenge response")
	}
	return domainChallenge{
		Domain:         domain,
		TxtRecordName:  txtName,
		TxtRecordValue: txtValue,
		ExpiresAt:      expiresAt,
	}, nil
}

func parseVerifySignal(payload registrybroker.JSONObject) (bool, string) {
	signalRaw, ok := payload["signal"].(map[string]any)
	if !ok {
		return false, "missing_signal"
	}
	okValue, _ := signalRaw["ok"].(bool)
	reason := strings.TrimSpace(getString(signalRaw, "reason"))
	if reason == "" {
		reason = "pending_dns"
	}
	return okValue, reason
}

func upsertCloudflareTXT(
	apiToken string,
	zoneID string,
	recordName string,
	recordValue string,
) error {
	queryPath := fmt.Sprintf(
		"/zones/%s/dns_records?type=TXT&name=%s",
		url.PathEscape(zoneID),
		url.QueryEscape(recordName),
	)

	var listEnvelope cloudflareEnvelope[[]cloudflareRecord]
	if err := cloudflareRequest(apiToken, queryPath, http.MethodGet, nil, &listEnvelope); err != nil {
		return err
	}

	body := map[string]any{
		"type":    "TXT",
		"name":    recordName,
		"content": recordValue,
		"ttl":     defaultCloudflareTTL,
	}
	bodyBytes, _ := json.Marshal(body)

	if len(listEnvelope.Result) > 0 && strings.TrimSpace(listEnvelope.Result[0].ID) != "" {
		updatePath := fmt.Sprintf(
			"/zones/%s/dns_records/%s",
			url.PathEscape(zoneID),
			url.PathEscape(listEnvelope.Result[0].ID),
		)
		var updateEnvelope cloudflareEnvelope[cloudflareRecord]
		return cloudflareRequest(apiToken, updatePath, http.MethodPut, bodyBytes, &updateEnvelope)
	}

	createPath := fmt.Sprintf("/zones/%s/dns_records", url.PathEscape(zoneID))
	var createEnvelope cloudflareEnvelope[cloudflareRecord]
	return cloudflareRequest(apiToken, createPath, http.MethodPost, bodyBytes, &createEnvelope)
}

func cloudflareRequest[T any](
	apiToken string,
	path string,
	method string,
	body []byte,
	target *T,
) error {
	request, err := http.NewRequest(method, "https://api.cloudflare.com/client/v4"+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+apiToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return err
	}

	switch envelope := any(target).(type) {
	case *cloudflareEnvelope[[]cloudflareRecord]:
		if response.StatusCode >= 300 || !envelope.Success {
			return fmt.Errorf("cloudflare list request failed")
		}
	case *cloudflareEnvelope[cloudflareRecord]:
		if response.StatusCode >= 300 || !envelope.Success {
			return fmt.Errorf("cloudflare write request failed")
		}
	}

	return nil
}

func canonicalLedgerNetwork(network string) string {
	normalized := strings.ToLower(strings.TrimSpace(network))
	switch normalized {
	case "mainnet", "hedera:mainnet":
		return "hedera:mainnet"
	default:
		return "hedera:testnet"
	}
}

func getObjectList(source map[string]any, key string) []map[string]any {
	raw, ok := source[key].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, entry := range raw {
		if typed, entryOK := entry.(map[string]any); entryOK {
			out = append(out, typed)
		}
	}
	return out
}

func getString(source map[string]any, key string) string {
	raw, ok := source[key].(string)
	if !ok {
		return ""
	}
	return raw
}

func getNumber(source map[string]any, key string) float64 {
	raw, ok := source[key]
	if !ok || raw == nil {
		return 0
	}
	switch typed := raw.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func getEnvOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func intPtr(value int) *int {
	return &value
}
