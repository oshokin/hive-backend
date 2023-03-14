package user

import (
	"fmt"
	"time"

	validator "github.com/asaskevich/govalidator"
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

	GenderType string
)

const (
	Male    GenderType = "MALE"
	Female  GenderType = "FEMALE"
	Unknown GenderType = "UNKNOWN"
)

func (u *User) Validate() error {
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

func (u *LoginCredentials) Validate() error {
	if !validator.IsEmail(u.Email) {
		return fmt.Errorf("invalid email format")
	}

	if len(u.Password) == 0 {
		return fmt.Errorf("password is required")
	}

	return nil
}
