package hcs16

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewClientCoverage(t *testing.T) {
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

	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if client.HederaClient() == nil {
		t.Fatal("expected hedera client")
	}
	if client.MirrorClient() == nil {
		t.Fatal("expected mirror client")
	}
}

func TestParseTopicMemoCoverage(t *testing.T) {
	client := &Client{}
	
	res := client.ParseTopicMemo("hcs-16:0.0.1:0")
	if res.TopicType != FloraTopicTypeCommunication {
		t.Fatal("unexpected type")
	}
	if res.FloraAccountID != "0.0.1" {
		t.Fatal("unexpected account id")
	}

	res2 := client.ParseTopicMemo("hcs-16:0.0.1:1")
	if res2.TopicType != FloraTopicTypeTransaction {
		t.Fatal("unexpected type")
	}

	res3 := client.ParseTopicMemo("hcs-16:0.0.1:2")
	if res3.TopicType != FloraTopicTypeState {
		t.Fatal("unexpected type")
	}

	res4 := client.ParseTopicMemo("hcs-16:0.0.1:9")
	if res4 != nil {
		t.Fatal("expected nil")
	}

	res5 := client.ParseTopicMemo("hcs-16:0.0.1")
	if res5 != nil {
		t.Fatal("expected nil")
	}
}

func TestAssembleKeyListCoverage(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	pubKey := pk.PublicKey()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key": {"key": "` + pubKey.String() + `"}}`))
	}))
	defer ts.Close()

	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
		MirrorBaseURL:      ts.URL,
	})

	list, err := client.AssembleKeyList(context.Background(), []string{"0.0.2", "0.0.3"}, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if list == nil {
		t.Fatal("expected list")
	}

	_, err = client.AssembleSubmitKeyList(context.Background(), []string{"0.0.2"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBuildFloraTopicCreateTxsCoverage(t *testing.T) {
	client := &Client{}
	txs, err := client.BuildFloraTopicCreateTxs("0.0.1", hedera.NewKeyList(), hedera.NewKeyList(), "0.0.2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(txs) != 3 {
		t.Fatal("expected 3 txs")
	}
}

func TestExecuteNetworkCommandsFail(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1", // invalid operator
		OperatorPrivateKey: pk.String(),
	})
	
	client.HederaClient().Close() // Force execution fail

	_, err := client.CreateFloraTopic(context.Background(), CreateFloraTopicOptions{
		FloraAccountID: "0.0.2",
		TopicType: FloraTopicTypeCommunication,
	})
	if err == nil { t.Fatal("expected fail") }

	_, _, err = client.CreateFloraAccount(context.Background(), CreateFloraAccountOptions{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendFloraCreated(context.Background(), "0.0.1", "0.0.2", "0.0.3", FloraTopics{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendTransaction(context.Background(), "0.0.1", "0.0.2", "s", "d")
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendStateUpdate(context.Background(), "0.0.1", "0.0.2", "hash", nil, "0.0.3", nil, "m", "t", nil)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendFloraJoinRequest(context.Background(), "0.0.1", "0.0.2", "0.0.3", 1, "t", 2, nil)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendFloraJoinVote(context.Background(), "0.0.1", "0.0.2", "0.0.3", true, 1, 2, nil)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SendFloraJoinAccepted(context.Background(), "0.0.1", "0.0.2", []string{"m"}, nil, nil)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.SignSchedule(context.Background(), "0.0.1", pk)
	if err == nil { t.Fatal("expected fail") }

	_, err = client.PublishFloraCreated(context.Background(), "0.0.1", "0.0.2", "0.0.3", FloraTopics{})
	if err == nil { t.Fatal("expected fail") }

	_, err = client.CreateFloraProfile(context.Background(), CreateFloraProfileOptions{
		FloraAccountID: "0.0.2",
	})
	// Mock inscriber failure inside CreateFloraProfile
	if err == nil { t.Fatal("expected fail") }
}

func TestConvertFloraMembers(t *testing.T) {
	res := convertFloraMembers([]FloraMember{{AccountID: "0.0.1"}, {AccountID: "0.0.2"}})
	if len(res) != 2 || res[0].AccountID != "0.0.1" {
		t.Fatal("unexpected output")
	}
}
