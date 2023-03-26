package user

import "time"

type (
	User struct {
		ID           int64
		Email        string
		PasswordHash string
		CityID       int16
		FirstName    string
		LastName     string
		Birthdate    time.Time
		Gender       string
		Interests    string
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
)
