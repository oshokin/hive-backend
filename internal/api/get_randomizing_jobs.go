package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
	"github.com/oshokin/hive-backend/internal/service/randomizing_job"
)

type (
	getRandomizingJobsItem struct {
		ID            int64      `json:"id"`
		ExpectedCount int64      `json:"expected_count"`
		CurrentCount  int64      `json:"current_count"`
		Status        string     `json:"status"`
		StartedAt     *time.Time `json:"started_at"`
		FinishedAt    *time.Time `json:"finished_at"`
		ErrorMessage  string     `json:"error_message"`
	}

	getRandomizingJobsResponse struct {
		Items   []*getRandomizingJobsItem `json:"items"`
		HasNext bool                      `json:"has_next"`
	}
)

func (s *server) getRandomizingJobsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		queryParams    = r.URL.Query()
		limit, _       = strconv.ParseUint(queryParams.Get("limit"), 10, 64)
		cursor, _      = strconv.ParseInt(queryParams.Get("cursor"), 10, 64)
		statuses       = queryParams["status"]
		ctx            = r.Context()
		serviceRequest = &randomizing_job.GetListRequest{
			Status: getRandomizingJobStatuses(statuses),
			Limit:  limit,
			Cursor: cursor,
		}
	)

	res, err := s.randomizingJobService.GetList(ctx, serviceRequest)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to get randomizing jobs list: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.fillGetRandomizingJobsResponse(res))
}

func (s *server) fillGetRandomizingJobsResponse(res *randomizing_job.GetListResponse) *getRandomizingJobsResponse {
	if res == nil {
		return nil
	}

	items := make([]*getRandomizingJobsItem, 0, len(res.Items))

	for _, v := range res.Items {
		if v == nil {
			continue
		}

		items = append(items, &getRandomizingJobsItem{
			ID:            v.ID,
			ExpectedCount: v.ExpectedCount,
			CurrentCount:  v.CurrentCount,
			Status:        string(v.Status),
			StartedAt:     v.StartedAt,
			FinishedAt:    v.FinishedAt,
			ErrorMessage:  v.ErrorMessage,
		})
	}

	return &getRandomizingJobsResponse{
		Items:   items,
		HasNext: res.HasNext,
	}
}

func getRandomizingJobStatuses(v []string) []randomizing_job.JobStatus {
	result := make([]randomizing_job.JobStatus, 0, len(v))

	for _, status := range v {
		switch status {
		case string(randomizing_job.JobStatusQueued):
			result = append(result, randomizing_job.JobStatusQueued)
		case string(randomizing_job.JobStatusProcessing):
			result = append(result, randomizing_job.JobStatusProcessing)
		case string(randomizing_job.JobStatusCancelled):
			result = append(result, randomizing_job.JobStatusCancelled)
		case string(randomizing_job.JobStatusCompleted):
			result = append(result, randomizing_job.JobStatusCompleted)
		case string(randomizing_job.JobStatusFailed):
			result = append(result, randomizing_job.JobStatusFailed)
		}
	}

	return result
}
