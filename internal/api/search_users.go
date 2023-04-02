package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type searchUsersResponse struct {
	Items   []*User `json:"items"`
	HasNext bool    `json:"has_next"`
}

func (s *server) searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	var (
		queryParams    = r.URL.Query()
		firstName      = queryParams.Get("first_name")
		lastName       = queryParams.Get("last_name")
		limit, _       = strconv.ParseUint(queryParams.Get("limit"), 10, 64)
		cursor, _      = strconv.ParseInt(queryParams.Get("cursor"), 10, 64)
		ctx            = r.Context()
		serviceRequest = &user_service.SearchByNamePrefixesRequest{
			FirstName: firstName,
			LastName:  lastName,
			Limit:     limit,
			Cursor:    cursor,
		}
	)

	res, err := s.userService.SearchByNamePrefixes(ctx, serviceRequest)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to search users: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.fillSearchUsersResponse(res))
}

func (s *server) fillSearchUsersResponse(res *user_service.SearchByNamePrefixesResponse) *searchUsersResponse {
	if res == nil {
		return nil
	}

	items := make([]*User, 0, len(res.Items))

	for _, v := range res.Items {
		if v == nil {
			continue
		}

		items = append(items, s.getUserModel(v))
	}

	return &searchUsersResponse{
		Items:   items,
		HasNext: res.HasNext,
	}
}
