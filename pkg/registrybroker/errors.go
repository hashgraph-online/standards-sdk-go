package registrybroker

import "fmt"

type RegistryBrokerError struct {
	Message    string
	Status     int
	StatusText string
	Body       any
}

// Error performs the requested operation.
func (e *RegistryBrokerError) Error() string {
	if e == nil {
		return "registry broker request failed"
	}
	if e.Status > 0 {
		return fmt.Sprintf("%s (status=%d %s)", e.Message, e.Status, e.StatusText)
	}
	return e.Message
}

type RegistryBrokerParseError struct {
	Message string
	Body    string
	Cause   error
}

// Error performs the requested operation.
func (e *RegistryBrokerParseError) Error() string {
	if e == nil {
		return "registry broker parse error"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}
