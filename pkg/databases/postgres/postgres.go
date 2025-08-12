package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teatah/rclone/pkg/config"
)

func ConnectPool(ctx context.Context, config *config.Config) (*pgxpool.Pool, error) {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.PostgresDB.User, config.PostgresDB.Password),
		Host:   fmt.Sprintf("%s:%s", config.PostgresDB.Host, config.PostgresDB.Port),
		Path:   config.PostgresDB.Name,
	}
	connString := u.String()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
