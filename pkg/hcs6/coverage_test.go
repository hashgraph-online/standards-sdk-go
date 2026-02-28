package hcs6

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

	_, err = NewClient(ClientConfig{Network: "testnet"})
	if err == nil {
		t.Fatal("expected missing op")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1"})
	if err == nil {
		t.Fatal("expected missing pk")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1", OperatorPrivateKey: "invalid"})
	if err == nil {
		t.Fatal("expected invalid pk")
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

	_, err = client.RegisterEntry(ctx, "invalid", RegisterEntryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.SubmitMessage(ctx, "0.0.1", Message{P: "hcs-6", Op: OperationRegister, TopicID: "0.0.2"}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.GetRegistry(ctx, "invalid", QueryRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestCovBuildAndParseTopicMemo(t *testing.T) {
	memo := BuildTopicMemo(86400)
	parsed, ok := ParseTopicMemo(memo)
	if !ok {
		t.Fatal("expected parse success")
	}
	if parsed.TTL != 86400 {
		t.Fatal("expected ttl")
	}

	_, ok2 := ParseTopicMemo("")
	if ok2 {
		t.Fatal("expected fail")
	}

	_, ok3 := ParseTopicMemo("hcs-6:abc")
	if ok3 {
		t.Fatal("expected fail")
	}
}

func TestCovBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo()
	if memo == "" {
		t.Fatal("expected memo")
	}
}

func TestCovBuildHRL(t *testing.T) {
	hrl := BuildHRL("0.0.123")
	if hrl == "" {
		t.Fatal("expected hrl")
	}
}

func TestCovValidateTopicID(t *testing.T) {
	if !ValidateTopicID("0.0.1") {
		t.Fatal("expected valid")
	}
	if ValidateTopicID("invalid") {
		t.Fatal("expected invalid")
	}
}

func TestCovValidateTTL(t *testing.T) {
	if !ValidateTTL(86400) {
		t.Fatal("expected valid")
	}
	if ValidateTTL(0) {
		t.Fatal("expected invalid")
	}
}
