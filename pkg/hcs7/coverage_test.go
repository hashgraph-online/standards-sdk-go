package hcs7

import (
	"context"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestCovNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if client.MirrorClient() == nil {
		t.Fatal("expected mirror")
	}
}

func TestCovOperationsFailure(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	ctx := context.Background()

	_, err := client.CreateRegistry(ctx, CreateRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.RegisterConfig(ctx, RegisterConfigOptions{RegistryTopicID: "invalid"})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.RegisterMetadata(ctx, RegisterMetadataOptions{RegistryTopicID: "invalid"})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.GetRegistry(ctx, "invalid", QueryRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}
