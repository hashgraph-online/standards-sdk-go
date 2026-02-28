package hcs15

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashgraph-online/standards-sdk-go/pkg/mirror"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

type Client struct {
	hederaClient *hedera.Client
	mirrorClient *mirror.Client
}

func NewClient(config ClientConfig) (*Client, error) {
	network, err := shared.NormalizeNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	operatorID := strings.TrimSpace(config.OperatorAccountID)
	if operatorID == "" {
		return nil, fmt.Errorf("operator account ID is required")
	}
	operatorKey := strings.TrimSpace(config.OperatorPrivateKey)
	if operatorKey == "" {
		return nil, fmt.Errorf("operator private key is required")
	}

	parsedOperatorID, err := hedera.AccountIDFromString(operatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid operator account ID: %w", err)
	}
	parsedOperatorKey, err := shared.ParsePrivateKey(operatorKey)
	if err != nil {
		return nil, err
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return nil, err
	}
	hederaClient.SetOperator(parsedOperatorID, parsedOperatorKey)

	mirrorClient, err := mirror.NewClient(mirror.Config{
		Network: network,
		BaseURL: config.MirrorBaseURL,
		APIKey:  config.MirrorAPIKey,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		hederaClient: hederaClient,
		mirrorClient: mirrorClient,
	}, nil
}

func (c *Client) HederaClient() *hedera.Client {
	return c.hederaClient
}

func (c *Client) MirrorClient() *mirror.Client {
	return c.mirrorClient
}

func (c *Client) CreateBaseAccount(
	ctx context.Context,
	options BaseAccountCreateOptions,
) (BaseAccountCreateResult, error) {
	_ = ctx

	privateKey, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		return BaseAccountCreateResult{}, fmt.Errorf("failed to generate ecdsa private key: %w", err)
	}
	publicKey := privateKey.PublicKey()

	initialBalance := options.InitialBalanceHbar
	if initialBalance <= 0 {
		initialBalance = 10
	}

	transaction, err := BuildBaseAccountCreateTx(BaseAccountCreateTxParams{
		PublicKey:                     publicKey,
		InitialBalanceHbar:            initialBalance,
		MaxAutomaticTokenAssociations: options.MaxAutomaticTokenAssociations,
		AccountMemo:                   options.AccountMemo,
		TransactionMemo:               options.TransactionMemo,
	})
	if err != nil {
		return BaseAccountCreateResult{}, err
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return BaseAccountCreateResult{}, fmt.Errorf("failed to execute base account create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return BaseAccountCreateResult{}, fmt.Errorf("failed to retrieve base account create receipt: %w", err)
	}
	if receipt.AccountID == nil {
		return BaseAccountCreateResult{}, fmt.Errorf("HCS-15 BASE_ACCOUNT_CREATE_FAILED")
	}

	return BaseAccountCreateResult{
		AccountID:     receipt.AccountID.String(),
		PrivateKey:    privateKey,
		PrivateKeyRaw: privateKey.StringRaw(),
		PublicKey:     publicKey,
		EVMAddress:    normalizeEVMAddress(publicKey.ToEvmAddress()),
		Receipt:       receipt,
	}, nil
}

func (c *Client) CreatePetalAccount(
	ctx context.Context,
	options PetalAccountCreateOptions,
) (PetalAccountCreateResult, error) {
	_ = ctx

	basePrivateKey := strings.TrimSpace(options.BasePrivateKey)
	if basePrivateKey == "" {
		return PetalAccountCreateResult{}, fmt.Errorf("base private key is required")
	}

	parsedBaseKey, err := hedera.PrivateKeyFromStringECDSA(basePrivateKey)
	if err != nil {
		return PetalAccountCreateResult{}, fmt.Errorf("invalid base private key: %w", err)
	}
	publicKey := parsedBaseKey.PublicKey()

	initialBalance := options.InitialBalanceHbar
	if initialBalance <= 0 {
		initialBalance = 1
	}

	transaction, err := BuildPetalAccountCreateTx(PetalAccountCreateTxParams{
		PublicKey:                     publicKey,
		InitialBalanceHbar:            initialBalance,
		MaxAutomaticTokenAssociations: options.MaxAutomaticTokenAssociations,
		AccountMemo:                   options.AccountMemo,
		TransactionMemo:               options.TransactionMemo,
	})
	if err != nil {
		return PetalAccountCreateResult{}, err
	}

	response, err := transaction.Execute(c.hederaClient)
	if err != nil {
		return PetalAccountCreateResult{}, fmt.Errorf("failed to execute petal account create transaction: %w", err)
	}
	receipt, err := response.GetReceipt(c.hederaClient)
	if err != nil {
		return PetalAccountCreateResult{}, fmt.Errorf("failed to retrieve petal account create receipt: %w", err)
	}
	if receipt.AccountID == nil {
		return PetalAccountCreateResult{}, fmt.Errorf("HCS-15 PETAL_ACCOUNT_CREATE_FAILED")
	}

	return PetalAccountCreateResult{
		AccountID: receipt.AccountID.String(),
		Receipt:   receipt,
	}, nil
}

func (c *Client) VerifyPetalAccount(
	ctx context.Context,
	petalAccountID string,
	baseAccountID string,
) (bool, error) {
	if strings.TrimSpace(petalAccountID) == "" {
		return false, fmt.Errorf("petal account ID is required")
	}
	if strings.TrimSpace(baseAccountID) == "" {
		return false, fmt.Errorf("base account ID is required")
	}

	petalInfo, err := c.mirrorClient.GetAccount(ctx, petalAccountID)
	if err != nil {
		return false, err
	}
	baseInfo, err := c.mirrorClient.GetAccount(ctx, baseAccountID)
	if err != nil {
		return false, err
	}

	petalKey := extractMirrorKey(petalInfo.Key)
	baseKey := extractMirrorKey(baseInfo.Key)
	if petalKey == "" || baseKey == "" {
		return false, nil
	}

	return petalKey == baseKey, nil
}

func normalizeEVMAddress(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		return trimmed
	}
	return "0x" + trimmed
}

func extractMirrorKey(raw map[string]any) string {
	if raw == nil {
		return ""
	}
	return extractKeyCandidate(raw)
}

func extractKeyCandidate(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		for _, key := range []string{"key", "ECDSA_secp256k1", "ed25519"} {
			if nested, ok := typed[key]; ok {
				candidate := extractKeyCandidate(nested)
				if candidate != "" {
					return candidate
				}
			}
		}
		for _, nested := range typed {
			switch nested.(type) {
			case map[string]any, []any:
				candidate := extractKeyCandidate(nested)
				if candidate != "" {
					return candidate
				}
			}
		}
	case []any:
		for _, nested := range typed {
			candidate := extractKeyCandidate(nested)
			if candidate != "" {
				return candidate
			}
		}
	}
	return ""
}
