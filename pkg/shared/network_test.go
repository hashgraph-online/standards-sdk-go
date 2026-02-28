package shared

import (
	"testing"
)

func TestNormalizeNetworkMainnet(t *testing.T) {
	result, err := NormalizeNetwork("mainnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NetworkMainnet {
		t.Fatalf("expected %q, got %q", NetworkMainnet, result)
	}
}

func TestNormalizeNetworkTestnet(t *testing.T) {
	result, err := NormalizeNetwork("testnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NetworkTestnet {
		t.Fatalf("expected %q, got %q", NetworkTestnet, result)
	}
}

func TestNormalizeNetworkCaseInsensitive(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"MAINNET", NetworkMainnet},
		{"Mainnet", NetworkMainnet},
		{"TESTNET", NetworkTestnet},
		{"Testnet", NetworkTestnet},
		{"  mainnet  ", NetworkMainnet},
		{"  testnet  ", NetworkTestnet},
	}

	for _, tc := range cases {
		result, err := NormalizeNetwork(tc.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Fatalf("expected %q for input %q, got %q", tc.expected, tc.input, result)
		}
	}
}

func TestNormalizeNetworkEmpty(t *testing.T) {
	result, err := NormalizeNetwork("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NetworkTestnet {
		t.Fatalf("expected %q for empty input, got %q", NetworkTestnet, result)
	}
}

func TestNormalizeNetworkWhitespaceOnly(t *testing.T) {
	result, err := NormalizeNetwork("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != NetworkTestnet {
		t.Fatalf("expected %q for whitespace input, got %q", NetworkTestnet, result)
	}
}

func TestNormalizeNetworkUnsupported(t *testing.T) {
	_, err := NormalizeNetwork("devnet")
	if err == nil {
		t.Fatal("expected error for unsupported network")
	}
}

func TestNewHederaClientMainnet(t *testing.T) {
	client, err := NewHederaClient("mainnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewHederaClientTestnet(t *testing.T) {
	client, err := NewHederaClient("testnet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewHederaClientUnsupported(t *testing.T) {
	_, err := NewHederaClient("badnet")
	if err == nil {
		t.Fatal("expected error for unsupported network")
	}
}
