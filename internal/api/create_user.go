package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"
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
			http.StatusBadRequest,
			fmt.Sprintf("failed to decode request: %s", err.Error()))
		return
	}

	var (
		ctx    = r.Context()
		cityID = req.CityID
	)

	city, err := s.cityService.GetByID(ctx, cityID)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to check if city exists by ID: %s", err.Error()))
		return
	}

	if city == nil {
		s.renderError(w, r,
			http.StatusBadRequest,
			fmt.Sprintf("city with ID %d is not found", cityID))
		return
	}

	var (
		email = req.Email
		user  = &user_service.User{
			Email:     email,
			Password:  req.Password,
			CityID:    cityID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Birthdate: time.Time(req.Birthdate),
			Gender:    user_service.GenderType(req.Gender),
			Interests: req.Interests,
		}
	)

	if err := user.Validate(); err != nil {
		s.renderError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	userExists, err := s.userService.CheckIfExistsByEmail(ctx, email)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to check if user exists by e-mail: %s", err.Error()))
		return
	}

	if userExists {
		s.renderError(w, r, http.StatusConflict, "email is already taken")
		return
	}

	userID, err := s.userService.Add(ctx, user)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to register user: %s", err.Error()))
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &createUserResponse{
		UserID: userID,
	})
}
