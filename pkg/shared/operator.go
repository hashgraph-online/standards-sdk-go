package shared

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type OperatorConfig struct {
	AccountID  string
	PrivateKey string
	Network    string
}

var dotenvLoadOnce sync.Once

// OperatorConfigFromEnv performs the requested operation.
func OperatorConfigFromEnv() (OperatorConfig, error) {
	loadDotEnvIfPresent()

	network := firstNonEmptyEnv("HEDERA_NETWORK", "NETWORK")
	if network == "" {
		network = NetworkTestnet
	}

	accountID := firstNonEmptyEnv("HEDERA_ACCOUNT_ID", "HEDERA_OPERATOR_ID", "ACCOUNT_ID")
	if accountID == "" {
		accountID = firstNonEmptyEnv("OPERATOR_ID")
	}
	privateKey := firstNonEmptyEnv("HEDERA_PRIVATE_KEY", "HEDERA_OPERATOR_KEY", "PRIVATE_KEY")
	if privateKey == "" {
		privateKey = firstNonEmptyEnv("OPERATOR_KEY")
	}

	switch strings.ToLower(network) {
	case NetworkMainnet:
		if scopedAccount := firstNonEmptyEnv(
			"MAINNET_HEDERA_ACCOUNT_ID",
			"MAINNET_HEDERA_OPERATOR_ID",
			"MAINNET_OPERATOR_ID",
		); scopedAccount != "" {
			accountID = scopedAccount
		}
		if scopedKey := firstNonEmptyEnv(
			"MAINNET_HEDERA_PRIVATE_KEY",
			"MAINNET_HEDERA_OPERATOR_KEY",
			"MAINNET_OPERATOR_KEY",
		); scopedKey != "" {
			privateKey = scopedKey
		}
	case NetworkTestnet:
		if scopedAccount := firstNonEmptyEnv(
			"TESTNET_HEDERA_ACCOUNT_ID",
			"TESTNET_HEDERA_OPERATOR_ID",
			"TESTNET_OPERATOR_ID",
		); scopedAccount != "" {
			accountID = scopedAccount
		}
		if scopedKey := firstNonEmptyEnv(
			"TESTNET_HEDERA_PRIVATE_KEY",
			"TESTNET_HEDERA_OPERATOR_KEY",
			"TESTNET_OPERATOR_KEY",
		); scopedKey != "" {
			privateKey = scopedKey
		}
	}

	if accountID == "" {
		return OperatorConfig{}, fmt.Errorf("HEDERA_ACCOUNT_ID is required")
	}
	if privateKey == "" {
		return OperatorConfig{}, fmt.Errorf("HEDERA_PRIVATE_KEY is required")
	}

	return OperatorConfig{
		AccountID:  accountID,
		PrivateKey: privateKey,
		Network:    network,
	}, nil
}

func loadDotEnvIfPresent() {
	dotenvLoadOnce.Do(func() {
		startPaths := make([]string, 0, 2)

		if cwd, err := os.Getwd(); err == nil {
			startPaths = append(startPaths, cwd)
		}
		if _, currentFile, _, ok := runtime.Caller(0); ok {
			startPaths = append(startPaths, filepath.Dir(currentFile))
		}

		seenCandidates := make(map[string]struct{})
		for _, start := range startPaths {
			current := start
			for {
				candidate := filepath.Join(current, ".env")
				if _, exists := seenCandidates[candidate]; !exists {
					seenCandidates[candidate] = struct{}{}
					if _, statErr := os.Stat(candidate); statErr == nil {
						if loadDotEnvFile(candidate) {
							return
						}
						return
					}
				}

				parent := filepath.Dir(current)
				if parent == current {
					break
				}
				current = parent
			}
		}
	})
}

func loadDotEnvFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	loadedAny := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		separator := strings.Index(line, "=")
		if separator <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:separator])
		if !isValidEnvKey(key) {
			continue
		}
		if _, alreadySet := os.LookupEnv(key); alreadySet {
			continue
		}

		value := strings.TrimSpace(line[separator+1:])
		if len(value) >= 2 {
			first := value[0]
			last := value[len(value)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if setErr := os.Setenv(key, value); setErr == nil {
			loadedAny = true
		}
	}

	return loadedAny
}

func isValidEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for index, character := range key {
		if (character >= 'A' && character <= 'Z') ||
			(character >= 'a' && character <= 'z') ||
			(index > 0 && character >= '0' && character <= '9') ||
			character == '_' {
			continue
		}
		return false
	}
	return true
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

// ParsePrivateKey parses the provided input value.
func ParsePrivateKey(raw string) (hedera.PrivateKey, error) {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return hedera.PrivateKey{}, fmt.Errorf("private key cannot be empty")
	}

	ed25519Key, edErr := hedera.PrivateKeyFromStringEd25519(candidate)
	if edErr == nil {
		return ed25519Key, nil
	}

	ecdsaKey, ecdsaErr := hedera.PrivateKeyFromStringECDSA(candidate)
	if ecdsaErr == nil {
		return ecdsaKey, nil
	}

	genericKey, genericErr := hedera.PrivateKeyFromString(candidate)
	if genericErr == nil {
		return genericKey, nil
	}

	return hedera.PrivateKey{}, fmt.Errorf(
		"failed to parse private key as ED25519 (%v), ECDSA (%v), or generic (%v)",
		edErr,
		ecdsaErr,
		genericErr,
	)
}
