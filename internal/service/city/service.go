package city

import (
	"context"

	repo "github.com/oshokin/hive-backend/internal/repository/city"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type (
	// Service defines the methods to manage cities.
	Service interface {
		// CheckIfExistByIDs checks if cities with the given IDs exist in the repository
		// Returns a map where each city ID is a key and struct{}{} is a value
		CheckIfExistByIDs(ctx context.Context, cityIDs []int16) (map[int16]struct{}, error)

		// GetAll retrieves all cities from the repository
		GetAll(ctx context.Context) ([]*City, error)

		// GetByID retrieves a city by its ID from the repository
		GetByID(ctx context.Context, id int16) (*City, error)

		// GetList retrieves a list of cities from the repository, filtered by the parameters in the GetListRequest
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
	}

	service struct {
		repository repo.Repository
	}
)

// NewService returns a new instance of the city service.
func NewService(r repo.Repository) Service {
	return &service{
		repository: r,
	}
}

func (s *service) CheckIfExistByIDs(ctx context.Context, cityIDs []int16) (map[int16]struct{}, error) {
	return s.repository.CheckIfExistByIDs(ctx, cityIDs)
}

func (s *service) GetAll(ctx context.Context) ([]*City, error) {
	res, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return s.getServiceModels(res), nil
}

func (s *service) GetByID(ctx context.Context, id int16) (*City, error) {
	res, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.getServiceModel(res), nil
}

func (s *service) GetList(ctx context.Context, r *GetListRequest) (*GetListResponse, error) {
	if err := r.validate(); err != nil {
		return nil, common.NewError(common.ErrStatusBadRequest, err)
	}

	limit := r.Limit
	if limit == 0 {
		limit = maxCitiesLimit
	}

	res, err := s.repository.GetList(ctx, &repo.GetListRequest{
		Search: r.Search,
		Limit:  limit,
		Cursor: r.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &GetListResponse{
		Items:   s.getServiceModels(res.Items),
		HasNext: res.HasNext,
	}, nil
}
