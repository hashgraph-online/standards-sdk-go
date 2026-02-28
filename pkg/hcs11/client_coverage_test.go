package hcs11

import (
	"context"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestClientGettersAndFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network: "testnet",
		Auth: Auth{
			OperatorID: "0.0.1",
			PrivateKey: pk.String(),
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if client.HederaClient() == nil || client.MirrorClient() == nil {
		t.Fatal("expected clients")
	}
	if client.OperatorID() != "0.0.1" {
		t.Fatal("expected id")
	}

	// Close the client so operations fail
	client.HederaClient().Close()
	ctx := context.Background()

	_ = client.CreatePersonalProfile("me", nil)

	_ = client.SetProfileForAccountMemo("0.0.1", 1)

	_, err = client.UpdateAccountMemoWithProfile(ctx, "0.0.1", "0.0.2")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.InscribeImage(ctx, []byte{1}, "img", InscribeImageOptions{})
	if err == nil { t.Fatal("expected fail") }

	resp, err := client.InscribeProfile(ctx, HCS11Profile{}, InscribeProfileOptions{})
	if err != nil || resp.Success { t.Fatal("expected no err but false success for invalid profile") }

	_, err = client.CreateAndInscribeProfile(ctx, HCS11Profile{}, false, InscribeProfileOptions{})
	if err != nil { t.Fatal("expected no generic error from wrapper") }
}

func TestGetAgentTypeFromMetadata(t *testing.T) {
	client := &Client{}
	
	res := client.GetAgentTypeFromMetadata(AgentMetadata{Type: "autonomous"})
	if res != AIAgentTypeAutonomous {
		t.Fatal("expected autonomous")
	}

	res2 := client.GetAgentTypeFromMetadata(AgentMetadata{})
	if res2 != AIAgentTypeManual {
		t.Fatal("expected manual")
	}
}

func TestAttachUAIDIfMissing(t *testing.T) {
	client := &Client{}
	ctx := context.Background()
	
	p1 := &HCS11Profile{UAID: "uaid:myagent"}
	err := client.AttachUAIDIfMissing(ctx, p1)
	if err != nil || p1.UAID != "uaid:myagent" {
		t.Fatal("should match uaid")
	}

	p2 := &HCS11Profile{}
	err = client.AttachUAIDIfMissing(ctx, p2)
	// without operator ID it just returns early
	if err != nil || p2.UAID != "" {
		t.Fatal("expected empty")
	}
}
