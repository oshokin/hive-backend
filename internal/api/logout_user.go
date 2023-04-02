package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type logoutUserResponse struct {
	Success bool `json:"success"`
}

func (s *server) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID := ctx.Value(userIDHeader)
	if userID == 0 {
		s.renderError(w, r, errAccessDenied)
		return
	}

	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		s.renderError(w, r, errAccessDenied)
		return
	}

	refreshToken := refreshTokenCookie.Value

	claims, err := s.verifyRefreshToken(refreshToken)
	if err != nil {
		s.cache.Delete(refreshToken)
		s.renderError(w, r, common.NewError(common.ErrStatusUnauthorized, err))

		return
	}

	if claims.UserID != userID {
		s.renderError(w, r, errAccessDenied)
		return
	}

	s.cache.Delete(refreshToken)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &logoutUserResponse{
		Success: true,
	})
}
