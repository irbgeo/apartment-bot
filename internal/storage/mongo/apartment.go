package mongo // nolint: dupl

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	apartmentCollection = "apartment"
)

func (s *mongoDB) apartmentCollectionSetting() error {
	_, err := s.db.Collection(apartmentCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{{Key: "location", Value: "2dsphere"}},
		},
	)

	return err
}

func (s *mongoDB) UpdateApartment(ctx context.Context, a server.Apartment) error {
	filter := filter{
		ApartmentID: &a.ID,
	}
	return s.upsert(ctx, apartmentCollection, filter, toMongoApartment(a))
}

func (s *mongoDB) SaveApartment(ctx context.Context, a server.Apartment) error {
	return s.insert(ctx, apartmentCollection, toMongoApartment(a))
}

func (s *mongoDB) Apartment(ctx context.Context, f server.Filter) (<-chan server.Apartment, error) {
	resultCh, err := find[apartment](ctx, s, apartmentCollection, toMongoFilter(f))
	if err != nil {
		return nil, err
	}

	apartmentCh := make(chan server.Apartment)
	go func() {
		defer close(apartmentCh)
		for {
			select {
			case <-ctx.Done():
				return
			case a, ok := <-resultCh:
				if !ok {
					return
				}

				apartmentCh <- toApartment(a)
			}
		}
	}()

	return apartmentCh, nil
}

func (s *mongoDB) ApartmentCount(ctx context.Context, f server.Filter) (int64, error) {
	return s.count(ctx, apartmentCollection, toMongoFilter(f))
}

func (s *mongoDB) DeleteApartment(ctx context.Context, a server.Apartment) error {
	f := server.Filter{
		ApartmentID: &a.ID,
	}
	return s.delete(ctx, apartmentCollection, toMongoFilter(f))
}

func (s *mongoDB) DeleteApartments(ctx context.Context) error {
	return s.deleteAll(ctx, apartmentCollection)
}
