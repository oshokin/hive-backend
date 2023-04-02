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
	createRandomizingJobRequest struct {
		ExpectedCount int64 `json:"expected_count"`
	}

	createRandomizingJobResponse struct {
		JobID int64 `json:"job_id"`
	}
)

func (s *server) createRandomizingJobHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req createRandomizingJobRequest
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

	jobID, err := s.randomizingJobService.Create(ctx, req.ExpectedCount)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to create randomizing job: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &createRandomizingJobResponse{
		JobID: jobID,
	})
}
