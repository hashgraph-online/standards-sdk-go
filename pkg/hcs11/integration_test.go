package hcs11

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS11Integration_FetchProfileByAccountID(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live integration tests")
	}

	accountID := strings.TrimSpace(os.Getenv("HCS11_INTEGRATION_ACCOUNT_ID"))
	network := strings.TrimSpace(os.Getenv("HCS11_INTEGRATION_NETWORK"))
	if network == "" {
		network = shared.NetworkTestnet
	}
	if accountID == "" {
		operatorConfig, err := shared.OperatorConfigFromEnv()
		if err != nil {
			t.Skipf("skipping integration test: %v", err)
		}
		accountID = operatorConfig.AccountID
		network = operatorConfig.Network
	}

	client, err := NewClient(ClientConfig{
		Network: network,
	})
	if err != nil {
		t.Fatalf("failed to initialize hcs11 client: %v", err)
	}

	response, err := client.FetchProfileByAccountID(context.Background(), accountID, network)
	if err != nil {
		t.Fatalf("FetchProfileByAccountID failed: %v", err)
	}
	if !response.Success {
		t.Skipf("account does not expose a resolvable HCS-11 memo: %s", response.Error)
	}
	if response.Profile == nil {
		t.Fatalf("expected profile data")
	}
	t.Logf("resolved HCS-11 profile display_name=%s type=%d", response.Profile.DisplayName, response.Profile.Type)
}
