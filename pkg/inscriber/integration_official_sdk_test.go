package inscriber

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestInscriberIntegration_OfficialSDKPresignedTransferExecute(t *testing.T) {
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

	authClient := NewAuthClient(os.Getenv("INSCRIPTION_AUTH_BASE_URL"))
	authResult, err := authClient.Authenticate(context.Background(), accountID, privateKey, network)
	if err != nil {
		t.Fatalf("failed to authenticate inscription client: %v", err)
	}

	client, err := NewClient(Config{
		APIKey:  authResult.APIKey,
		Network: network,
		BaseURL: os.Getenv("INSCRIPTION_API_BASE_URL"),
	})
	if err != nil {
		t.Fatalf("failed to create inscription client: %v", err)
	}

	job, err := client.StartInscription(context.Background(), StartInscriptionRequest{
		File: FileInput{
			Type:     "base64",
			Base64:   base64.StdEncoding.EncodeToString([]byte("go-sdk official execute integration")),
			FileName: "official-execute.txt",
			MimeType: "text/plain",
		},
		HolderID: accountID,
		Mode:     ModeFile,
	})
	if err != nil {
		t.Fatalf("failed to start inscription: %v", err)
	}

	rawBytes, _, err := decodeTransactionBytes(job.TransactionBytes)
	if err != nil {
		t.Fatalf("failed to decode transaction bytes: %v", err)
	}

	operatorAccountID, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		t.Fatalf("failed to parse operator account ID: %v", err)
	}
	operatorPrivateKey, err := parseOperatorPrivateKey(context.Background(), string(network), accountID, privateKey)
	if err != nil {
		t.Fatalf("failed to parse operator private key: %v", err)
	}

	transaction, err := decodeTransferTransaction(rawBytes)
	if err != nil {
		t.Fatalf("failed to decode transfer transaction: %v", err)
	}
	transaction.SetRegenerateTransactionID(false)

	signableBodies, err := transaction.GetSignableNodeBodyBytesList()
	if err != nil {
		t.Fatalf("failed to inspect signable node bodies: %v", err)
	}
	for _, signableBody := range signableBodies {
		signature := operatorPrivateKey.Sign(signableBody.Body)
		_, err = transaction.AddSignatureV2(
			operatorPrivateKey.PublicKey(),
			signature,
			signableBody.TransactionID,
			signableBody.NodeID,
		)
		if err != nil {
			t.Fatalf("failed to add operator signature via AddSignatureV2: %v", err)
		}
	}

	executionClient, err := shared.NewHederaClient(string(network))
	if err != nil {
		t.Fatalf("failed to create Hedera client: %v", err)
	}
	defer executionClient.Close()
	executionClient.SetOperator(operatorAccountID, operatorPrivateKey)

	_, err = transaction.Execute(executionClient)
	if err == nil {
		t.Fatalf("expected official SDK execute to fail for a third-party payer inscription transfer")
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "INVALID_SIGNATURE") {
		t.Fatalf("expected INVALID_SIGNATURE from official SDK execute, got %v", err)
	}

	transactionID, err := ExecuteTransaction(context.Background(), job.TransactionBytes, HederaClientConfig{
		AccountID:  accountID,
		PrivateKey: privateKey,
		Network:    network,
	})
	if err != nil {
		t.Fatalf("expected ExecuteTransaction to succeed after official SDK failure: %v", err)
	}
	if strings.TrimSpace(transactionID) == "" {
		t.Fatalf("expected ExecuteTransaction to return a transaction ID")
	}
}
