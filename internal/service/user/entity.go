package user

import (
	"fmt"
	"time"

	validator "github.com/asaskevich/govalidator"
	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
)

type (
	User struct {
		ID           int64
		Email        string
		Password     string
		PasswordHash string
		CityID       int16
		FirstName    string
		LastName     string
		Birthdate    time.Time
		Gender       GenderType
		Interests    string
	}

	LoginCredentials struct {
		Email    string
		Password string
	}

	LoginData struct {
		ID           int64
		PasswordHash string
	}

	SearchByNamePrefixesRequest struct {
		FirstName string
		LastName  string
		Limit     uint64
		Cursor    int64
	}

	SearchByNamePrefixesResponse struct {
		Items   []*User
		HasNext bool
	}

	GenderType string
)

const (
	Male    GenderType = "MALE"
	Female  GenderType = "FEMALE"
	Unknown GenderType = "UNKNOWN"
)

const maxUsersLimit = 50

func (s *service) GetServiceModel(source *user_repo.User) *User {
	if source == nil {
		return nil
	}

	return &User{
		ID:        source.ID,
		Email:     source.Email,
		CityID:    source.CityID,
		FirstName: source.FirstName,
		LastName:  source.LastName,
		Birthdate: source.Birthdate,
		Gender:    GenderType(source.Gender),
		Interests: source.Interests,
	}
}

func (s *service) GetServiceModels(source []*user_repo.User) []*User {
	result := make([]*User, 0, len(source))
	for _, v := range source {
		sm := s.GetServiceModel(v)
		if sm == nil {
			continue
		}

		result = append(result, sm)
	}

	return result
}

func (u *User) validate() error {
	if !validator.IsEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	if len(u.Password) == 0 {
		return fmt.Errorf("password is required")
	}

	if u.CityID <= 0 {
		return fmt.Errorf("invalid city ID")
	}

	if len(u.FirstName) == 0 {
		return fmt.Errorf("first name is required")
	}

	if len(u.LastName) == 0 {
		return fmt.Errorf("last name is required")
	}

	if u.Birthdate.IsZero() {
		return fmt.Errorf("birthdate is required")
	}

	if u.Gender != Male && u.Gender != Female && u.Gender != Unknown {
		return fmt.Errorf("invalid gender")
	}

	return nil
}

func (u *LoginCredentials) validate() error {
	if !validator.IsEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	if len(u.Password) == 0 {
		return fmt.Errorf("password is required")
	}

	return nil
}

func (r *SearchByNamePrefixesRequest) validate() error {
	if len(r.FirstName) == 0 {
		return fmt.Errorf("first name is required")
	}

	if len(r.LastName) == 0 {
		return fmt.Errorf("last name is required")
	}

	if r.Limit > maxUsersLimit {
		return fmt.Errorf("limit cannot be greater than %d", maxUsersLimit)
	}

	if r.Cursor < 0 {
		return fmt.Errorf("cursor must be greater than or equal to 0")
	}

	return nil
}
