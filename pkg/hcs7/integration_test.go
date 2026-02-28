package hcs7

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS7Integration_CreateRegisterResolve(t *testing.T) {
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
		t.Fatalf("failed to create hcs7 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	registryResult, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create hcs7 registry: %v", err)
	}
	targetRegistry, err := client.CreateRegistry(ctx, CreateRegistryOptions{
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create target registry: %v", err)
	}

	_, err = client.RegisterMetadata(ctx, RegisterMetadataOptions{
		RegistryTopicID: registryResult.TopicID,
		MetadataTopicID: targetRegistry.TopicID,
		Weight:          100,
		Tags:            []string{"go-sdk", "hcs-7"},
		Data: map[string]any{
			"ra": "test",
		},
		Memo: "go-sdk-hcs7-metadata",
	})
	if err != nil {
		t.Fatalf("failed to register metadata: %v", err)
	}

	var registry RegistryTopic
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
		t.Fatalf("failed to fetch registry: %v", err)
	}
	if len(registry.Entries) == 0 {
		t.Fatalf("expected at least one entry")
	}
}

