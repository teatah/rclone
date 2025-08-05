package session

import (
	"context"

	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/token"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
)

type SessionCtxValue string

type SessionManager interface {
	Create(context.Context, *user.User) (*Session, error)
	Check(ctx context.Context, tokenString string) (*Session, error)
}

type Session struct {
	ID          string
	UserId      string
	ExpiresAt   int64
	TokenClaims *token.TokenClaims
}

func NewSession(user *user.User) (*Session, error) {
	tokenClaims := token.TokenClaims{}

	token, err := tokenClaims.CreateJwt(user)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:          token,
		UserId:      tokenClaims.User.ID,
		ExpiresAt:   tokenClaims.ExpiresAt,
		TokenClaims: &tokenClaims,
	}, nil
}
