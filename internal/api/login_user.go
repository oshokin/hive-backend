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
			common.NewError(common.ErrStatusBadRequest,
				fmt.Errorf("failed to decode request: %w", err)))

		return
	}

	var (
		ctx   = r.Context()
		creds = &user_service.LoginCredentials{
			Email:    req.Email,
			Password: req.Password,
		}
	)

	userID, err := s.userService.GetIDByLoginCredentials(ctx, creds)
	if err != nil {
		var e *common.Error
		if errors.As(err, &e) {
			s.renderError(w, r, e)
		} else {
			s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
				fmt.Errorf("failed to login user: %w", err)))
		}

		return
	}

	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to generate access token: %w", err)))

		return
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		s.renderError(w, r, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to generate refresh token: %w", err)))

		return
	}

	s.cache.Set(refreshToken, userID, refreshTokenDuration)
	s.setAuthorizationCookies(w, accessToken, refreshToken)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &loginUserResponse{
		Success: true,
	})
}

func (s *server) setAuthorizationCookies(w http.ResponseWriter, accessToken, refreshToken string) {
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
}
