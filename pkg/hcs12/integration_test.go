package hcs12

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS12Integration_CreateRegistryAndRegister(t *testing.T) {
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
		t.Fatalf("failed to create hcs12 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	createResult, err := client.CreateRegistryTopic(ctx, CreateRegistryTopicOptions{
		RegistryType:        RegistryTypeAction,
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create registry topic: %v", err)
	}
	if !createResult.Success {
		t.Fatalf("expected successful topic create result")
	}

	_, err = client.RegisterAction(ctx, createResult.TopicID, ActionRegistration{
		P:           "hcs-12",
		Op:          "register",
		Name:        "demo-action",
		Version:     "1.0.0",
		Description: "go sdk integration action",
	}, "")
	if err != nil {
		t.Fatalf("failed to register action: %v", err)
	}

	var entries []RegistryEntry
	for attempt := 0; attempt < 20; attempt++ {
		entries, err = client.GetEntries(ctx, createResult.TopicID, QueryOptions{
			Limit: 20,
			Order: "desc",
		})
		if err == nil && len(entries) > 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to fetch entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one registry entry")
	}
}
