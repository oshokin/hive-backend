// Package randomizing_job provides a service to randomize users
package randomizing_job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oshokin/hive-backend/internal/common"
	"github.com/oshokin/hive-backend/internal/logger"
	randomizing_job_repo "github.com/oshokin/hive-backend/internal/repository/randomizing_job"
	common_service "github.com/oshokin/hive-backend/internal/service/common"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type (
	// Service provides methods for managing RandomizingJob instances.
	Service interface {
		// Start starts the randomizing jobs service.
		Start(ctx context.Context)
		// Stop stops the randomizing jobs service.
		Stop(ctx context.Context)
		// Create creates a new RandomizingJob.
		Create(ctx context.Context, expectedCount int64) (int64, error)
		// GetByID gets a RandomizingJob by ID.
		GetByID(ctx context.Context, id int64) (*RandomizingJob, error)
		// GetList gets a list of RandomizingJobs based on the search criteria.
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
		// Cancel cancels a running RandomizingJob.
		Cancel(ctx context.Context, id int64) error
	}

	service struct {
		randomizingJobRepository randomizing_job_repo.Repository
		userService              user_service.Service
		runningJobs              map[int64]context.CancelFunc
		mu                       sync.Mutex
	}
)

const (
	defaultTimeout   = 5 * time.Second
	batchPortionSize = 10000
)

var (
	jobSearchRequest = &GetListRequest{
		Status: []JobStatus{JobStatusQueued, JobStatusProcessing},
	}

	statusUpdateFields = &randomizing_job_repo.UpdateFields{
		Status:       true,
		FinishedAt:   true,
		ErrorMessage: true,
	}

	currentCountUpdateFields = &randomizing_job_repo.UpdateFields{
		CurrentCount: true,
	}
)

// NewService returns a new instance of the randomizing jobs service.
func NewService(r randomizing_job_repo.Repository, u user_service.Service) Service {
	return &service{
		randomizingJobRepository: r,
		userService:              u,
		runningJobs:              make(map[int64]context.CancelFunc),
	}
}

func (s *service) Start(ctx context.Context) {
	logger.Infof(ctx, "starting randomizing job service")

	go s.start(ctx)

	logger.Infof(ctx, "starting randomizing job is running")
}

func (s *service) Stop(ctx context.Context) {
	logger.Info(ctx, "shutting down randomizing job service")
	s.mu.Lock()

	for _, cancel := range s.runningJobs {
		if cancel == nil {
			continue
		}

		cancel()
	}

	s.mu.Unlock()
	logger.Info(ctx, "randomizing job service stopped")
}

func (s *service) Create(ctx context.Context, expectedCount int64) (int64, error) {
	if expectedCount <= 0 {
		return 0, common_service.NewError(common_service.ErrStatusBadRequest,
			errors.New("expected users count must be greater than 0"))
	}

	return s.randomizingJobRepository.Create(ctx, expectedCount)
}

func (s *service) GetByID(ctx context.Context, id int64) (*RandomizingJob, error) {
	if id <= 0 {
		return nil, common_service.NewError(common_service.ErrStatusBadRequest,
			errors.New("job ID must be greater than 0"))
	}

	res, err := s.randomizingJobRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.getServiceModel(res), nil
}

func (s *service) GetList(ctx context.Context, r *GetListRequest) (*GetListResponse, error) {
	if err := r.validate(); err != nil {
		return nil, common_service.NewError(common_service.ErrStatusBadRequest, err)
	}

	if r.Limit == 0 {
		r.Limit = maxJobsLimit
	}

	res, err := s.randomizingJobRepository.GetList(ctx, s.getListRequestRepoModel(r))
	if err != nil {
		return nil, err
	}

	return &GetListResponse{
		Items:   s.getServiceModels(res.Items),
		HasNext: res.HasNext,
	}, nil
}

func (s *service) Cancel(ctx context.Context, id int64) error {
	job, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.cancelJob(ctx, job)
}

