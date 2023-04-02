package common

// Error represents an error with an associated ErrorStatus.
type Error struct {
	Type ErrorStatus // Type is the type of error that occurred.
	Err  error       // Err is the underlying error that caused the error.
}

// NewError creates a new Error object with the given ErrorStatus and error.
func NewError(status ErrorStatus, err error) *Error {
	return &Error{
		Type: status,
		Err:  err,
	}
}

// Error returns a string representation of the error message.
func (e *Error) Error() string {
	return e.Err.Error()
}
