package randomizing_job

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	repo "github.com/oshokin/hive-backend/internal/repository/randomizing_job"
)

type (
	// RandomizingJob represents a single job that needs to be processed.
	RandomizingJob struct {
		ID            int64      // unique identifier for the job
		ExpectedCount int64      // the total number of items that should be processed by this job
		CurrentCount  int64      // the number of items that have already been processed by this job
		Status        JobStatus  // the current status of the job (queued, processing, cancelled, completed, failed)
		StartedAt     *time.Time // the time when the job was started (nil if not started yet)
		FinishedAt    *time.Time // the time when the job was finished (nil if not finished yet)
		ErrorMessage  string     // the error message associated with the job (empty string if no error)
	}

	// GetListRequest represents a request to get a list of jobs.
	GetListRequest struct {
		Status []JobStatus // list of job statuses to filter by (empty means all statuses)
		Limit  uint64      // maximum number of jobs to return in a single response
		Cursor int64       // cursor for pagination (0 means the beginning of the list)
	}

	// GetListResponse represents a response containing a list of jobs.
	GetListResponse struct {
		Items   []*RandomizingJob // the list of jobs
		HasNext bool              // whether there are more jobs to retrieve
	}

	// JobStatus represents the status of a job.
	JobStatus string
)

// Possible values for JobStatus.
const (
	JobStatusQueued     JobStatus = "QUEUED"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCancelled  JobStatus = "CANCELLED"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// maxJobsLimit defines the maximum number of jobs to be returned in a single request.
const maxJobsLimit = 50

func (s *service) getServiceModel(source *repo.RandomizingJob) *RandomizingJob {
	if source == nil {
		return nil
	}

	return &RandomizingJob{
		ID:            source.ID,
		ExpectedCount: source.ExpectedCount,
		CurrentCount:  source.CurrentCount,
		Status:        JobStatus(source.Status),
		StartedAt:     source.StartedAt,
		FinishedAt:    source.FinishedAt,
		ErrorMessage:  source.ErrorMessage,
	}
}

func (s *service) getServiceModels(source []*repo.RandomizingJob) []*RandomizingJob {
	result := make([]*RandomizingJob, 0, len(source))

	for _, v := range source {
		sm := s.getServiceModel(v)
		if sm == nil {
			continue
		}

		result = append(result, sm)
	}

	return result
}

func (s *service) getRepoModel(source *RandomizingJob) *repo.RandomizingJob {
	if source == nil {
		return nil
	}

	return &repo.RandomizingJob{
		ID:            source.ID,
		ExpectedCount: source.ExpectedCount,
		CurrentCount:  source.CurrentCount,
		Status:        string(source.Status),
		StartedAt:     source.StartedAt,
		FinishedAt:    source.FinishedAt,
		ErrorMessage:  source.ErrorMessage,
	}
}

func (s *service) getListRequestRepoModel(r *GetListRequest) *repo.GetListRequest {
	if r == nil {
		return nil
	}

	statuses := make([]string, 0, len(r.Status))
	for _, status := range r.Status {
		statuses = append(statuses, string(status))
	}

	return &repo.GetListRequest{
		Status: statuses,
		Limit:  r.Limit,
		Cursor: r.Cursor,
	}
}

// String returns a string representation of the RandomizingJob object.
func (rj *RandomizingJob) String() string {
	var sb strings.Builder

	sb.WriteString("randomizing_job{id=")
	sb.WriteString(strconv.FormatInt(rj.ID, 10))
	sb.WriteString(", expected_count=")
	sb.WriteString(strconv.FormatInt(rj.ExpectedCount, 10))
	sb.WriteString(", current_count=")
	sb.WriteString(strconv.FormatInt(rj.CurrentCount, 10))
	sb.WriteString(", status=")
	sb.WriteString(string(rj.Status))
	sb.WriteString(", started_at=")

	if rj.StartedAt == nil {
		sb.WriteString("nil")
	} else {
		sb.WriteString(rj.StartedAt.Format(time.RFC3339Nano))
	}

	sb.WriteString(", finished_at=")

	if rj.StartedAt == nil {
		sb.WriteString("nil")
	} else {
		sb.WriteString(rj.FinishedAt.Format(time.RFC3339Nano))
	}

	sb.WriteString(", error_message")
	sb.WriteString(rj.ErrorMessage)
	sb.WriteString("}")

	return sb.String()
}

func (r *GetListRequest) validate() error {
	if r == nil {
		return nil
	}

	if r.Limit > maxJobsLimit {
		return fmt.Errorf("maximum jobs count in one request is %d items", maxJobsLimit)
	}

	return nil
}
