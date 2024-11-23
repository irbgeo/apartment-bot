package mongo // nolint: dupl

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var filterCollection = "filter"

func (s *mongoDB) filterCollectionSetting() error {
	_, err := s.db.Collection(filterCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (s *mongoDB) SaveFilter(ctx context.Context, f server.Filter) error {
	filter := toMongoFilter(f)
	filter.Name = nil

	newFilter := toMongoFilter(f)
	return s.upsert(ctx, filterCollection, filter, newFilter)
}

func (s *mongoDB) Filters(ctx context.Context, f server.Filter) ([]server.Filter, error) {
	resultCh, err := find[filter](ctx, s, filterCollection, toMongoFilter(f))
	if err != nil {
		return nil, err
	}

	result := make([]server.Filter, 0)
	for t := range resultCh {
		result = append(result, toFilter(t))
	}

	return result, nil
}

func (s *mongoDB) DeleteFilter(ctx context.Context, f server.Filter) error {
	return s.delete(ctx, filterCollection, toMongoFilter(f))
}
