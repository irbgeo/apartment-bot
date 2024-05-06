package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var cityCollection = "city"

func (s *mongoDB) cityCollectionSetting() error {
	_, err := s.db.Collection(cityCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "city_name", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (s *mongoDB) SaveCity(ctx context.Context, c server.City) error {
	f := filter{
		CityName: &c.Name,
	}
	return s.upsert(ctx, cityCollection, f, toMongoCity(c))
}

func (s *mongoDB) Cities(ctx context.Context) ([]server.City, error) {
	resultCh, err := find[city](ctx, s, cityCollection, filter{})
	if err != nil {
		return nil, err
	}

	result := make([]server.City, 0)
	for t := range resultCh {
		result = append(result, toCity(t))
	}

	return result, nil
}
