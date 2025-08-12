package token

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/teatah/rclone/pkg/user"
)

type TokenClaims struct {
	*jwt.StandardClaims

	User UserClaims `json:"user"`
}

type UserClaims struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

var secret = []byte("secret")

func (tc *TokenClaims) CreateJwt(u *user.User) (string, error) {
	issueTime := time.Now()
	expTime := issueTime.Add(time.Minute * 1 / 2).Unix()

	tc.User = UserClaims{
		Username: u.Username,
		ID:       u.ID,
	}
	tc.StandardClaims = &jwt.StandardClaims{
		IssuedAt:  issueTime.Unix(),
		ExpiresAt: expTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error creating jwt: %w", err)
	}

	return tokenString, nil
}

func ParseJwt(token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header)
		}
		return secret, nil
	})

	if err != nil || !t.Valid {
		log.Println("token validation skipped - will be valided according to data from the database")
	}

	return claims, nil
}

func TokenFromHeader(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")
	tokenString = strings.ReplaceAll(tokenString, "Bearer ", "")

	return tokenString
}
