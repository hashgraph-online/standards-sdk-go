package hcs7

import "testing"

func TestValidateEVMMessage(t *testing.T) {
	err := ValidateMessage(Message{
		P:    "hcs-7",
		Op:   OperationRegisterConfig,
		Type: ConfigTypeEVM,
		Config: EvmConfigPayload{
			ContractAddress: "0x1111111111111111111111111111111111111111",
			Abi: AbiDefinition{
				Name:            "resolve",
				Inputs:          []AbiIO{{Type: "string"}},
				Outputs:         []AbiIO{{Type: "string"}},
				StateMutability: "view",
				Type:            "function",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

func TestValidateMetadataMessage(t *testing.T) {
	err := ValidateMessage(Message{
		P:       "hcs-7",
		Op:      OperationRegister,
		TopicID: "0.0.1001",
		Data: map[string]any{
			"weight": 1,
			"tags":   []string{"validator"},
		},
	})
	if err != nil {
		t.Fatalf("expected valid message, got %v", err)
	}
}

