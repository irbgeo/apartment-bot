package mongo

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const connectURILayout = "mongodb://%s:%s@%s"

type mongoDB struct {
	db *mongo.Database
}

func NewStorage(cfg Config) (*mongoDB, error) {
	username := url.QueryEscape(cfg.Username)
	password := url.QueryEscape(cfg.Password)

	uri := fmt.Sprintf(connectURILayout, username, password, cfg.Address)

	opts := options.Client().ApplyURI(uri)

	// Skip TLS verification for local development
	if strings.Contains(cfg.Address, "443") {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true}) // nolint: gosec
	}

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	m := &mongoDB{
		db: client.Database(cfg.Database),
	}

	return m, nil
}

func (s *mongoDB) insert(ctx context.Context, collectionName string, obj any) error {
	_, err := s.db.Collection(collectionName).InsertOne(ctx, obj, options.InsertOne())

	return err
}

func (s *mongoDB) upsert(ctx context.Context, collectionName string, f filter, obj any) error {
	_, err := s.db.Collection(collectionName).UpdateOne(ctx, f.forCollection(collectionName), bson.M{"$set": obj}, options.Update().SetUpsert(true))

	return err
}

func find[T any](ctx context.Context, m *mongoDB, collectionName string, f filter) (<-chan T, error) {
	cur, err := m.db.Collection(collectionName).Find(ctx, f.forCollection(collectionName))
	if err != nil {
		return nil, err
	}

	resultCh := make(chan T)
	go func() {
		defer close(resultCh)
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var storageDocument T

			if err := cur.Decode(&storageDocument); err != nil {
				slog.Error("decode", "collection name", collectionName, "err", err)
			}
			resultCh <- storageDocument
		}
	}()

	return resultCh, nil
}

func (s *mongoDB) count(ctx context.Context, collectionName string, f filter) (int64, error) {
	cur, err := s.db.Collection(collectionName).CountDocuments(ctx, f.forCollection(collectionName))
	if err != nil {
		return 0, err
	}
	return cur, nil
}

func (s *mongoDB) delete(ctx context.Context, collectionName string, f filter) error {
	_, err := s.db.Collection(collectionName).DeleteOne(ctx, f.forCollection(collectionName))
	return err
}

func (s *mongoDB) deleteAll(ctx context.Context, collectionName string) error {
	_, err := s.db.Collection(collectionName).DeleteMany(ctx, bson.M{})
	return err
}
