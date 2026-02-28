package hcs12

import (
"context"
"testing"

"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestCovNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil { t.Fatal("expected err") }

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
Network:            "testnet",
OperatorAccountID:  "0.0.1",
OperatorPrivateKey: pk.String(),
	})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if client.MirrorClient() == nil { t.Fatal("expected mirror") }
}

func TestCovOperationsFailure(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
Network:            "testnet",
OperatorAccountID:  "0.0.1",
OperatorPrivateKey: pk.String(),
	})
	ctx := context.Background()

	_, err := client.CreateRegistryTopic(ctx, CreateRegistryTopicOptions{
RegistryType: RegistryTypeAction,
})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.RegisterAction(ctx, "invalid", ActionRegistration{}, "")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.RegisterAssembly(ctx, "invalid", AssemblyRegistration{}, "")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.RegisterHashLink(ctx, "invalid", HashLinksRegistration{}, "")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.GetEntries(ctx, "invalid", QueryOptions{})
	if err == nil { t.Fatal("expected fail") }
}

func TestCovBuildTransactionMemo(t *testing.T) {
	memo := BuildTransactionMemo(0, 0)
	if memo == "" { t.Fatal("expected memo") }
}
