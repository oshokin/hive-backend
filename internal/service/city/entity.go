package city

import (
	"fmt"

	repo "github.com/oshokin/hive-backend/internal/repository/city"
)

type (
	City struct {
		ID   int16
		Name string
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
)

func (s *service) GetServiceModel(source *repo.City) *City {
	if source == nil {
		return nil
	}

	return &City{
		ID:   source.ID,
		Name: source.Name,
	}
}

func (s *service) GetServiceModels(source []*repo.City) []*City {
	result := make([]*City, 0, len(source))
	for _, v := range source {
		sm := s.GetServiceModel(v)
		if sm == nil {
			continue
		}

		result = append(result, sm)
	}

	return result
}

func (r *GetListRequest) Validate() error {
	if r.Limit > maxCitiesLimit {
		return fmt.Errorf("maximum cities count in one request is %d items", maxCitiesLimit)
	}

	return nil
}
