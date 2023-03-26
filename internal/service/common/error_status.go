package common

import "net/http"

type ErrorStatus uint8

const (
	ErrStatusUnknown ErrorStatus = iota
	ErrStatusBadRequest
	ErrStatusUnauthorized
	ErrStatusForbidden
	ErrStatusNotFound
	ErrStatusConflict
	ErrStatusInternalError
)

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
	case ErrStatusInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (es ErrorStatus) String() string {
	switch es {
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
		return "UNKNOWN_ERROR"
	}
}
