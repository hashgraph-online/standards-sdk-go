package hcs21

type ErrorCode string

const (
	ErrorCodeSizeExceeded      ErrorCode = "size_exceeded"
	ErrorCodeInvalidPayload    ErrorCode = "invalid_payload"
	ErrorCodeMissingSignature  ErrorCode = "missing_signature"
	ErrorCodeVerificationFailed ErrorCode = "verification_failed"
)

type ValidationError struct {
	Code    ErrorCode
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

