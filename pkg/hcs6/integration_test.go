package hcs6

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS6Integration_CreateRegisterResolve(t *testing.T) {
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
		t.Fatalf("failed to create hcs6 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	registryResult, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create hcs6 registry: %v", err)
	}
	if registryResult.TopicID == "" {
		t.Fatalf("expected topic id")
	}

	targetRegistry, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create target registry: %v", err)
	}

	_, err = client.RegisterEntry(ctx, registryResult.TopicID, RegisterEntryOptions{
		TargetTopicID: targetRegistry.TopicID,
		Memo:          "go-sdk-hcs6-integration",
	})
	if err != nil {
		t.Fatalf("failed to register hcs6 entry: %v", err)
	}

	var registry TopicRegistry
	for attempt := 0; attempt < 20; attempt++ {
		registry, err = client.GetRegistry(ctx, registryResult.TopicID, QueryRegistryOptions{
			Limit: 20,
			Order: "desc",
		})
		if err == nil && len(registry.Entries) > 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to resolve registry: %v", err)
	}
	if len(registry.Entries) == 0 {
		t.Fatalf("expected at least one registry entry")
	}
	if registry.Entries[0].Message.TopicID != targetRegistry.TopicID {
		t.Fatalf("unexpected target topic id: got %s want %s", registry.Entries[0].Message.TopicID, targetRegistry.TopicID)
	}
}

