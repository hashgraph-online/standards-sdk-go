package registrybroker

import (
	"testing"
)

func TestEncryptionRoundTrip(t *testing.T) {
	t.Parallel()

	client, err := NewRegistryBrokerClient(RegistryBrokerClientOptions{})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	alice, err := client.GenerateEphemeralKeyPair()
	if err != nil {
		t.Fatalf("failed to generate alice key pair: %v", err)
	}
	bob, err := client.GenerateEphemeralKeyPair()
	if err != nil {
		t.Fatalf("failed to generate bob key pair: %v", err)
	}

	aliceSecret, err := client.DeriveSharedSecret(DeriveSharedSecretOptions{
		PrivateKey:    alice.PrivateKey,
		PeerPublicKey: bob.PublicKey,
	})
	if err != nil {
		t.Fatalf("failed to derive alice secret: %v", err)
	}
	bobSecret, err := client.DeriveSharedSecret(DeriveSharedSecretOptions{
		PrivateKey:    bob.PrivateKey,
		PeerPublicKey: alice.PublicKey,
	})
	if err != nil {
		t.Fatalf("failed to derive bob secret: %v", err)
	}

	if len(aliceSecret) != len(bobSecret) {
		t.Fatalf("shared secret lengths mismatch: %d != %d", len(aliceSecret), len(bobSecret))
	}
	for index := range aliceSecret {
		if aliceSecret[index] != bobSecret[index] {
			t.Fatalf("shared secret mismatch at index %d", index)
		}
	}

	envelope, err := client.BuildCipherEnvelope(EncryptCipherEnvelopeOptions{
		Plaintext:    "hello registry broker",
		SessionID:    "session-1",
		SharedSecret: aliceSecret,
		Recipients: []CipherEnvelopeRecipient{
			{UAID: "uaid:aid:test"},
		},
	})
	if err != nil {
		t.Fatalf("failed to encrypt envelope: %v", err)
	}

	plaintext, err := client.OpenCipherEnvelope(DecryptCipherEnvelopeOptions{
		Envelope:     envelope,
		SharedSecret: bobSecret,
	})
	if err != nil {
		t.Fatalf("failed to decrypt envelope: %v", err)
	}
	if plaintext != "hello registry broker" {
		t.Fatalf("unexpected plaintext %s", plaintext)
	}
}
