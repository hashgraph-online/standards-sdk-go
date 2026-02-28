package hcs18

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS18Integration_DiscoveryFlow(t *testing.T) {
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
		t.Fatalf("failed to create hcs18 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	topicID, _, err := client.CreateDiscoveryTopic(ctx, CreateDiscoveryTopicOptions{
		TTLSeconds:         3600,
		UseOperatorAsAdmin: true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create discovery topic: %v", err)
	}

	_, err = client.Announce(ctx, topicID, AnnounceData{
		Account: operatorConfig.AccountID,
		Petal: PetalDescriptor{
			Name:     "go-sdk-petal",
			Priority: 1,
		},
		Capabilities: CapabilityDetails{
			Protocols: []string{"hcs-10"},
		},
	}, "")
	if err != nil {
		t.Fatalf("failed to publish announce: %v", err)
	}

	_, err = client.Propose(ctx, topicID, ProposeData{
		Proposer: operatorConfig.AccountID,
		Members: []ProposeMember{
			{
				Account:  operatorConfig.AccountID,
				Priority: 1,
			},
		},
		Config: ProposeConfig{
			Name:      "go-sdk-flora",
			Threshold: 1,
		},
	}, "")
	if err != nil {
		t.Fatalf("failed to publish propose: %v", err)
	}

	var records []MessageRecord
	for attempt := 0; attempt < 20; attempt++ {
		records, err = client.GetDiscoveryMessages(ctx, topicID, "", 20, "desc")
		if err == nil && len(records) >= 2 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to fetch discovery messages: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected at least 2 discovery messages, got %d", len(records))
	}
}

