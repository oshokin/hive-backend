package city

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type (
	Repository interface {
		GetByID(ctx context.Context, id int16) (*City, error)
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
	}

	GetListRequest struct {
		Search string
		Limit  uint64
		Cursor int16
	}

	GetListResponse struct {
		Items   []*City
		HasNext bool
	}

	repository struct {
		db *pgxpool.Pool
	}
)

const (
	tableName  = "cities"
	columnID   = "id"
	columnName = "name"
)

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id int16) (*City, error) {
	query, args, err := sq.Select(columnID, columnName).
		From(tableName).
		Where(sq.Eq{columnID: id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	city := &City{}
	err = r.db.QueryRow(ctx, query, args...).Scan(&city.ID, &city.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read query results: %w", err)
	}

	return city, nil
}

func (r *repository) GetList(ctx context.Context,
	req *GetListRequest) (*GetListResponse, error) {

	sortByName := fmt.Sprintf("%s ASC", columnName)

	selectQB := sq.StatementBuilder.
		Select(columnID, columnName).
		From(tableName).
		OrderBy(sortByName).
		Limit(req.Limit).
		PlaceholderFormat(sq.Dollar)
	if req.Cursor != 0 {
		selectQB = selectQB.Where(sq.Gt{columnID: req.Cursor})
	}

	if req.Search != "" {
		selectQB = selectQB.Where(sq.Like{columnName: fmt.Sprintf("%%%s%%", req.Search)})
	}

	statsQB := sq.StatementBuilder.
		Select(fmt.Sprintf("COUNT(%s)", columnID)).
		From(tableName).
		Limit(req.Limit + 1).
		PlaceholderFormat(sq.Dollar)
	if req.Cursor != 0 {
		statsQB = statsQB.Where(sq.Gt{columnID: req.Cursor})
	}

	if req.Search != "" {
		statsQB = statsQB.Where(sq.Like{columnName: fmt.Sprintf("%%%s%%", req.Search)})
	}

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	statsQuery, statsArgs, err := statsQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build stats query: %w", err)
	}

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run select query: %w", err)
	}
	defer rows.Close()

	var cities []*City
	for rows.Next() {
		var (
			city City
			err  = rows.Scan(&city.ID, &city.Name)
		)

		if err != nil {
			return nil, fmt.Errorf("failed to read select query results: %w", err)
		}

		cities = append(cities, &city)
	}

	var (
		countRow   = r.db.QueryRow(ctx, statsQuery, statsArgs...)
		statsCount uint64
	)

	err = countRow.Scan(&statsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats query results: %w", err)
	}

	return &GetListResponse{
		Items:   cities,
		HasNext: statsCount > uint64(len(cities)),
	}, nil
}
