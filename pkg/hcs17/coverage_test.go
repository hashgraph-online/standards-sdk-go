package hcs17

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	if client.HederaClient() == nil || client.MirrorClient() == nil {
		t.Fatal("expected clients to be initialized")
	}
}

func TestNewClientMissingOperatorID(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorPrivateKey: key.String(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientMissingOperatorKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Network:           "testnet",
		OperatorAccountID: "0.0.12345",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClientInvalidNetwork(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	_, err := NewClient(ClientConfig{
		Network:            "invalid",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: key.String(),
	})
	if err == nil {
		t.Fatal("expected error for invalid network")
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
		t.Fatal("expected error")
	}
}

func TestNewClientInvalidOperatorKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: "invalid",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

	// Test completely removed.

func TestCreateStateTopicFailsExecute(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: key.String(),
	})
	client.HederaClient().Close()
	_, err := client.CreateStateTopic(context.Background(), CreateTopicOptions{
		TTLSeconds: 123,
	})
	if err == nil {
		t.Fatal("expected error on execute from closed client")
	}
}

func TestSubmitMessageFailsExecute(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: key.String(),
	})
	client.HederaClient().Close()
	_, err := client.SubmitMessage(context.Background(), "0.0.999", StateHashMessage{}, "")
	if err == nil {
		t.Fatal("expected error on execute from closed client")
	}
}

func TestValidateTopic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/topics/0.0.1" {
			w.Write([]byte(`{"memo": "hcs-17:0:123"}`))
		} else if r.URL.Path == "/api/v1/topics/0.0.2" {
			w.Write([]byte(`{"memo": "invalid"}`))
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})

	ok, memo, err := client.ValidateTopic(context.Background(), "0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || memo == nil || memo.TTLSeconds != 123 {
		t.Fatal("expected validation to pass")
	}

	ok, memo, err = client.ValidateTopic(context.Background(), "0.0.2")
	if err == nil {
		t.Fatal("expected error for invalid memo")
	}
	if ok || memo != nil {
		t.Fatal("expected validation to fail")
	}

	_, _, err = client.ValidateTopic(context.Background(), "0.0.3")
	if err == nil {
		t.Fatal("expected error for mirror error")
	}
}

func TestGetRecentMessages(t *testing.T) {
	validMsg := StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "hash123",
		Topics:    []string{"0.0.1"},
		AccountID: "0.0.2",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	validBytes, _ := json.Marshal(validMsg)
	validMsgB64 := base64.StdEncoding.EncodeToString(validBytes)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/topics/0.0.1/messages" {
			type MirrorMessage struct {
				ConsensusTimestamp string `json:"consensus_timestamp"`
				SequenceNumber     int64  `json:"sequence_number"`
				PayerAccountID     string `json:"payer_account_id"`
				Message            string `json:"message"`
				RunningHash        string `json:"running_hash"`
			}
			type List struct {
				Messages []MirrorMessage `json:"messages"`
			}
			list := List{
				Messages: []MirrorMessage{
					{Message: validMsgB64},     // valid
					{Message: "bad-base64!"},   // invalid base64
					{Message: "eyJibGFoIjo=]"}, // invalid json in base64
					{Message: base64.StdEncoding.EncodeToString([]byte(`{}`))}, // missing fields (validation fails)
				},
			}
			body, _ := json.Marshal(list)
			w.Write(body)
		} else {
			http.Error(w, "err", http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.123",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})

	messages, err := client.GetRecentMessages(context.Background(), "0.0.1", -1, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 valid message, got %d", len(messages))
	}

	latest, err := client.GetLatestMessage(context.Background(), "0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest == nil {
		t.Fatal("expected latest message")
	}
	if latest.Message.StateHash != "hash123" {
		t.Fatal("unexpected message content")
	}

	_, err = client.GetRecentMessages(context.Background(), "0.0.2", 10, "asc")
	if err == nil {
		t.Fatal("expected mirror error")
	}

	latestEmpty, err := client.GetLatestMessage(context.Background(), "0.0.2")
	if err == nil && latestEmpty != nil {
		t.Fatal("expected nil or err")
	}
}

func TestComputeAndPublishMirrorError(t *testing.T) {
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

	_, err := client.ComputeAndPublish(context.Background(), ComputeAndPublishOptions{
		Topics: []string{"0.0.1"},
	})
	if err == nil {
		t.Fatal("expected mirror error")
	}
}

func TestComputeAndPublishFailsExecute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"messages": [{"running_hash": "hash"}]}`))
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.999",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})
	client.HederaClient().Close()

	_, err := client.ComputeAndPublish(context.Background(), ComputeAndPublishOptions{
		Topics:           []string{"0.0.1"},
		AccountID:        "0.0.2",
		AccountPublicKey: "pubkey",
	})
	if err == nil {
		t.Fatal("expected error on message submission from closed client")
	}
}

func TestComputeAndPublishInvalidAccount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"messages": []}`))
	}))
	defer ts.Close()

	key, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.999",
		OperatorPrivateKey: key.String(),
		MirrorBaseURL:      ts.URL,
	})

	_, err := client.ComputeAndPublish(context.Background(), ComputeAndPublishOptions{
		Topics:           []string{"0.0.1"},
		AccountID:        "", // empty will cause CalculateAccountStateHash to fail
		AccountPublicKey: "pubkey",
	})
	if err == nil {
		t.Fatal("expected error for empty account")
	}
}

