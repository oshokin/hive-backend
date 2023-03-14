package city

import (
	"context"

	city_repo "github.com/oshokin/hive-backend/internal/repository/city"
)

type (
	Service interface {
		GetByID(ctx context.Context, id int16) (*City, error)
		GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error)
	}

	service struct {
		repository city_repo.Repository
	}
)

const maxCitiesLimit = 50

func NewService(r city_repo.Repository) Service {
	return &service{repository: r}
}

func (s *service) GetByID(ctx context.Context, id int16) (*City, error) {
	res, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetServiceModel(res), nil
}

func (s *service) GetList(ctx context.Context, req *GetListRequest) (*GetListResponse, error) {
	res, err := s.repository.GetList(ctx, &city_repo.GetListRequest{
		Search: req.Search,
		Limit:  req.Limit,
		Cursor: req.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &GetListResponse{
		Items:   s.GetServiceModels(res.Items),
		HasNext: res.HasNext,
	}, nil
}
