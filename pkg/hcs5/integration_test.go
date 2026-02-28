package hcs5

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestHCS5Integration_MintWithExistingHCS1Topic(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}

	operatorConfig, err := resolveHCS5IntegrationOperatorConfig()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	tokenID := strings.TrimSpace(os.Getenv("HCS5_INTEGRATION_TOKEN_ID"))
	supplyKey := strings.TrimSpace(os.Getenv("HCS5_INTEGRATION_SUPPLY_KEY"))
	if tokenID == "" || supplyKey == "" {
		generatedTokenID, generatedSupplyKey, generateErr := createIntegrationToken(
			context.Background(),
			operatorConfig,
		)
		if generateErr != nil {
			t.Fatalf("failed to create integration token: %v", generateErr)
		}
		tokenID = generatedTokenID
		supplyKey = generatedSupplyKey
		t.Logf("generated integration token %s", tokenID)
	}

	authClient := inscriber.NewAuthClient("")
	network := inscriber.NetworkTestnet
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) {
		network = inscriber.NetworkMainnet
	}
	authResult, err := authClient.Authenticate(
		context.Background(),
		operatorConfig.AccountID,
		operatorConfig.PrivateKey,
		network,
	)
	if err != nil {
		t.Fatalf("failed to authenticate inscriber client: %v", err)
	}

	inscriberClient, err := inscriber.NewClient(inscriber.Config{
		APIKey:  authResult.APIKey,
		Network: network,
	})
	if err != nil {
		t.Fatalf("failed to initialize inscriber client: %v", err)
	}

	started, err := inscriberClient.StartInscription(context.Background(), inscriber.StartInscriptionRequest{
		File: inscriber.FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString([]byte("go-sdk hcs5 integration")),
			FileName: "hcs5-integration.txt",
			MimeType: "text/plain",
		},
		HolderID: operatorConfig.AccountID,
		Mode:     inscriber.ModeFile,
	})
	if err != nil {
		t.Fatalf("failed to start inscription: %v", err)
	}

	executedTransactionID, err := inscriber.ExecuteTransaction(
		context.Background(),
		started.TransactionBytes,
		inscriber.HederaClientConfig{
			AccountID:  operatorConfig.AccountID,
			PrivateKey: operatorConfig.PrivateKey,
			Network:    network,
		},
	)
	if err != nil {
		t.Fatalf("failed to execute inscription transaction: %v", err)
	}

	waited, err := inscriberClient.WaitForInscription(context.Background(), executedTransactionID, inscriber.WaitOptions{
		MaxAttempts: 90,
		Interval:    2 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to wait for inscription completion: %v", err)
	}
	if strings.TrimSpace(waited.TopicID) == "" {
		t.Fatalf("inscription completion did not include topic ID")
	}

	client, err := NewClient(ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to initialize hcs5 client: %v", err)
	}

	response, err := client.Mint(context.Background(), MintOptions{
		TokenID:         tokenID,
		MetadataTopicID: waited.TopicID,
		SupplyKey:       supplyKey,
		Memo:            "go-sdk hcs5 integration",
	})
	if err != nil {
		t.Fatalf("Mint failed: %v", err)
	}
	if !response.Success {
		t.Fatalf("Mint response reported failure: %+v", response)
	}
	if response.SerialNumber <= 0 {
		t.Fatalf("expected serial number, got %d", response.SerialNumber)
	}
}

func resolveHCS5IntegrationOperatorConfig() (shared.OperatorConfig, error) {
	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		return shared.OperatorConfig{}, err
	}

	network := strings.ToLower(strings.TrimSpace(os.Getenv("INSCRIBER_HEDERA_NETWORK")))
	if network == "" {
		network = shared.NetworkTestnet
	}

	switch network {
	case shared.NetworkMainnet:
		accountID := operatorConfig.AccountID
		privateKey := operatorConfig.PrivateKey
		if scopedAccountID := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
		return shared.OperatorConfig{
			AccountID:  accountID,
			PrivateKey: privateKey,
			Network:    shared.NetworkMainnet,
		}, nil
	case shared.NetworkTestnet:
		accountID := operatorConfig.AccountID
		privateKey := operatorConfig.PrivateKey
		if scopedAccountID := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
		return shared.OperatorConfig{
			AccountID:  accountID,
			PrivateKey: privateKey,
			Network:    shared.NetworkTestnet,
		}, nil
	}

	return operatorConfig, nil
}

func createIntegrationToken(ctx context.Context, operatorConfig shared.OperatorConfig) (string, string, error) {
	network, err := shared.NormalizeNetwork(operatorConfig.Network)
	if err != nil {
		return "", "", err
	}

	accountID, err := hedera.AccountIDFromString(strings.TrimSpace(operatorConfig.AccountID))
	if err != nil {
		return "", "", fmt.Errorf("invalid operator account ID: %w", err)
	}
	operatorKey, err := shared.ParsePrivateKey(operatorConfig.PrivateKey)
	if err != nil {
		return "", "", err
	}
	supplyKey, err := hedera.PrivateKeyGenerateEd25519()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate supply key: %w", err)
	}

	hederaClient, err := shared.NewHederaClient(network)
	if err != nil {
		return "", "", err
	}
	hederaClient.SetOperator(accountID, operatorKey)

	createTransaction, err := hedera.NewTokenCreateTransaction().
		SetTokenName(fmt.Sprintf("go-sdk-hcs5-%d", time.Now().UnixMilli())).
		SetTokenSymbol("GHS5").
		SetTokenType(hedera.TokenTypeNonFungibleUnique).
		SetSupplyType(hedera.TokenSupplyTypeFinite).
		SetMaxSupply(1000).
		SetInitialSupply(0).
		SetDecimals(0).
		SetTreasuryAccountID(accountID).
		SetAutoRenewAccount(accountID).
		SetAdminKey(operatorKey.PublicKey()).
		SetSupplyKey(supplyKey.PublicKey()).
		Execute(hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute token create transaction: %w", err)
	}

	receipt, err := createTransaction.GetReceipt(hederaClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to get token create receipt: %w", err)
	}
	if receipt.TokenID == nil {
		return "", "", fmt.Errorf("token create receipt did not include token ID")
	}

	return receipt.TokenID.String(), supplyKey.String(), nil
}
