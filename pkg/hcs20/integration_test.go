package hcs20

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs2"
	"github.com/hashgraph-online/standards-sdk-go/pkg/shared"
)

func TestHCS20Integration_EndToEnd(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run live Hedera integration tests")
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
		t.Fatalf("failed to create HCS-20 client: %v", err)
	}

	ctx := context.Background()
	registryTopicID, registryTxID, err := client.CreateRegistryTopic(ctx, hcs2.CreateRegistryOptions{
		RegistryType:        hcs2.RegistryTypeIndexed,
		TTL:                 3600,
		UseOperatorAsAdmin:  true,
		UseOperatorAsSubmit: true,
	})
	if err != nil {
		t.Fatalf("failed to create HCS-20 registry topic: %v", err)
	}
	t.Logf("created HCS-20 registry topic: %s (tx=%s)", registryTopicID, registryTxID)

	uniqueSuffix := time.Now().UTC().UnixNano()
	tick := fmt.Sprintf("goh20%d", uniqueSuffix%1_000_000_000)
	pointName := fmt.Sprintf("Go HCS-20 %d", uniqueSuffix)

	deployInfo, err := client.DeployPoints(ctx, DeployPointsOptions{
		Name:               pointName,
		Tick:               tick,
		Max:                "1000000",
		LimitPerMint:       "500000",
		UsePrivateTopic:    true,
		DisableMirrorCheck: false,
	})
	if err != nil {
		t.Fatalf("failed to deploy points: %v", err)
	}
	t.Logf("deployed points tick=%s privateTopic=%s", deployInfo.Tick, deployInfo.TopicID)

	registerResult, err := client.RegisterTopic(ctx, RegisterTopicOptions{
		TopicID:            deployInfo.TopicID,
		Name:               pointName,
		Metadata:           "go-sdk-hcs20-integration",
		IsPrivate:          true,
		DisableMirrorCheck: false,
	})
	if err != nil {
		t.Fatalf("failed to register private topic: %v", err)
	}
	t.Logf("registered private topic sequence=%d tx=%s", registerResult.SequenceNumber, registerResult.TransactionID)

	mintResult, err := client.MintPoints(ctx, MintPointsOptions{
		Tick:               tick,
		Amount:             "100",
		To:                 operatorConfig.AccountID,
		TopicID:            deployInfo.TopicID,
		DisableMirrorCheck: false,
	})
	if err != nil {
		t.Fatalf("failed to mint points: %v", err)
	}
	t.Logf("minted points sequence=%d tx=%s", mintResult.SequenceNumber, mintResult.TransactionID)

	transferResult, err := client.TransferPoints(ctx, TransferPointsOptions{
		Tick:               tick,
		Amount:             "10",
		From:               operatorConfig.AccountID,
		To:                 operatorConfig.AccountID,
		TopicID:            deployInfo.TopicID,
		DisableMirrorCheck: false,
	})
	if err != nil {
		t.Fatalf("failed to transfer points: %v", err)
	}
	t.Logf("transferred points sequence=%d tx=%s", transferResult.SequenceNumber, transferResult.TransactionID)

	burnResult, err := client.BurnPoints(ctx, BurnPointsOptions{
		Tick:               tick,
		Amount:             "25",
		From:               operatorConfig.AccountID,
		TopicID:            deployInfo.TopicID,
		DisableMirrorCheck: false,
	})
	if err != nil {
		t.Fatalf("failed to burn points: %v", err)
	}
	t.Logf("burned points sequence=%d tx=%s", burnResult.SequenceNumber, burnResult.TransactionID)

	indexer, err := NewPointsIndexer(IndexerConfig{
		Network:       operatorConfig.Network,
		MirrorBaseURL: client.MirrorClient().BaseURL(),
	})
	if err != nil {
		t.Fatalf("failed to create points indexer: %v", err)
	}

	const expectedSupply = "75"
	const expectedBalance = "75"
	var indexedInfo PointsInfo
	var infoExists bool
	for attempt := 0; attempt < 12; attempt++ {
		if err := indexer.IndexOnce(ctx, IndexOptions{
			IncludePublicTopic:   false,
			IncludeRegistryTopic: true,
			RegistryTopicID:      registryTopicID,
		}); err != nil {
			t.Fatalf("failed to index HCS-20 topics: %v", err)
		}

		indexedInfo, infoExists = indexer.GetPointsInfo(tick)
		if infoExists && indexedInfo.CurrentSupply == expectedSupply && indexer.GetBalance(tick, operatorConfig.AccountID) == expectedBalance {
			break
		}

		time.Sleep(3 * time.Second)
	}

	if !infoExists {
		t.Fatalf("expected indexed points info for tick %s", tick)
	}
	if indexedInfo.CurrentSupply != expectedSupply {
		t.Fatalf("expected indexed supply %s, got %s", expectedSupply, indexedInfo.CurrentSupply)
	}

	indexedBalance := indexer.GetBalance(tick, operatorConfig.AccountID)
	if indexedBalance != expectedBalance {
		t.Fatalf("expected indexed balance %s, got %s", expectedBalance, indexedBalance)
	}

	state := indexer.StateSnapshot()
	if len(state.Transactions) < 3 {
		t.Fatalf("expected at least 3 indexed transactions, got %d", len(state.Transactions))
	}
}
