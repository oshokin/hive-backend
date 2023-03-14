package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/render"
	user_service "github.com/oshokin/hive-backend/internal/service/user"
)

type loginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginUserResponse struct {
	Success bool `json:"success"`
}

func (s *server) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var (
		req loginUserRequest
		err = json.NewDecoder(r.Body).Decode(&req)
	)

	if err != nil {
		s.renderError(w, r,
			http.StatusBadRequest,
			fmt.Sprintf("failed to decode request: %s", err.Error()))
		return
	}

	email := req.Email
	creds := &user_service.LoginCredentials{
		Email:    email,
		Password: req.Password,
	}

	if err := creds.Validate(); err != nil {
		s.renderError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()
	loginData, err := s.userService.GetLoginDataByEmail(ctx, email)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to read user info: %s", err.Error()))
		return
	}

	if loginData == nil {
		s.renderError(w, r,
			http.StatusBadRequest, "invalid email or password")
		return
	}

	isPasswordCorrect, err := s.userService.IsPasswordCorrect(loginData.PasswordHash, req.Password)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to check password: %s", err.Error()))
		return
	}

	if !isPasswordCorrect {
		s.renderError(w, r,
			http.StatusBadRequest, "invalid email or password")
		return
	}

	accessToken, err := s.generateAccessToken(loginData.ID)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to generate access token: %s", err.Error()))
		return
	}

	refreshToken, err := s.generateRefreshToken(loginData.ID)
	if err != nil {
		s.renderError(w, r,
			http.StatusInternalServerError,
			fmt.Sprintf("failed to generate refresh token: %s", err.Error()))
		return
	}

	s.cache.Set(refreshToken, loginData.ID, refreshTokenDuration)

	now := time.Now()
	http.SetCookie(w, &http.Cookie{
		Name:    accessTokenCookieName,
		Value:   accessToken,
		Expires: now.Add(accessTokenDuration),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    refreshTokenCookieName,
		Value:   refreshToken,
		Expires: now.Add(refreshTokenDuration),
	})

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &loginUserResponse{
		Success: true,
	})
}
