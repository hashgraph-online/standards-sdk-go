package hcs15

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/go-sdk/pkg/shared"
)

func TestHCS15Integration_CreateBaseAndPetalAccounts(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}

	operatorConfig, err := shared.OperatorConfigFromEnv()
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if strings.EqualFold(operatorConfig.Network, shared.NetworkMainnet) && os.Getenv("ALLOW_MAINNET_INTEGRATION") != "1" {
		t.Skip("resolved mainnet credentials; set ALLOW_MAINNET_INTEGRATION=1 to allow live mainnet writes")
	}

	client, err := NewClient(ClientConfig{
		OperatorAccountID:  operatorConfig.AccountID,
		OperatorPrivateKey: operatorConfig.PrivateKey,
		Network:            operatorConfig.Network,
	})
	if err != nil {
		t.Fatalf("failed to create hcs15 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	baseResult, err := client.CreateBaseAccount(ctx, BaseAccountCreateOptions{
		InitialBalanceHbar: 2,
		AccountMemo:        "go-sdk-hcs15-base",
	})
	if err != nil {
		t.Fatalf("failed to create base account: %v", err)
	}
	t.Logf("created base account %s evm=%s", baseResult.AccountID, baseResult.EVMAddress)

	petalResult, err := client.CreatePetalAccount(ctx, PetalAccountCreateOptions{
		BasePrivateKey:     baseResult.PrivateKey.String(),
		InitialBalanceHbar: 1,
		AccountMemo:        "go-sdk-hcs15-petal",
	})
	if err != nil {
		t.Fatalf("failed to create petal account: %v", err)
	}
	t.Logf("created petal account %s", petalResult.AccountID)

	verified := false
	for attempt := 0; attempt < 20; attempt++ {
		isMatch, verifyErr := client.VerifyPetalAccount(ctx, petalResult.AccountID, baseResult.AccountID)
		if verifyErr == nil && isMatch {
			verified = true
			break
		}
		time.Sleep(3 * time.Second)
	}

	if !verified {
		t.Fatalf("failed to verify petal account key matches base account key")
	}
}
