package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	gofakeit "github.com/brianvoe/gofakeit/v6"
	"github.com/oshokin/hive-backend/internal/common"
	user_repo "github.com/oshokin/hive-backend/internal/repository/user"
	city_service "github.com/oshokin/hive-backend/internal/service/city"
	common_service "github.com/oshokin/hive-backend/internal/service/common"
	rus_name_gen "github.com/oshokin/russian-name-generator"
	"golang.org/x/crypto/bcrypt"
)

type (
	// Service defines the methods to manage users.
	Service interface {
		// Create a user with given user data.
		Create(ctx context.Context, u *User) (int64, error)
		// Create a batch of users with given user data.
		// Returns the number of created users, a map of user validation errors (if any),
		// and any error that occurred.
		CreateBatch(ctx context.Context, sourceList []*User) (int64, map[*User]error, error)
		// Generate random user data and create users with it.
		GenerateRandomData(ctx context.Context, count int64) ([]*User, error)
		// Get a user by ID.
		GetByID(ctx context.Context, id int64) (*User, error)
		// Get a user's ID by their login credentials.
		GetIDByLoginCredentials(ctx context.Context, creds *LoginCredentials) (int64, error)
		// Search for users by name prefixes.
		SearchByNamePrefixes(ctx context.Context, req *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error)
	}

	service struct {
		userRepository user_repo.Repository
		cityService    city_service.Service
	}
)

const (
	minAgeOfRandomUser = 10
	maxAgeOfRandomUser = 75
)

var (
	errEmailIsAlreadyTaken = common_service.NewError(common_service.ErrStatusConflict,
		errors.New("email is already taken"))
	errInvalidUserID = common_service.NewError(common_service.ErrStatusBadRequest,
		errors.New("user ID must be greater than 0"))
	errUserNotFound = common_service.NewError(common_service.ErrStatusNotFound,
		errors.New("user not found"))
	errInvalidCredentials = common_service.NewError(common_service.ErrStatusBadRequest,
		errors.New("invalid email or password"))
)

// NewService returns a new instance of the user service.
func NewService(r user_repo.Repository, c city_service.Service) Service {
	return &service{
		userRepository: r,
		cityService:    c,
	}
}

func (s *service) Create(ctx context.Context, u *User) (int64, error) {
	email := u.Email

	if err := u.validate(); err != nil {
		return 0, common_service.NewError(common_service.ErrStatusBadRequest, err)
	}

	cityID := u.CityID

	city, err := s.cityService.GetByID(ctx, cityID)
	if err != nil {
		return 0, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to check if city exists by ID: %w", err))
	}

	if city == nil {
		return 0, common_service.NewError(common_service.ErrStatusBadRequest,
			fmt.Errorf("city with ID %d is not found", cityID))
	}

	userExists, err := s.userRepository.CheckIfExistsByEmail(ctx, email)
	if err != nil {
		return 0, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to check if user exists by e-mail: %w", err))
	}

	if userExists {
		return 0, errEmailIsAlreadyTaken
	}

	passwordHash, err := s.hashPassword(u.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = string(passwordHash)

	userID, err := s.userRepository.Create(ctx, s.getRepoModel(u))
	if err != nil {
		return 0, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to create user: %w", err))
	}

	return userID, nil
}

func (s *service) CreateBatch(ctx context.Context, sourceList []*User) (int64, map[*User]error, error) {
	validList, validationErrors, err := s.validateBatch(ctx, sourceList)
	if err != nil {
		return 0, validationErrors, common_service.NewError(common_service.ErrStatusBadRequest, err)
	}

	if len(validList) == 0 {
		return 0, validationErrors, nil
	}

	createdCount, err := s.userRepository.CreateBatch(ctx, s.getRepoModels(validList))
	if err != nil {
		return 0, nil, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to create users: %w", err))
	}

	return createdCount, validationErrors, nil
}

func (s *service) GenerateRandomData(ctx context.Context, count int64) ([]*User, error) {
	cities, err := s.cityService.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}

	var (
		personFields = &rus_name_gen.PersonFields{
			Name:    true,
			Surname: true,
			Gender:  rus_name_gen.GenderAny,
		}
	)

	const maxEmptyIterationsCount = 100

	var (
		now                  = time.Now()
		end                  = now.AddDate(-minAgeOfRandomUser, 0, 0)
		start                = end.AddDate(-maxAgeOfRandomUser, 0, 0)
		usedEmails           = make(map[string]struct{}, count)
		users                = make([]*User, 0, count)
		idx                  int64
		emptyIterationsCount int64
	)

	for {
		if idx >= count || emptyIterationsCount >= maxEmptyIterationsCount {
			break
		}

		var (
			person    = rus_name_gen.Person(personFields)
			birthdate = gofakeit.DateRange(start, end)
			domain    = common.GetRandItemFromList(domains)
			email     = strings.Join([]string{
				rus_name_gen.Transliterate(strings.ToLower(person.Name)),
				"-",
				rus_name_gen.Transliterate(strings.ToLower(person.Surname)),
				"-",
				strconv.FormatInt(int64(birthdate.Year()), 10),
				"@",
				domain}, "")
		)

		if _, ok := usedEmails[email]; ok {
			emptyIterationsCount++
			continue
		}

		usedEmails[email] = struct{}{}

		var (
			city   = common.GetRandItemFromList(cities)
			gender = GenderMale
		)

		if person.Gender.IsFeminine() {
			gender = GenderFemale
		}

		users = append(users, &User{
			Email:     email,
			Password:  email,
			CityID:    city.ID,
			FirstName: person.Name,
			LastName:  person.Surname,
			Birthdate: birthdate,
			Gender:    gender,
			Interests: gofakeit.Hobby(),
		})

		idx++
	}

	return users, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, errInvalidUserID
	}

	u, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to read user info: %w", err))
	}

	if u == nil {
		return nil, errUserNotFound
	}

	return s.getServiceModel(u), nil
}

