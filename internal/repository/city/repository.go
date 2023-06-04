package city

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
	"github.com/oshokin/hive-backend/internal/db"
)

type (
	// Repository defines the methods for interacting with the City data store.
	Repository interface {
		// CheckIfExistByIDs checks if the City with the given IDs exist in the data store.
		// It returns a map with City IDs as keys and empty structs as values,
		// for the Cities that exist in the data store.
		CheckIfExistByIDs(ctx context.Context, cityIDs []int16) (map[int16]struct{}, error)

		// GetAll returns all Cities in the data store.
		GetAll(ctx context.Context) ([]*City, error)

		// GetByID returns the City with the given ID.
		GetByID(ctx context.Context, id int16) (*City, error)

		// GetList returns a paginated list of Cities matching the given search criteria.
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
	}

	repository struct {
		cluster *db.Cluster
	}
)

const (
	tableName  = "cities"
	columnID   = "id"
	columnName = "name"
)

// NewRepository creates a new Repository instance with the given database cluster.
func NewRepository(cluster *db.Cluster) Repository {
	return &repository{
		cluster: cluster,
	}
}

func (r *repository) CheckIfExistByIDs(ctx context.Context, cityIDs []int16) (map[int16]struct{}, error) {
	selectQB := sq.StatementBuilder.
		Select(columnID).
		From(tableName).
		Where(sq.Eq{columnID: cityIDs}).
		PlaceholderFormat(sq.Dollar)

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.cluster.ReadRR().Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}
	defer rows.Close()

	existingCityIDs := make(map[int16]struct{}, len(cityIDs))

	for rows.Next() {
		var cityID int16

		err = rows.Scan(&cityID)
		if err != nil {
			return nil, fmt.Errorf("failed to read query results: %w", err)
		}

		existingCityIDs[cityID] = struct{}{}
	}

	return existingCityIDs, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*City, error) {
	sortByName := fmt.Sprintf("%s ASC", columnName)

	selectQB := sq.StatementBuilder.
		Select(columnID, columnName).
		From(tableName).
		OrderBy(sortByName).
		PlaceholderFormat(sq.Dollar)

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := r.cluster.ReadRR().Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}
	defer rows.Close()

	var cities []*City

	for rows.Next() {
		var city City

		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to read query results: %w", err)
		}

		cities = append(cities, &city)
	}

	return cities, nil
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

	var city City

	err = r.cluster.ReadRR().QueryRow(ctx, query, args...).Scan(&city.ID, &city.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read query results: %w", err)
	}

	return &city, nil
}

func (r *repository) GetList(ctx context.Context,
	req *GetListRequest) (*GetListResponse, error) {
	sortByName := fmt.Sprintf("%s ASC", columnName)

	selectQB := sq.StatementBuilder.
		Select(columnID, columnName).
		From(tableName).
		OrderBy(sortByName).
		Limit(req.Limit + 1).
		PlaceholderFormat(sq.Dollar)
	if req.Cursor != 0 {
		selectQB = selectQB.Where(sq.Gt{columnID: req.Cursor})
	}

	if req.Search != "" {
		selectQB = selectQB.Where(sq.Like{columnName: fmt.Sprintf("%%%s%%", req.Search)})
	}

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.cluster.ReadRR().Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run select query: %w", err)
	}
	defer rows.Close()

	var (
		cities  []*City
		hasNext bool
	)

	for rows.Next() {
		if uint64(len(cities)) >= req.Limit {
			hasNext = true
			break
		}

		var city City

		err = rows.Scan(&city.ID, &city.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to read select query results: %w", err)
		}

		cities = append(cities, &city)
	}

	return &GetListResponse{
		Items:   cities,
		HasNext: hasNext,
	}, nil
}
