package hcs21

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

func TestCovBuildDeclaration(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.BuildDeclaration(BuildDeclarationParams{
		Op:        OperationRegister,
		AdapterID: "adapter-id",
		Entity:    "service",
		Package: AdapterPackage{
			Registry:  "npm",
			Name:      "adapter",
			Version:   "1.0.0",
			Integrity: "sha384-abc",
		},
		Manifest: "hcs://1/0.0.1",
		Config: map[string]any{
			"type": "state",
		},
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
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

	_, _, err := client.CreateRegistryTopic(ctx, CreateRegistryTopicOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, _, err2 := client.PublishDeclaration(ctx, PublishDeclarationOptions{
		TopicID: "invalid",
		Declaration: AdapterDeclaration{
			P:         "hcs-21",
			Op:        OperationRegister,
			AdapterID: "adapter",
			Entity:    "service",
			Package: AdapterPackage{
				Registry:  "npm",
				Name:      "adapter",
				Version:   "1.0.0",
				Integrity: "sha384-abc",
			},
			Manifest: "hcs://1/0.0.1",
			Config: map[string]any{
				"type": "state",
			},
		},
	})
	if err2 == nil {
		t.Fatal("expected fail")
	}

	_, err3 := client.FetchDeclarations(ctx, "invalid", FetchDeclarationsOptions{})
	if err3 == nil {
		t.Fatal("expected fail")
	}
}

func TestCovValidationError(t *testing.T) {
	e := &ValidationError{Message: "bad"}
	if e.Error() != "bad" {
		t.Fatal("expected bad")
	}
}

func TestCovBuildRegistryMemo(t *testing.T) {
	_, err := BuildRegistryMemo(86400, true, TopicTypeAdapterRegistry, "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	_, err = BuildRegistryMemo(86400, false, TopicTypeAdapterRegistry, "not-a-pointer")
	if err == nil {
		t.Fatal("expected error")
	}
}
