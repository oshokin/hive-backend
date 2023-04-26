package randomizing_job

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	// Repository defines the interface for interacting with the randomizing_jobs table.
	Repository interface {
		Create(ctx context.Context, expectedCount int64) (int64, error)
		GetByID(ctx context.Context, id int64) (*RandomizingJob, error)
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
		Update(ctx context.Context, job *RandomizingJob, fields *UpdateFields) error
	}

	repository struct {
		db *pgxpool.Pool
	}
)

const (
	tableName           = "randomizing_jobs"
	columnID            = "id"
	columnExpectedCount = "expected_count"
	columnCurrentCount  = "current_count"
	columnStatus        = "status"
	columnStartedAt     = "started_at"
	columnFinishedAt    = "finished_at"
	columnErrorMessage  = "error_message"

	// maxLimit defines the maximum limit for list requests.
	maxLimit = 50
)

// NewRepository creates a new repository with the given database connection pool.
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, expectedCount int64) (int64, error) {
	query, args, err := sq.Insert(tableName).
		Columns(columnExpectedCount).
		Values(expectedCount).
		Suffix(fmt.Sprintf("RETURNING %s", columnID)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	var id int64

	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to read query results: %w", err)
	}

	return id, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*RandomizingJob, error) {
	query, args, err := sq.Select(columnID,
		columnExpectedCount,
		columnCurrentCount,
		columnStatus,
		columnStartedAt,
		columnFinishedAt,
		columnErrorMessage).
		From(tableName).
		Where(sq.Eq{columnID: id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	job := new(RandomizingJob)

	err = r.db.QueryRow(ctx, query, args...).
		Scan(&job.ID,
			&job.ExpectedCount,
			&job.CurrentCount,
			&job.Status,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ErrorMessage)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read query results: %w", err)
	}

	return job, nil
}

func (r *repository) GetList(ctx context.Context,
	req *GetListRequest) (*GetListResponse, error) {
	sortByID := fmt.Sprintf("%s ASC", columnID)

	selectQB := sq.StatementBuilder.
		Select(columnID,
			columnExpectedCount,
			columnCurrentCount,
			columnStatus,
			columnStartedAt,
			columnFinishedAt,
			columnErrorMessage).
		From(tableName).
		OrderBy(sortByID).
		Limit(req.Limit + 1).
		PlaceholderFormat(sq.Dollar)
	if req.Cursor != 0 {
		selectQB = selectQB.Where(sq.Gt{columnID: req.Cursor})
	}

	if len(req.Status) > 0 {
		selectQB = selectQB.Where(sq.Eq{columnStatus: req.Status})
	}

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run select query: %w", err)
	}
	defer rows.Close()

	var (
		randomizingJobs []*RandomizingJob
		hasNext         bool
	)

	for rows.Next() {
		if uint64(len(randomizingJobs)) >= req.Limit {
			hasNext = true
			break
		}

		var job RandomizingJob

		err = rows.Scan(&job.ID,
			&job.ExpectedCount,
			&job.CurrentCount,
			&job.Status,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ErrorMessage)

		if err != nil {
			return nil, fmt.Errorf("failed to read select query results: %w", err)
		}

		randomizingJobs = append(randomizingJobs, &job)
	}

	return &GetListResponse{
		Items:   randomizingJobs,
		HasNext: hasNext,
	}, nil
}

func (r *repository) Update(ctx context.Context, job *RandomizingJob, fields *UpdateFields) error {
	if fields == nil {
		return nil
	}

	updateBuilder := sq.Update(tableName).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{columnID: job.ID})

	if fields.ExpectedCount {
		updateBuilder = updateBuilder.Set(columnExpectedCount, job.ExpectedCount)
	}

	if fields.CurrentCount {
		updateBuilder = updateBuilder.Set(columnCurrentCount, job.CurrentCount)
	}

	if fields.Status {
		updateBuilder = updateBuilder.Set(columnStatus, job.Status)
	}

	if fields.StartedAt {
		updateBuilder = updateBuilder.Set(columnStartedAt, job.StartedAt)
	}

	if fields.FinishedAt {
		updateBuilder = updateBuilder.Set(columnFinishedAt, job.FinishedAt)
	}

	if fields.ErrorMessage {
		updateBuilder = updateBuilder.Set(columnErrorMessage, job.ErrorMessage)
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	commandTag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}
