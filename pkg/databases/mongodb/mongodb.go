package mongodb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/teatah/rclone/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, config *config.Config) (*mongo.Client, error) {
	u := url.URL{
		Scheme: "mongodb",
		User:   url.UserPassword(config.MongoDB.User, config.MongoDB.Password),
		Host:   fmt.Sprintf("%s:%s", config.MongoDB.Host, config.MongoDB.Port),
	}
	connString := u.String()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)

	return client, err
}

func SetIndex(ctx context.Context, col *mongo.Collection, field string) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{"votes.user": 1},
	}

	_, err := col.Indexes().CreateOne(ctx, indexModel)

	return err
}
