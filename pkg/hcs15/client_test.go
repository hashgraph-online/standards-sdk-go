package hcs15

import "testing"

func TestNormalizeEVMAddress(t *testing.T) {
	if normalizeEVMAddress("abc123") != "0xabc123" {
		t.Fatalf("expected missing prefix to be added")
	}
	if normalizeEVMAddress("0xabc123") != "0xabc123" {
		t.Fatalf("expected existing prefix to remain unchanged")
	}
}

func TestExtractMirrorKey(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  string
	}{
		{
			name: "direct key",
			input: map[string]any{
				"key": "302e020100300506032b657004220420abc",
			},
			want: "302e020100300506032b657004220420abc",
		},
		{
			name: "nested threshold key",
			input: map[string]any{
				"thresholdKey": map[string]any{
					"keys": []any{
						map[string]any{
							"ecdsa_secp256k1": "ignored",
						},
						map[string]any{
							"key": map[string]any{
								"ECDSA_secp256k1": "302e020100300506032b657004220420def",
							},
						},
					},
				},
			},
			want: "302e020100300506032b657004220420def",
		},
		{
			name:  "empty",
			input: map[string]any{},
			want:  "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := extractMirrorKey(testCase.input)
			if got != testCase.want {
				t.Fatalf("extractMirrorKey mismatch: got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestCreatePetalAccount_RequiresBasePrivateKey(t *testing.T) {
	client := &Client{}
	_, err := client.CreatePetalAccount(t.Context(), PetalAccountCreateOptions{})
	if err == nil {
		t.Fatalf("expected error when base private key is missing")
	}
}
