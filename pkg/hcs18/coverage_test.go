package hcs18

import (
	"context"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	_, err = NewClient(ClientConfig{Network: "testnet"})
	if err == nil {
		t.Fatal("expected err missing op id")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1"})
	if err == nil {
		t.Fatal("expected err missing pk")
	}

	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "invalid", OperatorPrivateKey: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err = NewClient(ClientConfig{Network: "testnet", OperatorAccountID: "0.0.1", OperatorPrivateKey: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if client.MirrorClient() == nil {
		t.Fatal("expected mirror client")
	}
}

func TestClientOperationsFailure(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	ctx := context.Background()

	_, _, err := client.CreateDiscoveryTopic(ctx, CreateDiscoveryTopicOptions{
		UseOperatorAsAdmin: true,
	})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.SubmitMessage(ctx, "0.0.1", DiscoveryMessage{
		P:  "hcs-18",
		Op: OperationAnnounce,
		Data: AnnounceData{
			Account: "0.0.1",
		},
	}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.Announce(ctx, "0.0.1", AnnounceData{Account: "test"}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.Propose(ctx, "0.0.1", ProposeData{}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.Respond(ctx, "0.0.1", RespondData{}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.Complete(ctx, "0.0.1", CompleteData{}, "")
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.Withdraw(ctx, "0.0.1", WithdrawData{}, "")
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestIsProposalReady(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	proposal := TrackedProposal{
		Data: ProposeData{
			Members: []ProposeMember{{Account: "a"}, {Account: "b"}, {Account: "c"}},
		},
		Responses: map[string]TrackedResponse{
			"a": {Decision: "accept"},
			"b": {Decision: "accept"},
		},
	}
	if !client.IsProposalReady(proposal) {
		t.Fatal("expected ready")
	}

	proposal2 := TrackedProposal{
		Data: ProposeData{
			Members: []ProposeMember{{Account: "a"}, {Account: "b"}, {Account: "c"}},
		},
		Responses: map[string]TrackedResponse{
			"a": {Decision: "reject"},
		},
	}
	if client.IsProposalReady(proposal2) {
		t.Fatal("expected not ready")
	}
}

func TestDecodeDiscoveryMessage(t *testing.T) {
	_, err := decodeDiscoveryMessage("not-base64!!!")
	if err == nil {
		t.Fatal("expected err")
	}
}
