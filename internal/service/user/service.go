package user

import (
	"context"
	"errors"
	"fmt"

	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	"github.com/oshokin/hive-backend/internal/service/common"
	"golang.org/x/crypto/bcrypt"
)

type (
	Service interface {
		Add(ctx context.Context, u *User) (int64, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		GetIDByLoginCredentials(ctx context.Context, creds *LoginCredentials) (int64, error)
		SearchByNamePrefixes(ctx context.Context, req *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error)
	}

	service struct {
		userRepository user_repo.Repository
		cityService    city_service.Service
	}
)

var (
	errEmailIsAlreadyTaken = common.NewError(common.ErrStatusConflict,
		errors.New("email is already taken"))
	errInvalidUserID = common.NewError(common.ErrStatusBadRequest,
		errors.New("user ID must be greater than 0"))
	errUserNotFound = common.NewError(common.ErrStatusNotFound,
		errors.New("user not found"))
	errInvalidCredentials = common.NewError(common.ErrStatusBadRequest,
		errors.New("invalid email or password"))
)

func NewService(r user_repo.Repository, c city_service.Service) *service {
	return &service{
		userRepository: r,
		cityService:    c,
	}
}

func (s *service) Add(ctx context.Context, u *User) (int64, error) {
	email := u.Email
	if err := u.validate(); err != nil {
		return 0, common.NewError(common.ErrStatusBadRequest, err)
	}

	cityID := u.CityID
	city, err := s.cityService.GetByID(ctx, cityID)
	if err != nil {
		return 0, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to check if city exists by ID: %w", err))
	}

	if city == nil {
		return 0, common.NewError(common.ErrStatusBadRequest,
			fmt.Errorf("city with ID %d is not found", cityID))
	}

	userExists, err := s.userRepository.CheckIfExistsByEmail(ctx, email)
	if err != nil {
		return 0, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to check if user exists by e-mail: %w", err))
	}

	if userExists {
		return 0, errEmailIsAlreadyTaken
	}

	passwordHash, err := s.hashPassword(u.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	repoUser := &user_repo.User{
		Email:        u.Email,
		PasswordHash: string(passwordHash),
		CityID:       u.CityID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Birthdate:    u.Birthdate,
		Gender:       string(u.Gender),
		Interests:    u.Interests,
	}

	userID, err := s.userRepository.Add(ctx, repoUser)
	if err != nil {
		return 0, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to create user: %w", err))
	}

	return userID, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, errInvalidUserID
	}

	u, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to read user info: %w", err))
	}

	if u == nil {
		return nil, errUserNotFound
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

func (s *service) GetIDByLoginCredentials(ctx context.Context, creds *LoginCredentials) (int64, error) {
	if err := creds.validate(); err != nil {
		return 0, common.NewError(common.ErrStatusBadRequest, err)
	}

	loginData, err := s.userRepository.GetLoginDataByEmail(ctx, creds.Email)
	if err != nil {
		return 0, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to read user info: %w", err))
	}

	if loginData == nil {
		return 0, errInvalidCredentials
	}

	isPasswordCorrect, err := s.isPasswordCorrect(loginData.PasswordHash, creds.Password)
	if err != nil {
		return 0, common.NewError(common.ErrStatusInternalError,
			fmt.Errorf("failed to check password: %w", err))
	}

	if !isPasswordCorrect {
		return 0, errInvalidCredentials
	}

	return loginData.ID, nil
}

func (s *service) SearchByNamePrefixes(ctx context.Context, r *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error) {
	if err := r.validate(); err != nil {
		return nil, common.NewError(common.ErrStatusBadRequest, err)
	}

	limit := r.Limit
	if limit == 0 {
		limit = maxUsersLimit
	}

	res, err := s.userRepository.SearchByNamePrefixes(ctx, &user_repo.SearchByNamePrefixesRequest{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Limit:     limit,
		Cursor:    r.Cursor,
	})

	if err != nil {
		return nil, err
	}

	return &SearchByNamePrefixesResponse{
		Items:   s.GetServiceModels(res.Items),
		HasNext: res.HasNext,
	}, nil
}

func (s *service) hashPassword(password string) ([]byte, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hashBytes, nil
}

func (s *service) isPasswordCorrect(passwordHash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}
