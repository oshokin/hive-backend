package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/oshokin/hive-backend/internal/service/common"
)

type (
	UserClaims struct {
		UserID int64 `json:"user_id"`
		jwt.RegisteredClaims
	}

	userIDType string
)

const (
	cacheExpirationTime               = 15 * time.Minute
	cacheCleanupInterval              = 5 * time.Minute
	accessTokenDuration               = 15 * time.Minute
	refreshTokenDuration              = 24 * time.Hour
	accessTokenCookieName             = "access_token"
	refreshTokenCookieName            = "refresh_token"
	userIDHeader           userIDType = "user_id"
)

var errAccessDenied = common.NewError(common.ErrStatusUnauthorized, errors.New("access denied"))

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessTokenCookie, err := r.Cookie(accessTokenCookieName)
		if err != nil {
			s.renderError(w, r, errAccessDenied)
			return
		}

		refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
		if err != nil {
			s.renderError(w, r, errAccessDenied)
			return
		}

		accessToken := accessTokenCookie.Value
		accesClaims, err := s.verifyAccessToken(accessToken)
		if err != nil {
			s.renderError(w, r, common.NewError(common.ErrStatusUnauthorized, err))
			return
		}

		refreshToken := refreshTokenCookie.Value
		refreshClaims, err := s.verifyRefreshToken(refreshToken)
		if err != nil {
			s.cache.Delete(refreshToken)
			s.renderError(w, r, common.NewError(common.ErrStatusUnauthorized, err))
			return
		}

		userID, isFound := s.cache.Get(refreshToken)
		if !isFound {
			s.renderError(w, r, errAccessDenied)
			return
		}

		if userID != accesClaims.UserID || accesClaims.UserID != refreshClaims.UserID {
			s.renderError(w, r, errAccessDenied)
			return
		}

		ctx := r.Context()
		r = r.WithContext(context.WithValue(ctx, userIDHeader, accesClaims.UserID))

		next.ServeHTTP(w, r)
	})
}

func (s *server) generateAccessToken(userID int64) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   accessTokenCookieName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecretKey)
}

func (s *server) generateRefreshToken(userID int64) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   refreshTokenCookieName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecretKey)
}

func (s *server) verifyAccessToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecretKey, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid || claims.Subject != "access_token" {
		return nil, fmt.Errorf("invalid access token")
	}

	return claims, nil
}

func (s *server) verifyRefreshToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.jwtSecretKey, nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %v", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid || claims.Subject != refreshTokenCookieName {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}
