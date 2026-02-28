package hcs21

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS21Integration_CreatePublishResolve(t *testing.T) {
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
		t.Fatalf("failed to create hcs21 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	registryTopicID, _, err := client.CreateRegistryTopic(ctx, CreateRegistryTopicOptions{
		TTL:                 3600,
		Indexed:             true,
		Type:                TopicTypeAdapterRegistry,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create registry topic: %v", err)
	}

	declaration, err := client.BuildDeclaration(BuildDeclarationParams{
		Op:        OperationRegister,
		AdapterID: "adapter-go-sdk",
		Entity:    "service",
		Package: AdapterPackage{
			Registry:  "npm",
			Name:      "adapter-go-sdk",
			Version:   "1.0.0",
			Integrity: "sha384-demo",
		},
		Manifest: "hcs://1/0.0.1",
		Config: map[string]any{
			"type": "state",
		},
	})
	if err != nil {
		t.Fatalf("failed to build declaration: %v", err)
	}

	_, _, err = client.PublishDeclaration(ctx, PublishDeclarationOptions{
		TopicID:     registryTopicID,
		Declaration: declaration,
	})
	if err != nil {
		t.Fatalf("failed to publish declaration: %v", err)
	}

	var envelopes []AdapterDeclarationEnvelope
	for attempt := 0; attempt < 20; attempt++ {
		envelopes, err = client.FetchDeclarations(ctx, registryTopicID, FetchDeclarationsOptions{
			Limit: 25,
			Order: "desc",
		})
		if err == nil && len(envelopes) > 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to fetch declarations: %v", err)
	}
	if len(envelopes) == 0 {
		t.Fatalf("expected at least one declaration")
	}
}

