package inscriber

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/shared"
	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestInscriberIntegration_AuthStartExecute(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}
	if os.Getenv("RUN_INSCRIBER_INTEGRATION") != "1" {
		t.Skip("set RUN_INSCRIBER_INTEGRATION=1 to run live inscription integration test")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping inscriber integration test: %v", err)
	}

	networkValue := strings.ToLower(strings.TrimSpace(os.Getenv("INSCRIBER_HEDERA_NETWORK")))
	if networkValue == "" {
		networkValue = shared.NetworkTestnet
	}
	network := Network(networkValue)
	if network != NetworkTestnet && network != NetworkMainnet {
		t.Fatalf("INSCRIBER_HEDERA_NETWORK must be testnet or mainnet, got %q", networkValue)
	}

	accountID := operatorConfig.AccountID
	privateKey := operatorConfig.PrivateKey
	if network == NetworkTestnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}
	if network == NetworkMainnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(privateKey) == "" {
		t.Fatalf("inscriber integration requires account ID and private key for network %s", network)
	}

	authBaseURL := os.Getenv("INSCRIPTION_AUTH_BASE_URL")
	apiBaseURL := os.Getenv("INSCRIPTION_API_BASE_URL")
	holderID := accountID

	authClient := NewAuthClient(authBaseURL)
	ctx := context.Background()
	authResult, err := authClient.Authenticate(
		ctx,
		accountID,
		privateKey,
		network,
	)
	if err != nil {
		t.Fatalf("failed to authenticate inscription client: %v", err)
	}
	if strings.TrimSpace(authResult.APIKey) == "" {
		t.Fatalf("auth flow returned empty API key")
	}

	client, err := NewClient(Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: apiBaseURL,
	})
	if err != nil {
		t.Fatalf("failed to create inscription client: %v", err)
	}

	content := base64.StdEncoding.EncodeToString([]byte("go-sdk integration inscription"))
	job, err := client.StartInscription(ctx, StartInscriptionRequest{
		File: FileInput{
			Type:     "base64",
			Base64:   content,
			FileName: "integration.txt",
			MimeType: "text/plain",
		},
		HolderID: holderID,
		Mode:     ModeFile,
	})
	if err != nil {
		t.Fatalf("failed to start inscription: %v", err)
	}
	if strings.TrimSpace(job.TransactionBytes) == "" {
		t.Fatalf("start inscription did not return transaction bytes")
	}
	t.Logf(
		"inscription transaction bytes length=%d prefix=%s",
		len(job.TransactionBytes),
		safePrefix(job.TransactionBytes, 48),
	)

	decodedTransactionBytes, decodeLabel, err := decodeTransactionBytes(job.TransactionBytes)
	if err != nil {
		t.Fatalf("failed to decode transaction bytes for inspection: %v", err)
	}
	t.Logf("decoded transaction bytes using %s", decodeLabel)

	decodedTransaction, err := hedera.TransactionFromBytes(decodedTransactionBytes)
	if err != nil {
		t.Fatalf("failed to parse transaction bytes for inspection: %v", err)
	}

	inscriptionTransactionID, err := hedera.TransactionGetTransactionID(decodedTransaction)
	if err != nil {
		t.Fatalf("failed to read transaction ID from inscription bytes: %v", err)
	}
	payerAccount := ""
	if inscriptionTransactionID.AccountID != nil {
		payerAccount = inscriptionTransactionID.AccountID.String()
	}
	t.Logf("inscription transaction payer account=%s operator account=%s", payerAccount, accountID)

	signatures, err := hedera.TransactionGetSignatures(decodedTransaction)
	if err != nil {
		t.Fatalf("failed to inspect transaction signatures: %v", err)
	}
	t.Logf("inscription transaction has signatures for %d account(s)", len(signatures))

	transactionID, err := ExecuteTransaction(ctx, job.TransactionBytes, HederaClientConfig{
		AccountID:  accountID,
		PrivateKey: privateKey,
		Network:    network,
	})
	if err != nil {
		t.Fatalf("failed to execute inscription transaction: %v", err)
	}
	t.Logf("executed inscription transaction: %s", transactionID)

	waited, err := client.WaitForInscription(ctx, transactionID, WaitOptions{
		MaxAttempts: 180,
		Interval:    2 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed while waiting for inscription completion: %v", err)
	}
	if !waited.Completed && !strings.EqualFold(waited.Status, "completed") {
		t.Fatalf("inscription did not complete successfully, status=%s", waited.Status)
	}
}

