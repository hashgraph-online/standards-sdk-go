package hcs20

import "fmt"

type HCS20Error struct {
	Message string
}

func (errorValue HCS20Error) Error() string {
	return errorValue.Message
}

type PointsDeploymentError struct {
	HCS20Error
	Tick string
}

type PointsMintError struct {
	HCS20Error
	Tick            string
	RequestedAmount string
	AvailableSupply string
}

type PointsTransferError struct {
	HCS20Error
	Tick             string
	From             string
	To               string
	Amount           string
	AvailableBalance string
}

type PointsBurnError struct {
	HCS20Error
	Tick             string
	From             string
	Amount           string
	AvailableBalance string
}

type PointsValidationError struct {
	HCS20Error
	ValidationErrors []string
}

func NewPointsValidationError(message string, validationErrors []string) error {
	if len(validationErrors) == 0 {
		return PointsValidationError{
			HCS20Error: HCS20Error{Message: message},
		}
	}
	return PointsValidationError{
		HCS20Error:       HCS20Error{Message: fmt.Sprintf("%s: %v", message, validationErrors)},
		ValidationErrors: append([]string{}, validationErrors...),
	}
}

type PointsNotFoundError struct {
	HCS20Error
	Tick string
}

func NewPointsNotFoundError(tick string) error {
	return PointsNotFoundError{
		HCS20Error: HCS20Error{Message: fmt.Sprintf("points with tick %q not found", tick)},
		Tick:       tick,
	}
}

type TopicRegistrationError struct {
	HCS20Error
	TopicID string
}

type InvalidMessageFormatError struct {
	HCS20Error
}

type InvalidAccountFormatError struct {
	HCS20Error
	Account string
}

func NewInvalidAccountFormatError(account string) error {
	return InvalidAccountFormatError{
		HCS20Error: HCS20Error{Message: fmt.Sprintf("invalid Hedera account format: %s", account)},
		Account:    account,
	}
}

type InvalidTickFormatError struct {
	HCS20Error
	Tick string
}

func NewInvalidTickFormatError(tick string) error {
	return InvalidTickFormatError{
		HCS20Error: HCS20Error{Message: fmt.Sprintf("invalid tick format: %s", tick)},
		Tick:       tick,
	}
}

type InvalidNumberFormatError struct {
	HCS20Error
	Field string
	Value string
}

func NewInvalidNumberFormatError(field string, value string) error {
	return InvalidNumberFormatError{
		HCS20Error: HCS20Error{Message: fmt.Sprintf("invalid number format for %s: %s", field, value)},
		Field:      field,
		Value:      value,
	}
}
