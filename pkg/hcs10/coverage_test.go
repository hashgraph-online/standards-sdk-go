package hcs10

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

	_, _, err := client.CreateInboundTopic(ctx, CreateTopicOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, _, err = client.CreateOutboundTopic(ctx, CreateTopicOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, _, err = client.CreateConnectionTopic(ctx, CreateTopicOptions{InboundTopicID: "0.0.1"})
	if err == nil { t.Fatal("expected fail") }

	_, err2 := client.CreateRegistryTopic(ctx, CreateTopicOptions{})
	if err2 == nil { t.Fatal("expected fail") }

	_, err2 = client.SendConnectionRequest(ctx, "invalid", "0.0.1", "")
	if err2 == nil { t.Fatal("expected fail") }

	_, err2 = client.ConfirmConnection(ctx, "invalid", "0.0.2", "0.0.3", "0.0.1", 1, "")
	if err2 == nil { t.Fatal("expected fail") }

	_, err2 = client.SendMessage(ctx, "invalid", "0.0.1", "hello", "")
	if err2 == nil { t.Fatal("expected fail") }

	_, err2 = client.RegisterAgent(ctx, "invalid", "0.0.1", "0.0.2", "")
	if err2 == nil { t.Fatal("expected fail") }

	_, err2 = client.DeleteAgent(ctx, "invalid", "uid1", "")
	if err2 == nil { t.Fatal("expected fail") }

	_, err = client.GetTopicInfo(ctx, "invalid")
	if err == nil { t.Fatal("expected fail") }
}

func TestCovBuildMemos(t *testing.T) {
	m1 := BuildInboundMemo(86400, "0.0.1")
	if m1 == "" { t.Fatal("expected memo") }

	m2 := BuildOutboundMemo(86400)
	if m2 == "" { t.Fatal("expected memo") }

	m3 := BuildConnectionMemo(86400, "0.0.1", 1)
	if m3 == "" { t.Fatal("expected memo") }
}
