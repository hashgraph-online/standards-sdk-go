package hcs20

import (
	"encoding/json"
	"fmt"
	"strings"
)

// NormalizeTick returns the normalized tick.
func NormalizeTick(tick string) string {
	return strings.ToLower(strings.TrimSpace(tick))
}

// NormalizeAccountID returns a normalized account/topic identifier.
func NormalizeAccountID(identifier string) (string, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return "", NewInvalidAccountFormatError(identifier)
	}

	parts := strings.SplitN(trimmed, "-", 2)
	base := parts[0]
	if !hederaEntityRegex.MatchString(base) {
		return "", NewInvalidAccountFormatError(identifier)
	}

	if len(parts) == 2 {
		suffix := parts[1]
		if len(suffix) != 5 {
			return "", NewInvalidAccountFormatError(identifier)
		}
		for _, character := range suffix {
			if character < 'a' || character > 'z' {
				return "", NewInvalidAccountFormatError(identifier)
			}
		}
	}

	return base, nil
}

// ValidateNumberString validates number-string fields for HCS-20.
func ValidateNumberString(field string, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return NewInvalidNumberFormatError(field, value)
	}
	if len(trimmed) > MaxNumberLength {
		return NewInvalidNumberFormatError(field, value)
	}
	if !numberRegex.MatchString(trimmed) {
		return NewInvalidNumberFormatError(field, value)
	}
	return nil
}

// NormalizeMessage normalizes message fields for deterministic validation.
func NormalizeMessage(message Message) (Message, error) {
	normalized := message
	normalized.Protocol = strings.ToLower(strings.TrimSpace(message.Protocol))
	normalized.Operation = strings.ToLower(strings.TrimSpace(message.Operation))
	normalized.Name = strings.TrimSpace(message.Name)
	normalized.Tick = NormalizeTick(message.Tick)
	normalized.Max = strings.TrimSpace(message.Max)
	normalized.Limit = strings.TrimSpace(message.Limit)
	normalized.Metadata = strings.TrimSpace(message.Metadata)
	normalized.Memo = strings.TrimSpace(message.Memo)
	normalized.Amount = strings.TrimSpace(message.Amount)

	if message.To != "" {
		accountID, err := NormalizeAccountID(message.To)
		if err != nil {
			return normalized, err
		}
		normalized.To = accountID
	}
	if message.From != "" {
		accountID, err := NormalizeAccountID(message.From)
		if err != nil {
			return normalized, err
		}
		normalized.From = accountID
	}
	if message.TopicID != "" {
		topicID, err := NormalizeAccountID(message.TopicID)
		if err != nil {
			return normalized, err
		}
		normalized.TopicID = topicID
	}

	return normalized, nil
}

// ValidateMessage validates an HCS-20 message.
func ValidateMessage(message Message) error {
	normalized, err := NormalizeMessage(message)
	if err != nil {
		return err
	}

	if normalized.Protocol != ProtocolID {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: "p must be hcs-20"}}
	}

	if len(normalized.Memo) > MaxMemoLength {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: fmt.Sprintf("m must be <= %d characters", MaxMemoLength)}}
	}

	switch normalized.Operation {
	case OperationDeploy:
		return validateDeploy(normalized)
	case OperationMint:
		return validateMint(normalized)
	case OperationBurn:
		return validateBurn(normalized)
	case OperationTransfer:
		return validateTransfer(normalized)
	case OperationRegister:
		return validateRegister(normalized)
	default:
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: "op must be one of deploy|mint|burn|transfer|register"}}
	}
}

func validateDeploy(message Message) error {
	if message.Name == "" || len(message.Name) > MaxNameLength {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: fmt.Sprintf("name is required and must be <= %d characters", MaxNameLength)}}
	}
	if message.Tick == "" {
		return NewInvalidTickFormatError(message.Tick)
	}
	if err := ValidateNumberString("max", message.Max); err != nil {
		return err
	}
	if message.Limit != "" {
		if err := ValidateNumberString("lim", message.Limit); err != nil {
			return err
		}
	}
	if len(message.Metadata) > MaxMetadataLength {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: fmt.Sprintf("metadata must be <= %d characters", MaxMetadataLength)}}
	}
	return nil
}

func validateMint(message Message) error {
	if message.Tick == "" {
		return NewInvalidTickFormatError(message.Tick)
	}
	if err := ValidateNumberString("amt", message.Amount); err != nil {
		return err
	}
	if message.To == "" {
		return NewInvalidAccountFormatError(message.To)
	}
	return nil
}

func validateBurn(message Message) error {
	if message.Tick == "" {
		return NewInvalidTickFormatError(message.Tick)
	}
	if err := ValidateNumberString("amt", message.Amount); err != nil {
		return err
	}
	if message.From == "" {
		return NewInvalidAccountFormatError(message.From)
	}
	return nil
}

func validateTransfer(message Message) error {
	if message.Tick == "" {
		return NewInvalidTickFormatError(message.Tick)
	}
	if err := ValidateNumberString("amt", message.Amount); err != nil {
		return err
	}
	if message.From == "" {
		return NewInvalidAccountFormatError(message.From)
	}
	if message.To == "" {
		return NewInvalidAccountFormatError(message.To)
	}
	return nil
}

func validateRegister(message Message) error {
	if message.Name == "" || len(message.Name) > MaxNameLength {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: fmt.Sprintf("name is required and must be <= %d characters", MaxNameLength)}}
	}
	if len(message.Metadata) > MaxMetadataLength {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: fmt.Sprintf("metadata must be <= %d characters", MaxMetadataLength)}}
	}
	if message.Private == nil {
		return InvalidMessageFormatError{HCS20Error: HCS20Error{Message: "private is required for register"}}
	}
	if message.TopicID == "" {
		return NewInvalidAccountFormatError(message.TopicID)
	}
	return nil
}

// ParseMessageBytes decodes and validates an HCS-20 message payload.
func ParseMessageBytes(payload []byte) (Message, error) {
	var message Message
	if err := json.Unmarshal(payload, &message); err != nil {
		return Message{}, fmt.Errorf("failed to decode HCS-20 message: %w", err)
	}

	normalized, err := NormalizeMessage(message)
	if err != nil {
		return Message{}, err
	}
	if err := ValidateMessage(normalized); err != nil {
		return Message{}, err
	}

	return normalized, nil
}

// BuildMessagePayload validates and serializes an HCS-20 message.
func BuildMessagePayload(message Message) ([]byte, Message, error) {
	normalized, err := NormalizeMessage(message)
	if err != nil {
		return nil, Message{}, err
	}
	if err := ValidateMessage(normalized); err != nil {
		return nil, Message{}, err
	}

	payload, err := json.Marshal(normalized)
	if err != nil {
		return nil, Message{}, fmt.Errorf("failed to marshal HCS-20 message: %w", err)
	}

	return payload, normalized, nil
}
