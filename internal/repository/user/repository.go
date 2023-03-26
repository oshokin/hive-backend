package user

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oshokin/hive-backend/internal/repository/common"
)

type (
	Repository interface {
		Add(ctx context.Context, u *User) (int64, error)
		CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetLoginDataByEmail(ctx context.Context, email string) (*LoginData, error)
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

func (r *repository) SearchByNamePrefixes(ctx context.Context, req *SearchByNamePrefixesRequest) (*SearchByNamePrefixesResponse, error) {
	var (
		firstName = strings.Join([]string{common.EscapeLike(req.FirstName), "%"}, "")
		lastName  = strings.Join([]string{common.EscapeLike(req.LastName), "%"}, "")
	)

	selectQB := sq.StatementBuilder.
		Select(columnID,
			columnEmail,
			columnPasswordHash,
			columnCityID,
			columnFirstName,
			columnLastName,
			columnBirthdate,
			columnGender,
			columnInterests).
		From(tableName).
		Where(sq.Like{columnFirstName: firstName}).
		Where(sq.Like{columnLastName: lastName}).
		OrderBy(fmt.Sprintf("%s ASC", columnID)).
		Limit(req.Limit).
		PlaceholderFormat(sq.Dollar)

	if req.Cursor != 0 {
		selectQB = selectQB.Where(sq.Gt{columnID: req.Cursor})
	}

	selectQuery, selectArgs, err := selectQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	statsQB := sq.StatementBuilder.
		Select(fmt.Sprintf("COUNT(%s)", columnID)).
		From(tableName).
		Where(sq.Like{columnFirstName: firstName}).
		Where(sq.Like{columnLastName: lastName}).
		Limit(req.Limit + 1).
		PlaceholderFormat(sq.Dollar)

	if req.Cursor != 0 {
		statsQB = statsQB.Where(sq.Gt{columnID: req.Cursor})
	}

	statsQuery, statsArgs, err := statsQB.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build stats query: %w", err)
	}

	rows, err := r.db.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to run select query: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var (
			user User
			err  = rows.Scan(&user.ID,
				&user.Email,
				&user.PasswordHash,
				&user.CityID,
				&user.FirstName,
				&user.LastName,
				&user.Birthdate,
				&user.Gender,
				&user.Interests)
		)

		if err != nil {
			return nil, fmt.Errorf("failed to read select query results: %w", err)
		}

		users = append(users, &user)
	}

	var (
		countRow   = r.db.QueryRow(ctx, statsQuery, statsArgs...)
		statsCount uint64
	)

	err = countRow.Scan(&statsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read stats query results: %w", err)
	}

	return &SearchByNamePrefixesResponse{
		Items:   users,
		HasNext: statsCount > uint64(len(users)),
	}, nil
}
