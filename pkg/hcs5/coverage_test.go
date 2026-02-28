package hcs5

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashgraph-online/standards-sdk-go/pkg/inscriber"
	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewClient(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()

	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

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
	if client == nil {
		t.Fatal("expected client")
	}
}

func TestMint(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	res, err := client.Mint(context.Background(), MintOptions{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.Success || res.Error == "" {
		t.Fatal("expected missing metadata early exit")
	}

	_, err = client.Mint(context.Background(), MintOptions{
		MetadataTopicID: "0.0.2",
	})
	if err == nil {
		t.Fatal("expected built error empty token id")
	}

	_, err = client.Mint(context.Background(), MintOptions{
		MetadataTopicID: "0.0.2",
		TokenID:         "0.0.123",
		SupplyKey:       "invalid",
	})
	if err == nil {
		t.Fatal("expected supply key fail")
	}

	// Will fail execute because we are fake testnet
	_, err = client.Mint(context.Background(), MintOptions{
		MetadataTopicID: "0.0.2",
		TokenID:         "0.0.123",
	})
	if err == nil {
		t.Fatal("expected execution fail")
	}
}

func TestCreateHashinal(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	// Missing setup so this will immediately fail authentication request in the internal call
	_, err := client.CreateHashinal(context.Background(), CreateHashinalOptions{
		Request: inscriber.StartInscriptionRequest{},
	})
	if err == nil {
		t.Fatal("expected auth connection to fail")
	}
}

func TestBuildMintTxFailures(t *testing.T) {
	_, err := BuildMintTx("", "meta", "memo")
	if err == nil {
		t.Fatal("expected missing token id")
	}
	
	_, err = BuildMintTx("invalid-token", "meta", "memo")
	if err == nil {
		t.Fatal("expected invalid token id")
	}
}

func TestBuildMintWithHRLTxFailures(t *testing.T) {
	_, err := BuildMintWithHRLTx("0.0.1", "", "memo")
	if err == nil {
		t.Fatal("expected metadata missing err")
	}
}

func TestCreateHashinalAuthAndStartMocks(t *testing.T) {
	// Let's test the inner parts being mocked out so we hit lines 169+
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/request-signature" {
			w.Write([]byte(`{"message": "msg"}`))
		} else if r.URL.Path == "/api/auth/authenticate" {
			w.Write([]byte(`{"apiKey": "validKey", "user": {"sessionToken": "tkn"}}`))
		}
	}))
	defer authSrv.Close()

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/inscriptions/start-inscription" {
			w.Write([]byte(`{"tx_id": "0.0.1-123-123", "transactionBytes": ""}`))
		}
	}))
	defer apiSrv.Close()

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
		InscriberAuthURL:   authSrv.URL,
		InscriberAPIURL:    apiSrv.URL,
	})

	_, err := client.CreateHashinal(context.Background(), CreateHashinalOptions{
		Request: inscriber.StartInscriptionRequest{
			HolderID: "0.0.2",
			File:     inscriber.FileInput{Type: "url", URL: "http://example.com"},
		},
	})
	if err == nil {
		t.Fatal("expected error due to missing transactionBytes in mock response")
	}
}
