package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type createUserRequest struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	CityID    int16    `json:"city_id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Birthdate dateOnly `json:"birthdate"`
	Gender    string   `json:"gender"`
	Interests string   `json:"interests"`
}

type createUserResponse struct {
	UserID int64 `json:"user_id"`
}

func (s *server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req createUserRequest
		err = json.NewDecoder(r.Body).Decode(&req)
	)

	if err != nil {
		s.renderError(w, r,
			common.NewError(common.ErrStatusBadRequest,
				fmt.Errorf("failed to decode request: %w", err)))

		return
	}

	var (
		ctx = r.Context()
		u   = &user_service.User{
			Email:     req.Email,
			Password:  req.Password,
			CityID:    req.CityID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Birthdate: time.Time(req.Birthdate),
			Gender:    user_service.GenderType(req.Gender),
			Interests: req.Interests,
		}
	)

	userID, err := s.userService.Create(ctx, u)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to register user: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &createUserResponse{
		UserID: userID,
	})
}