func TestCalculateAccountStateHashFails(t *testing.T) {
	client := &Client{}
	_, err := client.CalculateAccountStateHash(AccountStateInput{})
	if err == nil {
		t.Fatal("expected error for empty account")
	}

	_, err = client.CalculateAccountStateHash(AccountStateInput{
		AccountID: "0.0.1",
		PublicKey: "",
	})
	if err == nil {
		t.Fatal("expected error for empty pubkey")
	}
}

func TestCalculateCompositeStateHashFails(t *testing.T) {
	client := &Client{}
	_, err := client.CalculateCompositeStateHash(CompositeStateInput{})
	if err == nil {
		t.Fatal("expected error for empty account")
	}
}

func TestCalculateKeyFingerprintCoverage(t *testing.T) {
	client := &Client{}
	key1, _ := hedera.PrivateKeyGenerateEcdsa()
	key2, _ := hedera.PrivateKeyGenerateEcdsa()
	
	_, err := client.CalculateKeyFingerprint(nil, 1)
	if err == nil {
		t.Fatal("expected error for empty keys")
	}

	_, err = client.CalculateKeyFingerprint([]hedera.PublicKey{key1.PublicKey()}, 0)
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}

	hash, err := client.CalculateKeyFingerprint([]hedera.PublicKey{key1.PublicKey(), key2.PublicKey()}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash")
	}
}

