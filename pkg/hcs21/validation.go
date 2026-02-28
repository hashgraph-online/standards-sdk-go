package hcs21

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateDeclaration performs the requested operation.
func ValidateDeclaration(declaration AdapterDeclaration) error {
	if strings.TrimSpace(declaration.P) != Protocol {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "declaration p must be hcs-21",
		}
	}
	switch declaration.Op {
	case OperationRegister, OperationUpdate, OperationDelete:
	default:
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: fmt.Sprintf("unsupported declaration op %q", declaration.Op),
		}
	}
	if strings.TrimSpace(declaration.AdapterID) == "" {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "adapter_id is required",
		}
	}
	if strings.TrimSpace(declaration.Entity) == "" {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "entity is required",
		}
	}
	if strings.TrimSpace(declaration.Package.Registry) == "" || strings.TrimSpace(declaration.Package.Name) == "" ||
		strings.TrimSpace(declaration.Package.Version) == "" || strings.TrimSpace(declaration.Package.Integrity) == "" {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "package registry/name/version/integrity are required",
		}
	}
	if !manifestPointerPattern.MatchString(strings.TrimSpace(declaration.Manifest)) {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "manifest must be immutable pointer",
		}
	}
	if declaration.Config == nil {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "config is required",
		}
	}
	configType, hasConfigType := declaration.Config["type"]
	if !hasConfigType || strings.TrimSpace(fmt.Sprintf("%v", configType)) == "" {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "config.type is required",
		}
	}

	bytes, err := json.Marshal(declaration)
	if err != nil {
		return &ValidationError{
			Code:    ErrorCodeInvalidPayload,
			Message: "failed to encode declaration",
		}
	}
	if len(bytes) > SafeMessageBytes {
		return &ValidationError{
			Code:    ErrorCodeSizeExceeded,
			Message: fmt.Sprintf("payload exceeds safe limit of %d bytes (%d)", SafeMessageBytes, len(bytes)),
		}
	}
	if len(bytes) > MaxMessageBytes {
		return &ValidationError{
			Code:    ErrorCodeSizeExceeded,
			Message: fmt.Sprintf("payload exceeds Hedera max %d bytes (%d)", MaxMessageBytes, len(bytes)),
		}
	}

	return nil
}
