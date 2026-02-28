package hcs17

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	hedera "github.com/hashgraph/hedera-sdk-go/v2"
)

func TestCalculateAccountStateHash(t *testing.T) {
	client := &Client{}

	result, err := client.CalculateAccountStateHash(AccountStateInput{
		AccountID: "0.0.12345",
		PublicKey: "FGHKLJHDGK",
		Topics: []TopicState{
			{TopicID: "0.0.67890", LatestRunningHash: "efgh5678"},
			{TopicID: "0.0.12345", LatestRunningHash: "abcd1234"},
		},
	})
	if err != nil {
		t.Fatalf("CalculateAccountStateHash failed: %v", err)
	}

	concatenated := "0.0.12345abcd12340.0.67890efgh5678FGHKLJHDGK"
	expectedHash := sha512.Sum384([]byte(concatenated))
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	if result.StateHash != expectedHashHex {
		t.Fatalf("state hash mismatch: got %s want %s", result.StateHash, expectedHashHex)
	}
}

func TestCalculateCompositeStateHash(t *testing.T) {
	client := &Client{}

	result, err := client.CalculateCompositeStateHash(CompositeStateInput{
		CompositeAccountID:            "0.0.777",
		CompositePublicKeyFingerprint: "0xffff",
		MemberStates: []CompositeMemberState{
			{AccountID: "0.0.222", StateHash: "0xbbb"},
			{AccountID: "0.0.111", StateHash: "0xaaa"},
		},
		CompositeTopics: []TopicState{
			{TopicID: "0.0.444", LatestRunningHash: "0xddd"},
			{TopicID: "0.0.333", LatestRunningHash: "0xccc"},
		},
	})
	if err != nil {
		t.Fatalf("CalculateCompositeStateHash failed: %v", err)
	}

	concatenated := "0.0.1110xaaa0.0.2220xbbb0.0.3330xccc0.0.4440xddd0xffff"
	expectedHash := sha512.Sum384([]byte(concatenated))
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	if result.StateHash != expectedHashHex {
		t.Fatalf("composite state hash mismatch: got %s want %s", result.StateHash, expectedHashHex)
	}
}

func TestCalculateKeyFingerprint(t *testing.T) {
	client := &Client{}

	keyOne, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate key one: %v", err)
	}
	keyTwo, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate key two: %v", err)
	}
	keyThree, err := hedera.PrivateKeyGenerateEcdsa()
	if err != nil {
		t.Fatalf("failed to generate key three: %v", err)
	}

	fingerprintOne, err := client.CalculateKeyFingerprint(
		[]hedera.PublicKey{keyThree.PublicKey(), keyOne.PublicKey(), keyTwo.PublicKey()},
		2,
	)
	if err != nil {
		t.Fatalf("CalculateKeyFingerprint failed: %v", err)
	}
	fingerprintTwo, err := client.CalculateKeyFingerprint(
		[]hedera.PublicKey{keyTwo.PublicKey(), keyThree.PublicKey(), keyOne.PublicKey()},
		2,
	)
	if err != nil {
		t.Fatalf("CalculateKeyFingerprint failed: %v", err)
	}

	if fingerprintOne != fingerprintTwo {
		t.Fatalf("expected deterministic fingerprint regardless of key order")
	}
}

func TestCreateStateHashMessage(t *testing.T) {
	client := &Client{}

	message := client.CreateStateHashMessage(
		"0x9a1cfb",
		"0.0.123456",
		[]string{"0.0.topic1", "0.0.topic2"},
		"sync",
		nil,
	)

	if message.Protocol != "hcs-17" {
		t.Fatalf("unexpected protocol: %s", message.Protocol)
	}
	if message.Operation != "state_hash" {
		t.Fatalf("unexpected operation: %s", message.Operation)
	}
	if message.StateHash != "0x9a1cfb" {
		t.Fatalf("unexpected state hash: %s", message.StateHash)
	}
	if message.Timestamp == "" {
		t.Fatalf("expected timestamp to be populated")
	}
}
