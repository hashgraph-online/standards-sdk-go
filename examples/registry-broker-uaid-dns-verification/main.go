package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

	"github.com/hashgraph-online/standards-sdk-go/pkg/registrybroker"
)

const defaultBaseURL = "https://hol.org/registry/api/v1"

type demoConfig struct {
	BaseURL    string
	UAID       string
	Persist    bool
	APIKey     string
	AccountID  string
	PrivateKey string
	Network    string
}

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := resolveConfig()
	if err != nil {
		return err
	}

	client, err := registrybroker.NewRegistryBrokerClient(registrybroker.RegistryBrokerClientOptions{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
	})
	if err != nil {
		return err
	}

	if cfg.APIKey == "" {
		if err := authenticateWithLedger(context.Background(), client, cfg.AccountID, cfg.PrivateKey, cfg.Network); err != nil {
			return err
		}
		client.SetDefaultHeader("x-account-id", cfg.AccountID)
		fmt.Printf("authenticated account %s on %s\n", cfg.AccountID, canonicalLedgerNetwork(cfg.Network))
	} else {
		fmt.Println("using provided API key authentication")
	}

	verifyResponse, err := client.VerifyUaidDNSTXT(
		context.Background(),
		registrybroker.VerificationDNSVerifyRequest{
			UAID:    cfg.UAID,
			Persist: boolPtr(cfg.Persist),
		},
	)
	if err != nil {
		return err
	}
	printResult("Live DNS Verification", verifyResponse)

	statusStored, err := client.GetVerificationDNSStatus(
		context.Background(),
		cfg.UAID,
		registrybroker.VerificationDNSStatusQuery{
			Refresh: boolPtr(false),
			Persist: boolPtr(false),
		},
	)
	if err != nil {
		return err
	}
	printResult("Status (Stored First)", statusStored)

	statusLive, err := client.GetVerificationDNSStatus(
		context.Background(),
		cfg.UAID,
		registrybroker.VerificationDNSStatusQuery{
			Refresh: boolPtr(true),
			Persist: boolPtr(false),
		},
	)
	if err != nil {
		return err
	}
	printResult("Status (Live Refresh)", statusLive)

	return nil
}

func resolveConfig() (demoConfig, error) {
	baseURL := strings.TrimSpace(firstNonEmpty(
		argValue("--base-url="),
		os.Getenv("REGISTRY_BROKER_BASE_URL"),
		defaultBaseURL,
	))
	uaid := strings.TrimSpace(firstNonEmpty(
		argValue("--uaid="),
		os.Getenv("UAID_DNS_DEMO_UAID"),
	))
	if uaid == "" {
		return demoConfig{}, fmt.Errorf("UAID_DNS_DEMO_UAID or --uaid is required")
	}

	persist := parseBoolWithDefault(firstNonEmpty(
		argValue("--persist="),
		os.Getenv("UAID_DNS_DEMO_PERSIST"),
	), true)
	apiKey := strings.TrimSpace(firstNonEmpty(
		argValue("--api-key="),
		os.Getenv("REGISTRY_BROKER_API_KEY"),
	))

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

	if apiKey == "" && (accountID == "" || privateKey == "") {
		return demoConfig{}, fmt.Errorf("set REGISTRY_BROKER_API_KEY or provide TESTNET_HEDERA_ACCOUNT_ID and TESTNET_HEDERA_PRIVATE_KEY")
	}

	return demoConfig{
		BaseURL:    baseURL,
		UAID:       uaid,
		Persist:    persist,
		APIKey:     apiKey,
		AccountID:  accountID,
		PrivateKey: privateKey,
		Network:    network,
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

func canonicalLedgerNetwork(network string) string {
	trimmed := strings.TrimSpace(strings.ToLower(network))
	if strings.HasPrefix(trimmed, "hedera:") {
		return trimmed
	}
	if trimmed == "mainnet" {
		return "hedera:mainnet"
	}
	return "hedera:testnet"
}

func printResult(label string, payload map[string]any) {
	pretty, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("\n=== %s ===\n%v\n", label, payload)
		return
	}
	fmt.Printf("\n=== %s ===\n%s\n", label, string(pretty))
}

func boolPtr(value bool) *bool {
	return &value
}

func parseBoolWithDefault(value string, fallback bool) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(trimmed)
	if err != nil {
		return fallback
	}
	return parsed
}

func argValue(prefix string) string {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(arg, prefix))
		}
	}
	return ""
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
