//nolint: goimports // because it's goimported already

// Package user provides an interface and implementation of methods for interacting with a user database.
package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oshokin/hive-backend/internal/repository/common"
)

type (
	// Repository interface defines methods for interacting with the user database.
	Repository interface {
		// CheckIfExistByEmails checks if users with the given email addresses already exist in the database.
		// Returns a map of existing email addresses.
		CheckIfExistByEmails(ctx context.Context, emails []string) (map[string]struct{}, error)

		// CheckIfExistsByEmail checks if a user with the given email address already exists in the database.
		CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)

		// Create creates a new user in the database.
		// Returns the ID of the newly created user.
		Create(ctx context.Context, u *User) (int64, error)

		// CreateBatch creates new users in the database.
		// Returns the number of created users.
		CreateBatch(ctx context.Context, users []*User) (int64, error)

		// GetByID returns a user with the given ID.
		GetByID(ctx context.Context, id int64) (*User, error)

		// GetByEmail returns a user with the given email address.
		GetByEmail(ctx context.Context, email string) (*User, error)

		// GetLoginDataByEmail returns login data (email and password hash) for the user with the given email address.
		GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error)

		// SearchByNamePrefixes returns a list of users whose first and last names start with the given prefixes.
		// Returns the number of total results and a slice of users.
		SearchByNamePrefixes(ctx context.Context, req *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error)
	}

	repository struct {
		db *pgxpool.Pool
	}
)

const (
	tableName          = "users"
	columnID           = "id"
	columnEmail        = "email"
	columnPasswordHash = "password_hash"
	columnCityID       = "city_id"
	columnFirstName    = "first_name"
	columnLastName     = "last_name"
	columnBirthdate    = "birthdate"
	columnGender       = "gender"
	columnInterests    = "interests"
)

var insertRows = []string{columnEmail,
	columnPasswordHash,
	columnCityID,
	columnFirstName,
	columnLastName,
	columnBirthdate,
	columnGender,
	columnInterests}

// NewRepository creates a new Repository instance.
func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) CheckIfExistByEmails(ctx context.Context, emails []string) (map[string]struct{}, error) {
	queryBuilder := sq.Select(columnEmail).
		From(tableName).
		Where(sq.Eq{columnEmail: emails}).
		Limit(uint64(len(emails))).
		PlaceholderFormat(sq.Dollar)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %w", err)
	}
	defer rows.Close()

	existingEmails := make(map[string]struct{})

	for rows.Next() {
		var email string

		if err = rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("failed to read query results: %w", err)
		}

		existingEmails[email] = struct{}{}
	}

	return existingEmails, nil
}

func (r *repository) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	sql, args, err := sq.Select(columnID).
		From(tableName).
		Where(sq.Eq{columnEmail: email}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to generate query: %w", err)
	}

	var id int64

	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err == nil {
		return id != 0, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return false, fmt.Errorf("failed to read query results: %w", err)
}

func (r *repository) Create(ctx context.Context, u *User) (int64, error) {
	sql, args, err := sq.Insert(tableName).
		Columns(insertRows...).
		Values(u.Email, u.PasswordHash, u.CityID, u.FirstName, u.LastName, u.Birthdate, u.Gender, u.Interests).
		Suffix(fmt.Sprintf("RETURNING \"%s\"", columnID)).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to generate query: %w", err)
	}

	var id int64

	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to read query results: %w", err)
	}

	return id, nil
}

func (r *repository) CreateBatch(ctx context.Context, users []*User) (int64, error) {
	if len(users) == 0 {
		return 0, nil
	}

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to acquire connection: %w", err)
	}

	defer conn.Release()

	rowSrc := pgx.CopyFromSlice(len(users),
		func(i int) ([]interface{}, error) {
			user := users[i]
			return []interface{}{
				user.Email,
				user.PasswordHash,
				user.CityID,
				user.FirstName,
				user.LastName,
				user.Birthdate,
				user.Gender,
				user.Interests}, nil
		})

	copyCount, err := r.db.CopyFrom(ctx,
		pgx.Identifier{tableName},
		insertRows,
		rowSrc)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	return copyCount, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*User, error) {
	sql, args, err := r.selectDefaultUserFields(1).
		Where(sq.Eq{columnID: id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	return r.scanUser(ctx, sql, args...)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	sql, args, err := r.selectDefaultUserFields(1).
		Where(sq.Eq{columnEmail: email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	return r.scanUser(ctx, sql, args...)
}

func (r *repository) GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error) {
	sql, args, err := sq.Select(columnID,
		columnPasswordHash).
		From(tableName).
		Where(sq.Eq{columnEmail: email}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	var u LoginData

	err = r.db.QueryRow(ctx, sql, args...).Scan(&u.ID,
		&u.PasswordHash)
	if err == nil {
		return &u, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, fmt.Errorf("failed to read query results: %w", err)
}

func (r *repository) SearchByNamePrefixes(ctx context.Context,
	req *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error) {
	var (
		firstName = strings.Join([]string{common.EscapeLike(req.FirstName), "%"}, "")
		lastName  = strings.Join([]string{common.EscapeLike(req.LastName), "%"}, "")
		sortByID  = fmt.Sprintf("%s ASC", columnID)
	)

	selectQB := r.selectDefaultUserFields(req.Limit + 1).
		Where(sq.Like{columnFirstName: firstName}).
		Where(sq.Like{columnLastName: lastName}).
		OrderBy(sortByID)

	if req.Cursor != 0 {
		selectQB = selectQB.Where(sq.Gt{columnID: req.Cursor})
	}

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run select query: %w", err)
	}
	defer rows.Close()

	var (
		users   []*User
		hasNext bool
	)

	for rows.Next() {
		if uint64(len(users)) >= req.Limit {
			hasNext = true
			break
		}

		var user User

		err = rows.Scan(&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.CityID,
			&user.FirstName,
			&user.LastName,
			&user.Birthdate,
			&user.Gender,
			&user.Interests)

		if err != nil {
			return nil, fmt.Errorf("failed to read select query results: %w", err)
		}

		users = append(users, &user)
	}

	return &SearchByNamePrefixesResponse{
		Items:   users,
		HasNext: hasNext,
	}, nil
}

func (r *repository) selectDefaultUserFields(limit uint64) sq.SelectBuilder {
	return sq.Select(columnID,
		columnEmail,
		columnPasswordHash,
		columnCityID,
		columnFirstName,
		columnLastName,
		columnBirthdate,
		columnGender,
		columnInterests).
		From(tableName).
		Limit(limit).
		PlaceholderFormat(sq.Dollar)
}

func (r *repository) scanUser(ctx context.Context, sql string, args ...any) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, sql, args...).Scan(&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CityID,
		&u.FirstName,
		&u.LastName,
		&u.Birthdate,
		&u.Gender,
		&u.Interests)
	if err == nil {
		return &u, nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return nil, fmt.Errorf("failed to read query results: %w", err)
}
