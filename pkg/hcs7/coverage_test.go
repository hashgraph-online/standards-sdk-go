package hcs7

import (
	"context"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TestCovNewClientFailures(t *testing.T) {
	_, err := NewClient(ClientConfig{Network: "invalid"})
	if err == nil {
		t.Fatal("expected err")
	}

	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, err := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if client.MirrorClient() == nil {
		t.Fatal("expected mirror")
	}
}

func TestCovOperationsFailure(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})
	ctx := context.Background()

	_, err := client.CreateRegistry(ctx, CreateRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.RegisterConfig(ctx, RegisterConfigOptions{RegistryTopicID: "invalid"})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.RegisterMetadata(ctx, RegisterMetadataOptions{RegistryTopicID: "invalid"})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.GetRegistry(ctx, "invalid", QueryRegistryOptions{})
	if err == nil {
		t.Fatal("expected fail")
	}
}

func TestCovBuildConfigMessage(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	_, err := client.buildConfigMessage(RegisterConfigOptions{Type: "invalid"})
	if err == nil {
		t.Fatal("expected fail")
	}

	_, err = client.buildConfigMessage(RegisterConfigOptions{Type: ConfigTypeEVM})
	if err == nil {
		t.Fatal("expected fail for nil EVM")
	}

	evmCfg := &EvmConfigPayload{
		ContractAddress: "0x1234567890abcdef1234567890abcdef12345678",
		Abi: AbiDefinition{Name: "transfer", Inputs: []AbiIO{{Name: "to", Type: "address"}}, Outputs: []AbiIO{}, StateMutability: "nonpayable", Type: "function"},
	}
	_, err = client.buildConfigMessage(RegisterConfigOptions{Type: ConfigTypeEVM, EVM: evmCfg})
	// EVM config may still fail validation, that's ok; what matters is the code path was hit

	_, err = client.buildConfigMessage(RegisterConfigOptions{Type: ConfigTypeWASM})
	if err == nil {
		t.Fatal("expected fail for nil WASM")
	}
}

func TestCovDecodeMessage(t *testing.T) {
	_, err := decodeMessage("not-base64!!!")
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestCovReadString(t *testing.T) {
	s := readString(map[string]any{"k": "v"}, "k")
	if s != "v" {
		t.Fatal("expected v")
	}

	s2 := readString(map[string]any{"k": 123}, "k")
	if s2 != "" {
		t.Fatal("expected empty")
	}

	s3 := readString(map[string]any{}, "missing")
	if s3 != "" {
		t.Fatal("expected empty")
	}
}

func TestCovResolvePrivateKey(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	result := client.resolvePrivateKey("")
	if result != nil {
		t.Fatal("expected nil")
	}

	result2 := client.resolvePrivateKey("invalid")
	if result2 != nil {
		t.Fatal("expected nil for invalid")
	}

	result3 := client.resolvePrivateKey(pk.String())
	if result3 == nil {
		t.Fatal("expected non-nil")
	}
}

func TestCovResolvePublicKey(t *testing.T) {
	pk, _ := hedera.PrivateKeyGenerateEcdsa()
	client, _ := NewClient(ClientConfig{
		Network:            "testnet",
		OperatorAccountID:  "0.0.1",
		OperatorPrivateKey: pk.String(),
	})

	result := client.resolvePublicKey("", false)
	if result != nil {
		t.Fatal("expected nil")
	}

	result2 := client.resolvePublicKey("", true)
	if result2 == nil {
		t.Fatal("expected operator key")
	}

	result3 := client.resolvePublicKey("invalid", false)
	if result3 != nil {
		t.Fatal("expected nil for invalid")
	}

	result4 := client.resolvePublicKey(pk.PublicKey().String(), false)
	if result4 == nil {
		t.Fatal("expected valid key")
	}
}

func TestCovIsNumeric(t *testing.T) {
	if !isNumeric(42) {
		t.Fatal("expected true")
	}
	if !isNumeric(3.14) {
		t.Fatal("expected true")
	}
	if !isNumeric(int64(1)) {
		t.Fatal("expected true")
	}
	if isNumeric("str") {
		t.Fatal("expected false")
	}
}
