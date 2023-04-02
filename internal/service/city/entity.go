package city

import (
	"fmt"

	repo "github.com/oshokin/hive-backend/internal/repository/city"
)

type (
	// City represents a city entity with its ID and name.
	City struct {
		ID   int16
		Name string
	}

	// GetListRequest represents a request to get a list of cities.
	GetListRequest struct {
		Search string // Search parameter to filter cities.
		Limit  uint64 // Limit of cities to return.
		Cursor int16  // Cursor is used for pagination.
	}

	// GetListResponse represents a response containing a list of cities.
	GetListResponse struct {
		Items   []*City // List of cities.
		HasNext bool    // Indicates whether there are more items to be retrieved.
	}
)

const maxCitiesLimit = 50

func (s *service) getServiceModel(source *repo.City) *City {
	if source == nil {
		return nil
	}

	return &City{
		ID:   source.ID,
		Name: source.Name,
	}
}

func (s *service) getServiceModels(source []*repo.City) []*City {
	result := make([]*City, 0, len(source))

	for _, v := range source {
		sm := s.getServiceModel(v)
		if sm == nil {
			continue
		}

		result = append(result, sm)
	}

	return result
}

func (r *GetListRequest) validate() error {
	if r == nil {
		return nil
	}

	if r.Limit > maxCitiesLimit {
		return fmt.Errorf("maximum cities count in one request is %d items", maxCitiesLimit)
	}

	return nil
}
