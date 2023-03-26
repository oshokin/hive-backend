package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/logger"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *server) renderError(w http.ResponseWriter, r *http.Request, err *common.Error) {
	var (
		errType    = err.Type
		errMessage = err.Err.Error()
		errStatus  = errType.HTTPStatus()
		errClass   = errStatus / 100
		ctx        = r.Context()
	)

	if errClass == 4 {
		logger.Warn(ctx, errMessage)
	} else {
		logger.Error(ctx, errMessage)
	}

	render.Status(r, errType.HTTPStatus())
	render.JSON(w, r, &apiError{
		Code:    errType.String(),
		Message: err.Err.Error(),
	})
}
