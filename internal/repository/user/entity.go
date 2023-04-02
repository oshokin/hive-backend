package user

import "time"

type (
	// User represents a user entity in the database.
	User struct {
		ID           int64     // Unique identifier of the user.
		Email        string    // Email address of the user.
		PasswordHash string    // Hashed password of the user.
		CityID       int16     // ID of the city the user lives in.
		FirstName    string    // First name of the user.
		LastName     string    // Last name of the user.
		Birthdate    time.Time // Birthdate of the user.
		Gender       string    // Gender of the user.
		Interests    string    // Interests of the user.
	}

	// LoginData represents the ID and password hash of a user for authentication.
	LoginData struct {
		ID           int64  // Unique identifier of the user.
		PasswordHash string // Hashed password of the user.
	}

	// SearchByNamePrefixesRequest represents a request to search for users by name prefixes.
	SearchByNamePrefixesRequest struct {
		FirstName string // Prefix of the first name to search for.
		LastName  string // Prefix of the last name to search for.
		Limit     uint64 // Maximum number of results to return.
		Cursor    int64  // Cursor position for pagination.
	}

	// SearchByNamePrefixesResponse represents the response to a search by name prefixes request.
	SearchByNamePrefixesResponse struct {
		Items   []*User // List of matching user entities.
		HasNext bool    // Whether there are more results available.
	}
)