func TestInscriberIntegration_HighLevelInscribe_DefaultWebSocket(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}
	if os.Getenv("RUN_INSCRIBER_INTEGRATION") != "1" {
		t.Skip("set RUN_INSCRIBER_INTEGRATION=1 to run live inscription integration test")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping inscriber integration test: %v", err)
	}

	networkValue := strings.ToLower(strings.TrimSpace(os.Getenv("INSCRIBER_HEDERA_NETWORK")))
	if networkValue == "" {
		networkValue = shared.NetworkTestnet
	}
	network := Network(networkValue)
	if network == "" {
		network = NetworkTestnet
	}
	accountID := operatorConfig.AccountID
	privateKey := operatorConfig.PrivateKey
	if network == NetworkTestnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}
	if network == NetworkMainnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}

	waitForConfirmation := true
	result, err := Inscribe(
		context.Background(),
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   []byte("go-sdk high-level websocket integration"),
			FileName: "high-level.txt",
			MimeType: "text/plain",
		},
		HederaClientConfig{
			AccountID:  accountID,
			PrivateKey: privateKey,
			Network:    network,
		},
		InscriptionOptions{
			Mode:                ModeFile,
			Network:             network,
			WaitForConfirmation: &waitForConfirmation,
			WaitMaxAttempts:     120,
			WaitInterval:        2000,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("high-level Inscribe failed: %v", err)
	}

	if !result.Confirmed {
		t.Fatalf("expected confirmed inscription, got %+v", result)
	}

	inscriptionResult, ok := result.Result.(InscriptionResult)
	if !ok {
		t.Fatalf("expected InscriptionResult payload, got %T", result.Result)
	}
	if strings.TrimSpace(inscriptionResult.TransactionID) == "" {
		t.Fatalf("expected transaction ID in high-level result")
	}
}

func TestInscriberIntegration_GenerateQuote_BulkFiles(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}
	if os.Getenv("RUN_INSCRIBER_INTEGRATION") != "1" {
		t.Skip("set RUN_INSCRIBER_INTEGRATION=1 to run live inscription integration test")
	}

	accountID, privateKey, network, err := resolveInscriberIntegrationCredentials()
	if err != nil {
		t.Skipf("skipping inscriber integration test: %v", err)
	}

	zipBuffer, zipErr := buildIntegrationZip("skill.json", `{"name":"example-skill","version":"1.0.0"}`)
	if zipErr != nil {
		t.Fatalf("failed to build integration zip: %v", zipErr)
	}

	result, quoteErr := GenerateQuote(
		context.Background(),
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   zipBuffer,
			FileName: "skill-bundle.zip",
			MimeType: "application/zip",
		},
		HederaClientConfig{
			AccountID:  accountID,
			PrivateKey: privateKey,
			Network:    network,
		},
		InscriptionOptions{
			Mode: ModeBulkFiles,
		},
		nil,
	)
	if quoteErr != nil {
		t.Fatalf("GenerateQuote failed: %v", quoteErr)
	}
	if !result.Quote {
		t.Fatalf("expected quote response")
	}
	quote, ok := result.Result.(QuoteResult)
	if !ok {
		t.Fatalf("expected QuoteResult payload, got %T", result.Result)
	}
	if strings.TrimSpace(quote.TotalCostHBAR) == "" {
		t.Fatalf("expected quote totalCostHbar")
	}
}

func TestInscriberIntegration_RegistryBrokerSkillQuote(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}
	if os.Getenv("RUN_REGISTRY_BROKER_INTEGRATION") != "1" {
		t.Skip("set RUN_REGISTRY_BROKER_INTEGRATION=1 to run broker integration tests")
	}

	apiKey := strings.TrimSpace(os.Getenv("REGISTRY_BROKER_API_KEY"))
	if apiKey == "" {
		t.Skip("set REGISTRY_BROKER_API_KEY to run live broker quote integration")
	}

	zipBuffer, zipErr := buildIntegrationZip("skill.json", `{"name":"example-skill","version":"1.0.0"}`)
	if zipErr != nil {
		t.Fatalf("failed to build integration zip: %v", zipErr)
	}

	quote, quoteErr := GetRegistryBrokerQuote(
		context.Background(),
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   zipBuffer,
			FileName: "skill-bundle.zip",
			MimeType: "application/zip",
		},
		InscribeViaRegistryBrokerOptions{
			APIKey:   apiKey,
			Mode:     ModeBulkFiles,
			Metadata: map[string]any{"kind": "skill", "skillName": "example-skill"},
		},
	)
	if quoteErr != nil {
		t.Fatalf("GetRegistryBrokerQuote failed: %v", quoteErr)
	}
	if quote.Credits <= 0 && quote.TotalCostHBAR <= 0 {
		t.Fatalf("expected broker quote credits or totalCostHBAR to be populated")
	}
}

