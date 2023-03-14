package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
)

type (
	getCitiesItem struct {
		ID   int16  `json:"id"`
		Name string `json:"name"`
	}

	getCitiesRequest struct {
		Search string `json:"search"`
		Limit  uint64 `json:"limit"`
		Cursor int16  `json:"cursor"`
	}

	getCitiesResponse struct {
		Items   []*getCitiesItem `json:"items"`
		HasNext bool             `json:"has_next"`
	}
)

func (s *server) getCitiesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req getCitiesRequest
		err = json.NewDecoder(r.Body).Decode(&req)
	)

	if err != nil {
		s.renderError(w, r,
			http.StatusBadRequest,
			fmt.Sprintf("failed to decode request: %s", err.Error()))
		return
	}

	serviceRequest := &city_service.GetListRequest{
		Search: req.Search,
		Limit:  req.Limit,
		Cursor: req.Cursor,
	}

	if err = serviceRequest.Validate(); err != nil {
		s.renderError(w, r,
			http.StatusBadRequest,
			err.Error())
		return
	}

	ctx := r.Context()
	res, err := s.cityService.GetList(ctx, serviceRequest)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to get cities list: %s", err.Error()))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.fillGetCitiesResponse(res))
}

func (s *server) fillGetCitiesResponse(res *city_service.GetListResponse) *getCitiesResponse {
	items := make([]*getCitiesItem, 0, len(res.Items))
	for _, v := range res.Items {
		items = append(items, &getCitiesItem{
			ID:   v.ID,
			Name: v.Name,
		})
	}

	return &getCitiesResponse{
		Items:   items,
		HasNext: res.HasNext,
	}
}
