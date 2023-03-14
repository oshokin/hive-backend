package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/oshokin/hive-backend/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

type (
	Service interface {
		Add(ctx context.Context, u *User) (int64, error)
		CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error)
		IsPasswordCorrect(passwordHash, password string) (bool, error)
	}

	service struct {
		repository user.Repository
	}
)

func NewService(r user.Repository) *service {
	return &service{repository: r}
}

func (s *service) Add(ctx context.Context, serviceUser *User) (int64, error) {
	passwordHash, err := s.hashPassword(serviceUser.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	repoUser := &user.User{
		Email:        serviceUser.Email,
		PasswordHash: string(passwordHash),
		CityID:       serviceUser.CityID,
		FirstName:    serviceUser.FirstName,
		LastName:     serviceUser.LastName,
		Birthdate:    serviceUser.Birthdate,
		Gender:       string(serviceUser.Gender),
		Interests:    serviceUser.Interests,
	}

	userID, err := s.repository.Add(ctx, repoUser)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *service) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.repository.CheckIfExistsByEmail(ctx, email)
}

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	u, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, nil
	}

	return &User{
		ID:        u.ID,
		Email:     u.Email,
		CityID:    u.CityID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Birthdate: u.Birthdate,
		Gender:    GenderType(u.Gender),
		Interests: u.Interests,
	}, nil
}

func (s *service) GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error) {
	v, err := s.repository.GetLoginDataByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, nil
	}

	return &LoginData{
		ID:           v.ID,
		PasswordHash: v.PasswordHash,
	}, nil
}

func (s *service) IsPasswordCorrect(passwordHash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}

func (s *service) hashPassword(password string) ([]byte, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hashBytes, nil
}
