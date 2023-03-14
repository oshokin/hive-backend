package user

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type (
	Repository interface {
		Add(ctx context.Context, u *User) (int64, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error)
		CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
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

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Add(ctx context.Context, u *User) (int64, error) {
	sql, args, err := sq.Insert(tableName).
		Columns(columnEmail,
			columnPasswordHash,
			columnCityID,
			columnFirstName,
			columnLastName,
			columnBirthdate,
			columnGender,
			columnInterests).
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

func (r *repository) GetByID(ctx context.Context, id int64) (*User, error) {
	sql, args, err := sq.Select(columnID,
		columnEmail,
		columnPasswordHash,
		columnCityID,
		columnFirstName,
		columnLastName,
		columnBirthdate,
		columnGender,
		columnInterests).
		From(tableName).
		Where(sq.Eq{columnID: id}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	u := new(User)
	err = r.db.QueryRow(ctx, sql, args...).Scan(&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CityID,
		&u.FirstName,
		&u.LastName,
		&u.Birthdate,
		&u.Gender,
		&u.Interests)
	if err == nil {
		return u, nil
	}

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return nil, fmt.Errorf("failed to read query results: %w", err)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	sql, args, err := sq.Select(columnID,
		columnEmail,
		columnPasswordHash,
		columnCityID,
		columnFirstName,
		columnLastName,
		columnBirthdate,
		columnGender,
		columnInterests).
		From(tableName).
		Where(sq.Eq{columnEmail: email}).
		Limit(1).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}

	u := new(User)
	err = r.db.QueryRow(ctx, sql, args...).Scan(&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.CityID,
		&u.FirstName,
		&u.LastName,
		&u.Birthdate,
		&u.Gender,
		&u.Interests)
	if err == nil {
		return u, nil
	}

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return nil, fmt.Errorf("failed to read query results: %w", err)
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

	u := new(LoginData)
	err = r.db.QueryRow(ctx, sql, args...).Scan(&u.ID,
		&u.PasswordHash)
	if err == nil {
		return u, nil
	}

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	return nil, fmt.Errorf("failed to read query results: %w", err)
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

	if err == pgx.ErrNoRows {
		return false, nil
	}

	return false, fmt.Errorf("failed to read query results: %w", err)
}
