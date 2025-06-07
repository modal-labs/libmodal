package modal

// errors.go defines common error types for the public API.

// FunctionTimeoutError is returned when a function execution exceeds the allowed time limit.
type FunctionTimeoutError struct {
	Exception string
}

func (e FunctionTimeoutError) Error() string {
	return "FunctionTimeoutError: " + e.Exception
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

// NotFoundError is returned when a resource is not found.
type NotFoundError struct {
	Exception string
}

func (e NotFoundError) Error() string {
	return "NotFoundError: " + e.Exception
}

// InvalidError represents an invalid request or operation.
type InvalidError struct {
	Exception string
}

func (e InvalidError) Error() string {
	return "InvalidError: " + e.Exception
}

// QueueEmptyError is returned when an operation is attempted on an empty queue.
type QueueEmptyError struct {
	Exception string
}

func (e QueueEmptyError) Error() string {
	return "QueueEmptyError: " + e.Exception
}

// QueueFullError is returned when an operation is attempted on a full queue.
type QueueFullError struct {
	Exception string
}

func (e QueueFullError) Error() string {
	return "QueueFullError: " + e.Exception
}
