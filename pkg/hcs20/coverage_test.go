package hcs20

import (
	"context"
	"testing"
	"time"

	"github.com/hashgraph-online/standards-sdk-go/pkg/hcs2"
	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()

	_, err = NewClient(ClientConfig{Network: "testnet"})
	if err == nil {
		t.Fatal("expected err missing opt")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1"})
	if err == nil {
		t.Fatal("expected err missing pk")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "invalid-id", OperatorPrivateKey: pk.String()})
	if err == nil {
		t.Fatal("expected err invalid op string")
	}
	
	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1", OperatorPrivateKey: "invalid-pk"})
	if err == nil {
		t.Fatal("expected err invalid pk string")
	}
}

func TestClientGettersAndSetters(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if client.MirrorClient() == nil {
		t.Fatal("expected clients")
	}

	client.SetPublicTopicID("0.0.2")
	if client.PublicTopicID() != "0.0.2" {
		t.Fatal("expected 0.0.2")
	}

	client.SetRegistryTopicID("0.0.3")
	if client.RegistryTopicID() != "0.0.3" {
		t.Fatal("expected 0.0.3")
	}
}

func TestExecutionFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	// Force failure since we aren't supplying valid network or signature for these tests.
	ctx := context.Background()

	_, _, err := client.CreatePublicTopic(ctx, CreateTopicOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, _, err = client.CreateRegistryTopic(ctx, hcs2.CreateRegistryOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.DeployPoints(ctx, DeployPointsOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.MintPoints(ctx, MintPointsOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.TransferPoints(ctx, TransferPointsOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.BurnPoints(ctx, BurnPointsOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.RegisterTopic(ctx, RegisterTopicOptions{})
	if err == nil { t.Fatal("expected fail") }
}

func TestErrorsCoverage(t *testing.T) {
	err1 := NewPointsValidationError("val err", nil)
	if err1.Error() == "" {
		t.Fatal("expected err string")
	}

	err2 := NewPointsNotFoundError("not found")
	if err2.Error() == "" {
		t.Fatal("expected err string")
	}

	err3 := NewInvalidTickFormatError("bad tick")
	if err3.Error() == "" {
		t.Fatal("expected err string")
	}
}

func TestIndexerPollingFail(t *testing.T) {
	indexer, _ := NewPointsIndexer(IndexerConfig{
		Network: "testnet",
	})
	
	// Start polling immediately stops because config or context mock
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should immediately return
	indexer.StartPolling(ctx, IndexOptions{PublicTopicID: "0.0.1"}, 1*time.Second)

	// Since it's stopped/canceled via context, stop should just clear signal
	indexer.StopPolling()
	
	// And state snapshot should be 0
	if indexer.StateSnapshot().LastProcessedTimestamp != "" {
		t.Fatal("expected empty")
	}
}
