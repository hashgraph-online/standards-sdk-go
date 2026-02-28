package hcs16

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS16Integration_CreateFloraAndPublishMessages(t *testing.T) {
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
		t.Fatalf("failed to create hcs16 client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	flora, err := client.CreateFloraAccountWithTopics(ctx, CreateFloraAccountWithTopicsOptions{
		Members:            []string{operatorConfig.AccountID},
		Threshold:          1,
		InitialBalanceHbar: 2,
	})
	if err != nil {
		t.Fatalf("failed to create flora account with topics: %v", err)
	}
	t.Logf(
		"created flora account %s topics(comm=%s tx=%s state=%s)",
		flora.FloraAccountID,
		flora.Topics.Communication,
		flora.Topics.Transaction,
		flora.Topics.State,
	)

	if _, err := client.SendFloraCreated(
		ctx,
		flora.Topics.Communication,
		operatorConfig.AccountID,
		flora.FloraAccountID,
		flora.Topics,
	); err != nil {
		t.Fatalf("failed to publish flora_created message: %v", err)
	}

	if _, err := client.SendTransaction(
		ctx,
		flora.Topics.Transaction,
		operatorConfig.AccountID,
		"0.0.1",
		"go-sdk-hcs16-integration",
	); err != nil {
		t.Fatalf("failed to publish transaction message: %v", err)
	}

	epoch := int64(1)
	if _, err := client.SendStateUpdate(
		ctx,
		flora.Topics.State,
		operatorConfig.AccountID,
		"0xgo-sdk-hcs16-state",
		&epoch,
		flora.FloraAccountID,
		[]string{flora.Topics.Communication, flora.Topics.Transaction},
		"go-sdk-hcs16-state",
		"",
		nil,
	); err != nil {
		t.Fatalf("failed to publish state update message: %v", err)
	}

	var latest *FloraMessageRecord
	for attempt := 0; attempt < 20; attempt++ {
		record, fetchErr := client.GetLatestMessage(
			ctx,
			flora.Topics.Communication,
			FloraOperationFloraCreated,
		)
		if fetchErr == nil && record != nil {
			latest = record
			break
		}
		time.Sleep(3 * time.Second)
	}

	if latest == nil {
		t.Fatalf("failed to resolve latest flora_created message from mirror node")
	}
	t.Logf(
		"resolved flora_created message sequence=%d consensus_timestamp=%s",
		latest.SequenceNumber,
		latest.ConsensusTimestamp,
	)
}
