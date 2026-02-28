package hcs27

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
}

func TestExecutionFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	ctx := context.Background()

	_, _, err := client.CreateCheckpointTopic(ctx, CreateTopicOptions{
		UseOperatorAsAdmin: true,
	})
	// because `hederaClient` hits network, this usually fails
	if err == nil { t.Fatal("expected fail") }

	_, err = client.PublishCheckpoint(ctx, "0.0.1", CheckpointMetadata{}, "", "")
	if err == nil { t.Fatal("expected fail") }
}

func TestMirrorFailures(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"messages": []}`))
	}))
	defer ts.Close()

	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
		MirrorBaseURL:      ts.URL,
	})

	ctx := context.Background()

	_, err := client.GetCheckpoints(ctx, "0.0.1", nil)
	if err != nil { t.Fatalf("unexpected err: %v", err) }

	_, err = client.ResolveHCS1Reference(ctx, "hcs://1/0.0.1")
	if err == nil { t.Fatal("expected fail") }
}
