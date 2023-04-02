package user

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	validator "github.com/asaskevich/govalidator"
	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
)

type (
	// User represents a user entity with various attributes
	// such as ID, email, password, name, etc.
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

	// LoginCredentials represents the user's login credentials
	// with an email and password.
	LoginCredentials struct {
		Email    string
		Password string
	}

	// LoginData represents the data required
	// for user login such as ID and password hash.
	LoginData struct {
		ID           int64
		PasswordHash string
	}

	// SearchByNamePrefixesRequest represents a request to search users
	// by their first and last name prefixes.
	SearchByNamePrefixesRequest struct {
		FirstName string
		LastName  string
		Limit     uint64
		Cursor    int64
	}

	// SearchByNamePrefixesResponse represents the response to
	// a request to search users by their first and last name prefixes.
	SearchByNamePrefixesResponse struct {
		Items   []*User
		HasNext bool
	}

	// GenderType represents the gender of a user.
	GenderType string
)

// GenderType can have one of three possible values.
const (
	GenderMale    GenderType = "MALE"
	GenderFemale  GenderType = "FEMALE"
	GenderUnknown GenderType = "UNKNOWN"
)

const maxUsersLimit = 50

func (s *service) getServiceModel(source *user_repo.User) *User {
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

func (s *service) getServiceModels(source []*user_repo.User) []*User {
	result := make([]*User, 0, len(source))

	for _, v := range source {
		sm := s.getServiceModel(v)
		if sm == nil {
			continue
		}

		result = append(result, sm)
	}

	return result
}

func (s *service) getRepoModel(source *User) *user_repo.User {
	if source == nil {
		return nil
	}

	return &user_repo.User{
		Email:        source.Email,
		PasswordHash: source.PasswordHash,
		CityID:       source.CityID,
		FirstName:    source.FirstName,
		LastName:     source.LastName,
		Birthdate:    source.Birthdate,
		Gender:       string(source.Gender),
		Interests:    source.Interests,
	}
}

func (s *service) getRepoModels(source []*User) []*user_repo.User {
	result := make([]*user_repo.User, 0, len(source))

	for _, v := range source {
		rm := s.getRepoModel(v)
		if rm == nil {
			continue
		}

		result = append(result, rm)
	}

	return result
}

// String returns a string representation of the User object.
func (u *User) String() string {
	var sb strings.Builder

	sb.WriteString("user{id=")
	sb.WriteString(strconv.FormatInt(u.ID, 10))
	sb.WriteString(", email=")
	sb.WriteString(u.Email)
	sb.WriteString(", password=")
	sb.WriteString(u.Password)
	sb.WriteString(", password_hash=")
	sb.WriteString(u.PasswordHash)
	sb.WriteString(", city_id=")
	sb.WriteString(strconv.FormatInt(int64(u.CityID), 10))
	sb.WriteString(", first_name=")
	sb.WriteString(u.FirstName)
	sb.WriteString(", last_name=")
	sb.WriteString(u.LastName)
	sb.WriteString(", birthdate=")
	sb.WriteString(u.Birthdate.Format("2006-01-02"))
	sb.WriteString(", gender=")
	sb.WriteString(string(u.Gender))
	sb.WriteString(", interests=")
	sb.WriteString(u.Interests)
	sb.WriteString("}")

	return sb.String()
}

func (u *User) validate() error {
	if u == nil {
		return nil
	}

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

	if u.Gender != GenderMale && u.Gender != GenderFemale && u.Gender != GenderUnknown {
		return fmt.Errorf("invalid gender")
	}

	return nil
}

func (cr *LoginCredentials) validate() error {
	if cr == nil {
		return nil
	}

	if !validator.IsEmail(cr.Email) {
		return fmt.Errorf("invalid email format")
	}

	if len(cr.Password) == 0 {
		return fmt.Errorf("password is required")
	}

	return nil
}

func (r *SearchByNamePrefixesRequest) validate() error {
	if r == nil {
		return nil
	}

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
