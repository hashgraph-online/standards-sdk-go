package hcs15

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestNewClientSuccess(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.12345",
		OperatorPrivateKey: key.String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.HederaClient() == nil {
		t.Fatal("expected non-nil Hedera client")
	}
	if client.MirrorClient() == nil {
		t.Fatal("expected non-nil mirror client")
	}
}

func TestNewClientMissingOperatorID(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorPrivateKey: key.String(),
	})
	if err == nil {
		t.Fatal("expected error for missing operator ID")
	}
}

func TestNewClientMissingOperatorKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Network:           "testnet",
		OperatorAccountID: "0.0.12345",
	})
	if err == nil {
		t.Fatal("expected error for missing operator key")
	}
}

func TestNewClientInvalidOperatorID(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "invalid",
		OperatorPrivateKey: key.String(),
	})
	if err == nil {
		t.Fatal("expected error for invalid operator ID")
	}
}

func TestNewClientInvalidOperatorKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.12345",
		OperatorPrivateKey: "invalid-key",
	})
	if err == nil {
		t.Fatal("expected error for invalid operator key")
	}
}

func TestNormalizeEVMAddressCoverage(t *testing.T) {
	if normalizeEVMAddress("abc") != "0xabc" {
		t.Fatal("expected 0x prefix")
	}
	if normalizeEVMAddress("0xabc") != "0xabc" {
		t.Fatal("expected no change for 0x prefix")
	}
	if normalizeEVMAddress("0Xabc") != "0Xabc" {
		t.Fatal("expected no change for 0X prefix")
	}
	if normalizeEVMAddress("   ") != "" {
		t.Fatal("expected empty string for whitespace")
	}
}

func TestExtractMirrorKeyCoverage(t *testing.T) {
	if extractMirrorKey(nil) != "" {
		t.Fatal("expected empty string for nil")
	}
	if extractMirrorKey(map[string]any{"key": map[string]any{"ECDSA_secp256k1": "mykey"}}) != "mykey" {
		t.Fatal("expected mykey")
	}
}

func TestExtractKeyCandidateArray(t *testing.T) {
	arr := []any{map[string]any{"key": "arraykey"}}
	if extractKeyCandidate(arr) != "arraykey" {
		t.Fatal("expected arraykey")
	}
}

func TestVerifyPetalAccount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/accounts/0.0.1" {
			w.Write([]byte(`{"key": {"key": "samekey"}}`))
		} else if r.URL.Path == "/api/v1/accounts/0.0.2" {
			w.Write([]byte(`{"key": {"key": "samekey"}}`))
		} else if r.URL.Path == "/api/v1/accounts/0.0.3" {
			w.Write([]byte(`{"key": {"key": "diffkey"}}`))
		} else if r.URL.Path == "/api/v1/accounts/0.0.4" {
			w.Write([]byte(`{"key": null}`))
		} else {
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.999",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ok, err := client.VerifyPetalAccount(context.Background(), "0.0.1", "0.0.2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected accounts to match")
	}

	ok, err = client.VerifyPetalAccount(context.Background(), "0.0.1", "0.0.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected accounts to not match")
	}

	ok, err = client.VerifyPetalAccount(context.Background(), "0.0.1", "0.0.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected empty key to not match")
	}
}

func TestVerifyPetalAccountMissingArgs(t *testing.T) {
	client := &Client{}
	_, err := client.VerifyPetalAccount(context.Background(), "", "0.0.2")
	if err == nil {
		t.Fatal("expected error for missing petal account")
	}
	_, err = client.VerifyPetalAccount(context.Background(), "0.0.1", "")
	if err == nil {
		t.Fatal("expected error for missing base account")
	}
}

func TestBuildBaseAccountCreateTxEmptyKey(t *testing.T) {
	_, err := BuildBaseAccountCreateTx(BaseAccountCreateTxParams{})
	if err == nil {
		t.Fatal("expected error for empty public key")
	}
}

func TestBuildPetalAccountCreateTxEmptyKey(t *testing.T) {
	_, err := BuildPetalAccountCreateTx(PetalAccountCreateTxParams{})
	if err == nil {
		t.Fatal("expected error for empty public key")
	}
}

func TestNewClientBadNetwork(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "invalid-network-name",
		OperatorAccountID:  "0.0.12345",
		OperatorPrivateKey: key.String(),
	})
	if err == nil {
		t.Fatal("expected error for invalid network")
	}
}

func TestCreateBaseAccountFailsExecution(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1", // invalid operator
		OperatorPrivateKey: key.String(),
	})
	
	// Close client to ensure fast precheck failure
	client.HederaClient().Close()

	maxAssoc := int32(10)
	_, err := client.CreateBaseAccount(context.Background(), BaseAccountCreateOptions{
		InitialBalanceHbar: -5, // Test defaulting to 10
		MaxAutomaticTokenAssociations: &maxAssoc,
		AccountMemo: "memo",
		TransactionMemo: "txmemo",
	})
	if err == nil {
		t.Fatal("expected error on execute from closed client")
	}
}

func TestCreatePetalAccountFailsExecution(t *testing.T) {
	operatorKey, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: operatorKey.String(),
	})
	
	client.HederaClient().Close()

	petalKey, _ := hedera.PrivateKeyGenerateEcdsa()
	maxAssoc := int32(10)
	_, err := client.CreatePetalAccount(context.Background(), PetalAccountCreateOptions{
		BasePrivateKey: petalKey.String(),
		InitialBalanceHbar: -1, // defaults to 1
		MaxAutomaticTokenAssociations: &maxAssoc,
		AccountMemo: "memo",
		TransactionMemo: "txmemo",
	})
	if err == nil {
		t.Fatal("expected error on execute from closed client")
	}
}

func TestCreatePetalAccountInvalidKey(t *testing.T) {
	operatorKey, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: operatorKey.String(),
	})

	_, err := client.CreatePetalAccount(context.Background(), PetalAccountCreateOptions{
		BasePrivateKey: "invalid-key",
	})
	if err == nil {
		t.Fatal("expected error for invalid petal private key")
	}
	
	_, err = client.CreatePetalAccount(context.Background(), PetalAccountCreateOptions{
		BasePrivateKey: "",
	})
	if err == nil {
		t.Fatal("expected error for empty petal private key")
	}
}
func TestVerifyPetalAccountMirrorError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.999",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})

	_, err := client.VerifyPetalAccount(context.Background(), "0.0.1", "0.0.2")
	if err == nil {
		t.Fatal("expected error from mirror node")
	}
}

func TestNewClientInvalidMirrorURL(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.12345",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      "http://[::1]:namedport",
	})
	if err == nil {
		t.Fatal("expected error for invalid mirror url")
	}
}
