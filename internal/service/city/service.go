package city

import (
	"context"

	city_repo "github.com/oshokin/hive-backend/internal/repository/city"
	"github.com/oshokin/hive-backend/internal/service/common"
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

func (s *service) GetList(ctx context.Context, r *GetListRequest) (*GetListResponse, error) {
	if err := r.validate(); err != nil {
		return nil, common.NewError(common.ErrStatusBadRequest, err)
	}

	limit := r.Limit
	if limit == 0 {
		limit = maxCitiesLimit
	}

	res, err := s.repository.GetList(ctx, &city_repo.GetListRequest{
		Search: r.Search,
		Limit:  limit,
		Cursor: r.Cursor,
	})
	if err != nil {
		return nil, err
	}

	return &GetListResponse{
		Items:   s.GetServiceModels(res.Items),
		HasNext: res.HasNext,
	}, nil
}
