package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type (
	getCitiesItem struct {
		ID   int16  `json:"id"`
		Name string `json:"name"`
	}

	getCitiesResponse struct {
		Items   []*getCitiesItem `json:"items"`
		HasNext bool             `json:"has_next"`
	}
)

func (s *server) getCitiesHandler(w http.ResponseWriter, r *http.Request) {
	var (
		queryParams    = r.URL.Query()
		search         = queryParams.Get("search")
		limit, _       = strconv.ParseUint(queryParams.Get("limit"), 10, 64)
		cursor, _      = strconv.ParseInt(queryParams.Get("cursor"), 10, 16)
		ctx            = r.Context()
		serviceRequest = &city_service.GetListRequest{
			Search: search,
			Limit:  limit,
			Cursor: int16(cursor),
		}
	)

	res, err := s.cityService.GetList(ctx, serviceRequest)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to get cities list: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.fillGetCitiesResponse(res))
}

func (s *server) fillGetCitiesResponse(res *city_service.GetListResponse) *getCitiesResponse {
	if res == nil {
		return nil
	}

	items := make([]*getCitiesItem, 0, len(res.Items))

	for _, v := range res.Items {
		if v == nil {
			continue
		}

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
