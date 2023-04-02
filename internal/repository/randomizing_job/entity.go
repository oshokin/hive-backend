package randomizing_job

import "time"

type (
	// RandomizingJob ...
	RandomizingJob struct {
		ID            int64
		ExpectedCount int64
		CurrentCount  int64
		Status        string
		StartedAt     *time.Time
		FinishedAt    *time.Time
		ErrorMessage  string
	}

	// UpdateFields ...
	UpdateFields struct {
		ExpectedCount bool
		CurrentCount  bool
		Status        bool
		StartedAt     bool
		FinishedAt    bool
		ErrorMessage  bool
	}

	// GetListRequest ...
	GetListRequest struct {
		Status []string
		Limit  uint64
		Cursor int64
	}

	// GetListResponse ...
	GetListResponse struct {
		Items   []*RandomizingJob
		HasNext bool
	}
)
