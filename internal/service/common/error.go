package common

type Error struct {
	Type ErrorStatus
	Err  error
}

func NewError(status ErrorStatus, err error) *Error {
	return &Error{
		Type: status,
		Err:  err,
	}
}

func (e *Error) Error() string {
	return e.Err.Error()
}
