package hcs12

import "testing"

func TestValidatePayloadRegister(t *testing.T) {
	err := ValidatePayload(map[string]any{
		"p":    "hcs-12",
		"op":   "register",
		"name": "demo",
	})
	if err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
}

func TestValidatePayloadAddBlock(t *testing.T) {
	err := ValidatePayload(map[string]any{
		"p":          "hcs-12",
		"op":         "add-block",
		"block_t_id": "0.0.1234",
	})
	if err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
}
