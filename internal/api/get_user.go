package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CityID    int16  `json:"city_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
}

func (s *server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		s.renderError(w, r,
			common.NewError(common.ErrStatusBadRequest,
				fmt.Errorf("failed to parse user ID: %w", err)))
		return
	}

	user, err := s.userService.GetByID(r.Context(), userID)
	if err != nil {
		switch v := err.(type) {
		case *common.Error:
			s.renderError(w, r, v)
		default:
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to get user info: %w", err)))
		}

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.getUserModel(user))
}

func (s *server) getUserModel(user *user_service.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:        user.ID,
		Email:     user.Email,
		CityID:    user.CityID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Birthdate: user.Birthdate.Format(time.DateOnly),
		Gender:    string(user.Gender),
		Interests: user.Interests,
	}
}
