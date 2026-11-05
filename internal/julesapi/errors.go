package julesapi

import "fmt"

// Exit codes conforming to project specifications:
// 0: success
// 1: general error
// 2: usage error
// 3: auth error
// 4: not found
// 5: API/network error after retries
const (
	ExitSuccess      = 0
	ExitGeneralError = 1
	ExitUsageError   = 2
	ExitAuthError    = 3
	ExitNotFound     = 4
	ExitAPIError     = 5
)

// APIError represents an error returned by the Google Jules API.
type APIError struct {
	StatusCode int
	Message    string
	ExitCode   int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// NewAPIError converts an HTTP status code and message to an APIError with appropriate exit code.
func NewAPIError(statusCode int, message string) *APIError {
	exitCode := ExitAPIError
	switch statusCode {
	case 401, 403:
		exitCode = ExitAuthError
	case 404:
		exitCode = ExitNotFound
	}
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		ExitCode:   exitCode,
	}
}
