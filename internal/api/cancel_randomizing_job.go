package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type (
	cancelRandomizingJobRequest struct {
		ID int64 `json:"id"`
	}

	cancelRandomizingJobResponse struct {
		Success bool `json:"success"`
	}
)

func (s *server) cancelRandomizingJobHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req cancelRandomizingJobRequest
		err = json.NewDecoder(r.Body).Decode(&req)
	)

	if err != nil {
		s.renderError(w, r,
			common.NewError(common.ErrStatusBadRequest,
				fmt.Errorf("failed to decode request: %w", err)))

		return
	}

	var (
		ctx = r.Context()
	)

	err = s.randomizingJobService.Cancel(ctx, req.ID)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to cancel randomizing job: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &cancelRandomizingJobResponse{
		Success: true,
	})
}