func (s *service) GetIDByLoginCredentials(ctx context.Context, creds *LoginCredentials) (int64, error) {
	if err := creds.validate(); err != nil {
		return 0, common_service.NewError(common_service.ErrStatusBadRequest, err)
	}

	loginData, err := s.userRepository.GetLoginDataByEmail(ctx, creds.Email)
	if err != nil {
		return 0, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to read user info: %w", err))
	}

	if loginData == nil {
		return 0, errInvalidCredentials
	}

	isPasswordCorrect, err := s.isPasswordCorrect(loginData.PasswordHash, creds.Password)
	if err != nil {
		return 0, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to check password: %w", err))
	}

	if !isPasswordCorrect {
		return 0, errInvalidCredentials
	}

	return loginData.ID, nil
}

func (s *service) SearchByNamePrefixes(ctx context.Context,
	r *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error) {
	if err := r.validate(); err != nil {
		return nil, common_service.NewError(common_service.ErrStatusBadRequest, err)
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
		Items:   s.getServiceModels(res.Items),
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

func (s *service) validateBatch(ctx context.Context, sourceList []*User) ([]*User, map[*User]error, error) {
	var (
		validList        = make([]*User, 0, len(sourceList))
		validationErrors = make(map[*User]error, len(sourceList))
		emails           = make([]string, 0, len(sourceList))
		repeatedEmails   = make(map[string]struct{}, len(sourceList))
		cityIDs          = make([]int16, 0, len(sourceList))
		repeatedCityIDs  = make(map[int16]struct{}, len(sourceList))
	)

	for _, u := range sourceList {
		if u == nil {
			continue
		}

		if err := u.validate(); err != nil {
			validationErrors[u] = common_service.NewError(common_service.ErrStatusBadRequest, err)
			continue
		}

		email := u.Email
		if _, ok := repeatedEmails[email]; !ok {
			emails = append(emails, email)
			repeatedEmails[email] = struct{}{}
		}

		cityID := u.CityID
		if _, ok := repeatedCityIDs[cityID]; !ok {
			cityIDs = append(cityIDs, cityID)
			repeatedCityIDs[cityID] = struct{}{}
		}

		validList = append(validList, u)
	}

	if len(validList) == 0 {
		return validList, nil, nil
	}

	existingEmails, err := s.userRepository.CheckIfExistByEmails(ctx, emails)
	if err != nil {
		return nil, nil, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to check if users exist by email: %w", err))
	}

	existingCityIDs, err := s.cityService.CheckIfExistByIDs(ctx, cityIDs)
	if err != nil {
		return nil, nil, common_service.NewError(common_service.ErrStatusInternalError,
			fmt.Errorf("failed to check if cities exist by ID: %w", err))
	}

	var i int

	for _, u := range validList {
		email := u.Email
		if _, ok := existingEmails[email]; ok {
			validationErrors[u] = errEmailIsAlreadyTaken
			continue
		}

		cityID := u.CityID
		if _, ok := existingCityIDs[cityID]; !ok {
			validationErrors[u] = common_service.NewError(common_service.ErrStatusBadRequest,
				fmt.Errorf("city with ID %d is not found", cityID))
			continue
		}

		var passwordHash []byte

		passwordHash, err = s.hashPassword(u.Password)
		if err != nil {
			validationErrors[u] = common_service.NewError(common_service.ErrStatusBadRequest,
				fmt.Errorf("failed to hash password: %w", err))
			continue
		}

		u.PasswordHash = string(passwordHash)

		validList[i] = u
		i++
	}

	for j := i; j < len(validList); j++ {
		validList[j] = nil
	}

	validList = validList[:i]

	return validList, validationErrors, nil
}
