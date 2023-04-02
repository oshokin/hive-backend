package common

import "net/http"

// ErrorStatus represents a status code for an error.
type ErrorStatus uint8

// Constants representing different error status codes.
const (
	ErrStatusUnknown       ErrorStatus = iota // Unknown error status
	ErrStatusBadRequest                       // Bad request error status
	ErrStatusUnauthorized                     // Unauthorized error status
	ErrStatusForbidden                        // Forbidden error status
	ErrStatusNotFound                         // Not found error status
	ErrStatusConflict                         // Conflict error status
	ErrStatusInternalError                    // Internal error status

	unknownErrorCode = "UNKNOWN_ERROR"
)

// HTTPStatus returns the corresponding HTTP status code for the error status.
func (es ErrorStatus) HTTPStatus() int {
	switch es {
	case ErrStatusBadRequest:
		return http.StatusBadRequest
	case ErrStatusUnauthorized:
		return http.StatusUnauthorized
	case ErrStatusForbidden:
		return http.StatusForbidden
	case ErrStatusNotFound:
		return http.StatusNotFound
	case ErrStatusConflict:
		return http.StatusConflict
	case ErrStatusUnknown, ErrStatusInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// String returns the string representation of the error status.
func (es ErrorStatus) String() string {
	switch es {
	case ErrStatusUnknown:
		return unknownErrorCode
	case ErrStatusBadRequest:
		return "BAD_REQUEST"
	case ErrStatusUnauthorized:
		return "UNAUTHORIZED"
	case ErrStatusForbidden:
		return "FORBIDDEN"
	case ErrStatusNotFound:
		return "NOT_FOUND"
	case ErrStatusConflict:
		return "CONFLICT"
	case ErrStatusInternalError:
		return "INTERNAL_ERROR"
	default:
		return unknownErrorCode
	}
}
