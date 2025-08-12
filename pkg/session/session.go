package session

import (
	"context"

	"github.com/teatah/rclone/pkg/token"
	"github.com/teatah/rclone/pkg/user"
)

type SessionCtxValue string

type SessionManager interface {
	Create(context.Context, *user.User) (*Session, error)
	Check(ctx context.Context, tokenString string) (*Session, error)
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt int64
}

func NewSession(user *user.User) (*Session, error) {
	tokenClaims := token.TokenClaims{}

	token, err := tokenClaims.CreateJwt(user)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        token,
		UserID:    tokenClaims.User.ID,
		ExpiresAt: tokenClaims.ExpiresAt,
	}, nil
}