func (s *service) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "randomizing job context was cancelled")
			return
		default:
			lr, err := s.GetList(ctx, jobSearchRequest)
			if err != nil {
				logger.ErrorKV(ctx, "failed to get jobs", common.ErrorTag, err)
				time.Sleep(defaultTimeout)

				continue
			}

			for _, job := range lr.Items {
				jobCtx, cancel := context.WithCancel(ctx)
				s.addCancelForJob(job.ID, cancel)

				err = s.runJob(jobCtx, job)
				if err != nil {
					logger.ErrorKV(ctx, "failed to run job",
						common.RandomizingJobIDTag, job.ID,
						common.ErrorTag, err)
				}

				s.deleteJobCancel(job.ID)
			}
		}
	}
}

func (s *service) cancelJob(ctx context.Context, job *RandomizingJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := job.ID

	cancel, ok := s.runningJobs[id]
	if !ok {
		return fmt.Errorf("cancel function for job %d not found", id)
	}

	cancel()
	delete(s.runningJobs, id)

	job.Status = JobStatusCancelled
	job.ErrorMessage = "job was cancelled by user"

	return s.finishJob(ctx, job)
}

func (s *service) addCancelForJob(id int64, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runningJobs[id] = cancel
}

func (s *service) deleteJobCancel(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.runningJobs, id)
}

func (s *service) finishJob(ctx context.Context, job *RandomizingJob) error {
	finishedAt := time.Now()
	job.FinishedAt = &finishedAt

	logger.InfoKV(ctx, "finishing randomizing job",
		common.RandomizingJobIDTag, job.ID,
		common.RandomizingJobStatusTag, job.Status,
		common.RandomizingJobErrorMessageTag, job.ErrorMessage)

	return s.randomizingJobRepository.Update(ctx, s.getRepoModel(job), statusUpdateFields)
}

func (s *service) runJob(ctx context.Context, job *RandomizingJob) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			usersLeftToAdd := common.Min(batchPortionSize, job.ExpectedCount-job.CurrentCount)
			logger.InfoKV(ctx, "starting new adding new portion",
				common.RandomizingJobIDTag, job.ID,
				common.UsersLeftToAddTag, usersLeftToAdd)

			if usersLeftToAdd <= 0 {
				job.Status = JobStatusCompleted
				job.ErrorMessage = ""

				return s.finishJob(ctx, job)
			}

			users, err := s.userService.GenerateRandomData(ctx, usersLeftToAdd)
			if err != nil {
				job.Status = JobStatusFailed
				job.ErrorMessage = fmt.Sprintf("failed to generate random data: %s", err.Error())

				return s.finishJob(ctx, job)
			}

			usersCount, validationErrors, err := s.userService.CreateBatch(ctx, users)
			if err != nil {
				job.Status = JobStatusFailed
				job.ErrorMessage = fmt.Sprintf("failed to create batch: %s", err.Error())

				return s.finishJob(ctx, job)
			}

			for u, err := range validationErrors {
				logger.WarnKV(ctx, "there are errors in random data",
					common.RandomizingJobIDTag, job.ID,
					common.ErrorTag, err,
					common.UserTag, u)
			}

			job.CurrentCount += usersCount
			job.CurrentCount = common.Min(job.CurrentCount, job.ExpectedCount)

			logger.InfoKV(ctx, "added new users portion",
				common.RandomizingJobIDTag, job.ID,
				common.UsersLeftToAddTag, usersLeftToAdd,
				common.AddedUsersCountTag, usersCount,
				common.CurrentCountTag, job.CurrentCount)

			err = s.randomizingJobRepository.Update(ctx,
				s.getRepoModel(job),
				currentCountUpdateFields)
			if err != nil {
				job.Status = JobStatusFailed
				job.ErrorMessage = fmt.Sprintf("failed to update count: %s", err.Error())

				return s.finishJob(ctx, job)
			}
		}
	}
}
