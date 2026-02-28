package hcs7

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	hederaTopicIDPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	evmAddressPattern    = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
)

// ValidateMessage performs the requested operation.
func ValidateMessage(message Message) error {
	if strings.TrimSpace(message.P) != "hcs-7" {
		return fmt.Errorf("message p must be hcs-7")
	}
	if len(strings.TrimSpace(message.Memo)) > 500 {
		return fmt.Errorf("message memo exceeds 500 characters")
	}

	switch message.Op {
	case OperationRegisterConfig:
		if message.Type != ConfigTypeEVM && message.Type != ConfigTypeWASM {
			return fmt.Errorf("register-config requires t to be evm or wasm")
		}
		if message.Type == ConfigTypeEVM {
			payload, ok := message.Config.(EvmConfigPayload)
			if !ok {
				return fmt.Errorf("evm config payload is invalid")
			}
			if !evmAddressPattern.MatchString(strings.TrimSpace(payload.ContractAddress)) {
				return fmt.Errorf("evm contractAddress must be a 0x-prefixed 40-byte address")
			}
			if strings.TrimSpace(payload.Abi.Name) == "" {
				return fmt.Errorf("evm abi.name is required")
			}
		}
		if message.Type == ConfigTypeWASM {
			payload, ok := message.Config.(WasmConfigPayload)
			if !ok {
				return fmt.Errorf("wasm config payload is invalid")
			}
			if !hederaTopicIDPattern.MatchString(strings.TrimSpace(payload.WasmTopicID)) {
				return fmt.Errorf("wasm wasmTopicId must be a Hedera topic ID")
			}
			if payload.OutputType.Type != "string" || payload.OutputType.Format != "topic-id" {
				return fmt.Errorf("wasm outputType must be {type:string, format:topic-id}")
			}
		}
	case OperationRegister:
		if !hederaTopicIDPattern.MatchString(strings.TrimSpace(message.TopicID)) {
			return fmt.Errorf("register requires valid t_id")
		}
		data := message.Data
		if data == nil {
			return fmt.Errorf("register requires d object")
		}
		weightValue, hasWeight := data["weight"]
		if !hasWeight {
			return fmt.Errorf("register requires d.weight")
		}
		if !isNumeric(weightValue) {
			return fmt.Errorf("register d.weight must be numeric")
		}
		tagsValue, hasTags := data["tags"]
		if !hasTags {
			return fmt.Errorf("register requires d.tags")
		}
		if !isStringSlice(tagsValue) {
			return fmt.Errorf("register d.tags must be an array of strings")
		}
	default:
		return fmt.Errorf("unsupported operation %q", message.Op)
	}

	return nil
}

func isNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func isStringSlice(value any) bool {
	switch typed := value.(type) {
	case []string:
		return len(typed) > 0
	case []any:
		if len(typed) == 0 {
			return false
		}
		for _, item := range typed {
			if _, ok := item.(string); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

