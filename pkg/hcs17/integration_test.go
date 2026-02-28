package hcs17

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS17Integration_ComputeAndPublishStateHash(t *testing.T) {
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
		t.Fatalf("failed to create hcs17 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sourceTopicID, err := client.CreateStateTopic(ctx, CreateTopicOptions{
		TTLSeconds: 600,
	})
	if err != nil {
		t.Fatalf("failed to create source state topic: %v", err)
	}
	publishTopicID, err := client.CreateStateTopic(ctx, CreateTopicOptions{
		TTLSeconds: 600,
	})
	if err != nil {
		t.Fatalf("failed to create publish state topic: %v", err)
	}
	t.Logf("created HCS-17 topics source=%s publish=%s", sourceTopicID, publishTopicID)

	initialMessage := client.CreateStateHashMessage(
		"bootstrap-state-hash",
		operatorConfig.AccountID,
		[]string{},
		"go-sdk-hcs17-bootstrap",
		nil,
	)
	if _, err := client.SubmitMessage(ctx, sourceTopicID, initialMessage, ""); err != nil {
		t.Fatalf("failed to submit bootstrap source message: %v", err)
	}

	operatorKey, err := shared.ParsePrivateKey(operatorConfig.PrivateKey)
	if err != nil {
		t.Fatalf("failed to parse operator private key: %v", err)
	}

	result, err := client.ComputeAndPublish(ctx, ComputeAndPublishOptions{
		AccountID:        operatorConfig.AccountID,
		AccountPublicKey: operatorKey.PublicKey(),
		Topics:           []string{sourceTopicID},
		PublishTopicID:   publishTopicID,
		Memo:             "go-sdk-hcs17-compute",
	})
	if err != nil {
		t.Fatalf("failed to compute and publish state hash: %v", err)
	}
	t.Logf("published computed state hash=%s sequence=%d", result.StateHash, result.Receipt.TopicSequenceNumber)

	var latest *MessageRecord
	for attempt := 0; attempt < 20; attempt++ {
		record, fetchErr := client.GetLatestMessage(ctx, publishTopicID)
		if fetchErr == nil && record != nil {
			latest = record
			break
		}
		time.Sleep(3 * time.Second)
	}

	if latest == nil {
		t.Fatalf("failed to resolve published HCS-17 state hash message from mirror node")
	}
	if latest.Message.StateHash != result.StateHash {
		t.Fatalf("state hash mismatch: got %s want %s", latest.Message.StateHash, result.StateHash)
	}
}