func TestInscriberIntegration_RegistryBrokerSkillInscribe(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}
	if os.Getenv("RUN_REGISTRY_BROKER_INTEGRATION") != "1" {
		t.Skip("set RUN_REGISTRY_BROKER_INTEGRATION=1 to run broker integration tests")
	}

	apiKey := strings.TrimSpace(os.Getenv("REGISTRY_BROKER_API_KEY"))
	if apiKey == "" {
		t.Skip("set REGISTRY_BROKER_API_KEY to run live broker inscription integration")
	}

	zipBuffer, zipErr := buildIntegrationZip("skill.json", `{"name":"example-skill","version":"1.0.0"}`)
	if zipErr != nil {
		t.Fatalf("failed to build integration zip: %v", zipErr)
	}

	waitForConfirmation := true
	result, inscribeErr := InscribeSkillViaRegistryBroker(
		context.Background(),
		InscriptionInput{
			Type:     InscriptionInputTypeBuffer,
			Buffer:   zipBuffer,
			FileName: "skill-bundle.zip",
			MimeType: "application/zip",
		},
		SkillInscriptionOptions{
			InscribeViaRegistryBrokerOptions: InscribeViaRegistryBrokerOptions{
				APIKey:              apiKey,
				Mode:                ModeBulkFiles,
				WaitForConfirmation: &waitForConfirmation,
				WaitTimeoutMs:       180000,
			},
			SkillName:    "example-skill",
			SkillVersion: "1.0.0",
		},
	)
	if inscribeErr != nil {
		t.Fatalf("InscribeSkillViaRegistryBroker failed: %v", inscribeErr)
	}
	if !result.Confirmed {
		t.Fatalf("expected confirmed broker inscription, got %+v", result)
	}
	if strings.TrimSpace(result.HRL) == "" {
		t.Fatalf("expected hrl on completed skill inscription")
	}
}

func resolveInscriberIntegrationCredentials() (string, string, Network, error) {
	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		return "", "", "", err
	}

	networkValue := strings.ToLower(strings.TrimSpace(os.Getenv("INSCRIBER_HEDERA_NETWORK")))
	if networkValue == "" {
		networkValue = shared.NetworkTestnet
	}
	network := Network(networkValue)
	if network != NetworkTestnet && network != NetworkMainnet {
		return "", "", "", os.ErrInvalid
	}

	accountID := operatorConfig.AccountID
	privateKey := operatorConfig.PrivateKey
	if network == NetworkTestnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("TESTNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}
	if network == NetworkMainnet {
		if scopedAccountID := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_ACCOUNT_ID")); scopedAccountID != "" {
			accountID = scopedAccountID
		}
		if scopedPrivateKey := strings.TrimSpace(os.Getenv("MAINNET_HEDERA_PRIVATE_KEY")); scopedPrivateKey != "" {
			privateKey = scopedPrivateKey
		}
	}
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(privateKey) == "" {
		return "", "", "", os.ErrInvalid
	}

	return accountID, privateKey, network, nil
}

func buildIntegrationZip(entryName string, content string) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(buffer)

	entry, err := zipWriter.Create(entryName)
	if err != nil {
		return nil, err
	}
	if _, err := entry.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeTransactionBytes(transactionBytes string) ([]byte, string, error) {
	trimmed := strings.TrimSpace(transactionBytes)
	if trimmed == "" {
		return nil, "", os.ErrInvalid
	}

	base64Bytes, base64Err := base64.StdEncoding.DecodeString(trimmed)
	if base64Err == nil {
		return base64Bytes, "base64", nil
	}

	hexBytes, hexErr := hex.DecodeString(strings.TrimPrefix(trimmed, "0x"))
	if hexErr == nil {
		return hexBytes, "hex", nil
	}

	return nil, "", base64Err
}

func safePrefix(value string, length int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= length {
		return trimmed
	}
	return trimmed[:length]
}
