package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type getUserResponse struct {
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
			http.StatusBadRequest,
			fmt.Sprintf("failed to parse user ID: %s", err.Error()))
		return
	}

	if userID <= 0 {
		s.renderError(w, r,
			http.StatusBadRequest, "user ID must be greater than 0")
		return
	}

	user, err := s.userService.GetByID(r.Context(), userID)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to read user info: %s", err.Error()))
		return
	}

	if user == nil {
		s.renderError(w, r,
			http.StatusBadRequest, "user not found")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, s.fillGetUserResponse(user))
}

func (s *server) fillGetUserResponse(user *user_service.User) *getUserResponse {
	if user == nil {
		return nil
	}

	return &getUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CityID:    user.CityID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Birthdate: user.Birthdate.Format("2006-01-02"),
		Gender:    string(user.Gender),
		Interests: user.Interests,
	}
}
