package hcs10

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS10Integration_CreateTopicsAndSendMessage(t *testing.T) {
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
		t.Fatalf("failed to create hcs10 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	inboundTopicID, _, err := client.CreateInboundTopic(ctx, CreateTopicOptions{
		TTL:                 60,
		AccountID:           operatorConfig.AccountID,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create inbound topic: %v", err)
	}
	outboundTopicID, _, err := client.CreateOutboundTopic(ctx, CreateTopicOptions{
		TTL:                 60,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create outbound topic: %v", err)
	}
	connectionTopicID, _, err := client.CreateConnectionTopic(ctx, CreateTopicOptions{
		TTL:                 60,
		InboundTopicID:      inboundTopicID,
		ConnectionID:        1,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create connection topic: %v", err)
	}

	_, err = client.SendConnectionRequest(ctx, inboundTopicID, inboundTopicID+"@"+operatorConfig.AccountID, "go-sdk-hcs10-request")
	if err != nil {
		t.Fatalf("failed to send connection request: %v", err)
	}
	_, err = client.SendMessage(ctx, connectionTopicID, inboundTopicID+"@"+operatorConfig.AccountID, "hello", "go-sdk-hcs10-message")
	if err != nil {
		t.Fatalf("failed to send connection message: %v", err)
	}
	_, err = client.RegisterAgent(ctx, outboundTopicID, operatorConfig.AccountID, inboundTopicID, "go-sdk-hcs10-register")
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	var records []MessageRecord
	for attempt := 0; attempt < 20; attempt++ {
		records, err = client.GetMessageStream(ctx, connectionTopicID, "", 20, "desc")
		if err == nil && len(records) > 0 {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to read message stream: %v", err)
	}
	if len(records) == 0 {
		t.Fatalf("expected at least one message stream record")
	}
}

