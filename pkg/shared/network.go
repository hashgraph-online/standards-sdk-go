package shared

import (
	"fmt"
	"strings"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

const (
	NetworkMainnet = "mainnet"
	NetworkTestnet = "testnet"
)

// NormalizeNetwork performs the requested operation.
func NormalizeNetwork(network string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(network))
	if normalized == "" {
		return NetworkTestnet, nil
	}

	switch normalized {
	case NetworkMainnet, NetworkTestnet:
		return normalized, nil
	default:
		return "", fmt.Errorf("unsupported network %q", network)
	}
}

// NewHederaClient creates a new HederaClient.
func NewHederaClient(network string) (*hedera.Client, error) {
	normalized, err := NormalizeNetwork(network)
	if err != nil {
		return nil, err
	}

	if normalized == NetworkMainnet {
		return hedera.ClientForMainnet(), nil
	}

	return hedera.ClientForTestnet(), nil
}
