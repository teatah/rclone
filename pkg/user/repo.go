package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserAlreadyExists = errors.New("already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
)

type UserDBRepo struct {
	pgPool *pgxpool.Pool
}

func NewUserDBRepo(pgPool *pgxpool.Pool) *UserDBRepo {
	return &UserDBRepo{
		pgPool: pgPool,
	}
}

func (ur *UserDBRepo) Register(ctx context.Context, userRequest *UserRequest) (*User, error) {
	username := userRequest.Username
	password := userRequest.Password

	newUser, err := NewUserWithCredentials(username, password)
	if err != nil {
		return nil, err
	}

	err = ur.addUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (ur *UserDBRepo) Login(ctx context.Context, userRequest *UserRequest) (*User, error) {
	user, err := ur.UserByName(ctx, userRequest.Username)
	if err != nil {
		return nil, err
	}

	err = user.CheckPassword(userRequest.Password)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	return user, nil
}

func (ur *UserDBRepo) UserByName(ctx context.Context, username string) (*User, error) {
	var user User
	err := ur.pgPool.QueryRow(
		ctx,
		"SELECT id, username, password FROM users WHERE username=$1",
		username,
	).Scan(&user.ID, &user.Username, &user.password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (ur *UserDBRepo) UserByID(ctx context.Context, userID string) (*User, error) {
	var user User
	err := ur.pgPool.QueryRow(
		ctx,
		"SELECT id, username, password FROM users WHERE id=$1",
		userID,
	).Scan(&user.ID, &user.Username, &user.password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (ur *UserDBRepo) addUser(ctx context.Context, user *User) error {
	_, err := ur.pgPool.Exec(
		ctx,
		"INSERT INTO users (id, username, password) values ($1, $2, $3)",
		user.ID,
		user.Username,
		user.password,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return ErrUserAlreadyExists
		}
	}

	return err
}