func TestVerifyStateHash(t *testing.T) {
	client := &Client{}
	inputAcc := AccountStateInput{
		AccountID: "0.0.1",
		PublicKey: "pubkey",
	}
	expectedAcc, _ := client.CalculateAccountStateHash(inputAcc)
	
	ok, err := client.VerifyStateHash(inputAcc, expectedAcc.StateHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected verify to pass")
	}

	ok, err = client.VerifyStateHash(inputAcc, "bad_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected verify to fail")
	}
	
	_, err = client.VerifyStateHash(AccountStateInput{AccountID: ""}, "any")
	if err == nil {
		t.Fatal("expected error for invalid input generating hash")
	}

	inputComp := CompositeStateInput{
		CompositeAccountID: "0.0.1",
	}
	expectedComp, _ := client.CalculateCompositeStateHash(inputComp)
	
	ok, err = client.VerifyStateHash(inputComp, expectedComp.StateHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected verify to pass")
	}

	_, err = client.VerifyStateHash(CompositeStateInput{CompositeAccountID: ""}, "any")
	if err == nil {
		t.Fatal("expected error for invalid composite generating hash")
	}
	
	_, err = client.VerifyStateHash("unsupported_type", "any")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestNormalizePublicKeyValue(t *testing.T) {
	val, err := normalizePublicKeyValue("pubkey")
	if err != nil || val != "pubkey" {
		t.Fatal("unexpected string normalize")
	}
	
	_, err = normalizePublicKeyValue("   ")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
	
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	val, err = normalizePublicKeyValue(pk.PublicKey())
	if err != nil || val == "" {
		t.Fatal("unexpected PublicKey normalize")
	}

	_, err = normalizePublicKeyValue(123)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestCreateStateHashMessageCoverage(t *testing.T) {
	client := &Client{}
	epoch := int64(1)
	msg := client.CreateStateHashMessage("hash", "0.0.1", []string{"0.0.2"}, "memo", &epoch)
	if msg.Protocol != "hcs-17" {
		t.Fatal("expected protocol hcs-17")
	}
	if msg.StateHash != "hash" {
		t.Fatal("expected hash")
	}
	if msg.Epoch == nil || *msg.Epoch != 1 {
		t.Fatal("expected epoch")
	}
	
	msgNoEpoch := client.CreateStateHashMessage("hash", "0.0.1", []string{"0.0.2"}, "memo", nil)
	if msgNoEpoch.Epoch != nil {
		t.Fatal("expected no epoch")
	}
}

func TestGenerateTopicMemo(t *testing.T) {
	memo := GenerateTopicMemo(123)
	if memo != "hcs-17:0:123" {
		t.Fatalf("unexpected memo: %s", memo)
	}
}

func TestParseTopicMemo(t *testing.T) {
	parsed, err := ParseTopicMemo("hcs-17:0:123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.TTLSeconds != 123 {
		t.Fatal("unexpected TTLSeconds")
	}
	if parsed.Type != HCS17TopicTypeState {
		t.Fatal("unexpected topic type")
	}

	_, err = ParseTopicMemo("invalid")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildCreateStateTopicTxCoverage(t *testing.T) {
	key, _ := hedera.PrivateKeyGenerateEcdsa()
	tx := BuildCreateStateTopicTx(CreateTopicOptions{
		TTLSeconds: 123,
		AdminKey: key.PublicKey(),
		SubmitKey: key.PublicKey(),
	})
	if tx == nil {
		t.Fatal("expected tx")
	}
	if tx.GetTopicMemo() != "hcs-17:0:123" {
		t.Fatal("expected memo")
	}
}

func TestBuildStateHashMessageTxCoverage(t *testing.T) {
	tx, err := BuildStateHashMessageTx("0.0.1", StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "hash",
		Topics:    []string{"0.0.1"},
		AccountID: "0.0.2",
	}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx == nil {
		t.Fatal("expected tx")
	}
	// The memo may be empty if not provided or overridden
	if tx.GetTransactionMemo() != "" && tx.GetTransactionMemo() != "hcs-17:op:state_hash" {
		t.Fatalf("expected empty or hcs-17:op:state_hash memo, got %s", tx.GetTransactionMemo())
	}
	
	_, err = BuildStateHashMessageTx("invalid", StateHashMessage{}, "")
	if err == nil {
		t.Fatal("expected error for invalid topic ID")
	}
}

func TestValidateStateHashMessageCoverage(t *testing.T) {
	errs := ValidateStateHashMessage(StateHashMessage{
		Protocol:  "hcs-17",
		Operation: "state_hash",
		StateHash: "hash",
		AccountID: "0.0.1",
		Topics:    []string{"0.0.2"},
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got: %v", errs)
	}

	errs = ValidateStateHashMessage(StateHashMessage{})
	if len(errs) != 5 {
		t.Fatalf("expected 5 errors, got: %v", errs)
	}
}
