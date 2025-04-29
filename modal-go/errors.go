package modal

// errors.go defines common error types for the public API.

// TimeoutError is returned when a function execution exceeds the allowed time limit.
type TimeoutError struct {
	Exception string
}

func (e TimeoutError) Error() string {
	return "TimeoutError: " + e.Exception
}

// RemoteError represents an error on the Modal server, or a Python exception.
type RemoteError struct {
	Exception string
}

func (e RemoteError) Error() string {
	return "RemoteError: " + e.Exception
}

// InternalFailure is a retryable internal error from Modal.
type InternalFailure struct {
	Exception string
}

func (e InternalFailure) Error() string {
	return "InternalFailure: " + e.Exception
}
