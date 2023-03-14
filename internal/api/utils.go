package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/logger"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	badRequestErrorCode = "BAD_REQUEST"
	internalErrorCode   = "INTERNAL_ERROR"
)

func (s *server) getErrorCodeFromHTTPStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return badRequestErrorCode
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusNotAcceptable:
		return "NOT_ACCEPTABLE"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	case http.StatusInternalServerError:
		return internalErrorCode
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		if errorClass := status / 100; errorClass == 4 {
			return badRequestErrorCode
		}

		return internalErrorCode
	}
}

func (s *server) renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	var (
		errorClass = status / 100
		ctx        = r.Context()
	)

	if errorClass == 4 {
		logger.Warn(ctx, message)
	} else {
		logger.Error(ctx, message)
	}

	render.Status(r, status)
	render.JSON(w, r, &apiError{
		Code:    s.getErrorCodeFromHTTPStatus(status),
		Message: message,
	})
}
