package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teatah/rclone/pkg/token"
	"github.com/teatah/rclone/pkg/user"
)

type DBSessionManager struct {
	mu       *sync.RWMutex
	sessions map[string]*Session
	pgPool   *pgxpool.Pool
}

func NewDBSessionManager(pgPool *pgxpool.Pool) *DBSessionManager {
	return &DBSessionManager{
		mu:       &sync.RWMutex{},
		sessions: make(map[string]*Session, 5),
		pgPool:   pgPool,
	}
}

func (sm *DBSessionManager) Create(ctx context.Context, user *user.User) (*Session, error) {
	sess, err := NewSession(user)

	if err != nil {
		return nil, err
	}

	err = sm.addSession(ctx, sess)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (sm *DBSessionManager) Check(ctx context.Context, tokenString string) (*Session, error) {
	_, err := token.ParseJwt(tokenString)
	if err != nil {
		return nil, err
	}

	sess, err := sm.findSession(ctx, tokenString)

	return sess, err
}

func (sm *DBSessionManager) UsernameBySessionID(ctx context.Context, sessID string) (string, error) {
	var username string

	err := sm.pgPool.QueryRow(
		ctx,
		`SELECT username
		FROM users
		INNER JOIN sessions ON users.id = sessions.user_id
		WHERE sessions.id = $1`,
		sessID,
	).Scan(&username)

	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("session %s not found", sessID)
	}

	return username, err
}

func (sm *DBSessionManager) addSession(ctx context.Context, sess *Session) error {
	_, err := sm.pgPool.Exec(
		ctx,
		"INSERT INTO sessions (id, user_id, expires_at) values ($1, $2, $3)",
		sess.ID,
		sess.UserID,
		sess.ExpiresAt,
	)

	return err
}

func (sm *DBSessionManager) findSession(ctx context.Context, tokenString string) (*Session, error) {
	var sess Session
	err := sm.pgPool.QueryRow(
		ctx,
		"SELECT id, user_id, expires_at FROM sessions WHERE id=$1",
		tokenString,
	).Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session %s not found", tokenString)
	} else if err != nil {
		return nil, err
	}

	return &sess, err
}

func (sm *DBSessionManager) RemoveExpiredSessions(ctx context.Context) error {
	_, err := sm.pgPool.Exec(
		ctx,
		`DELETE FROM sessions
		WHERE expires_at < EXTRACT(EPOCH FROM NOW() AT TIME ZONE 'UTC')`,
	)

	return err
}
