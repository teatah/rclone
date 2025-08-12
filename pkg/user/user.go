package user

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       string
	Username string
	password []byte
}

type UserRepo interface {
	Register(context.Context, *UserRequest) (*User, error)
	Login(context.Context, *UserRequest) (*User, error)
	UserByName(context.Context, string) (*User, error)
	UserByID(context.Context, string) (*User, error)
}

func (u *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword(u.password, []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func NewUserWithCredentials(username string, password string) (*User, error) {
	userID := uuid.NewString()

	newUser := &User{
		ID:       userID,
		Username: username,
	}

	err := newUser.setPassword(password)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *User) setPassword(password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.password = hashedPass

	return nil
}
