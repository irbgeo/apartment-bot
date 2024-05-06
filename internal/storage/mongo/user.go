package mongo

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/irbgeo/apartment-bot/internal/message"
	"github.com/irbgeo/apartment-bot/internal/server"
)

var userCollection = "user"

func (s *mongoDB) InsertUser(ctx context.Context, u server.User) error {
	f := server.Filter{
		User: &server.User{
			ID: u.ID,
		},
	}

	err := s.upsert(ctx, userCollection, toMongoFilter(f), toMongoUser(u))
	if mongo.IsDuplicateKeyError(err) {
		slog.Error("user already exists", "user_id", u.ID)
		return nil
	}
	return err
}

func (s *mongoDB) DeleteUser(ctx context.Context, u server.User) error {
	f := server.Filter{
		User: &server.User{
			ID: u.ID,
		},
	}
	return s.delete(ctx, userCollection, toMongoFilter(f))
}

func (s *mongoDB) Users(ctx context.Context) (<-chan message.User, error) {
	resultCh, err := find[user](ctx, s, userCollection, filter{})
	if err != nil {
		return nil, err
	}

	userCh := make(chan message.User)
	go func() {
		defer close(userCh)
		for {
			select {
			case <-ctx.Done():
				return
			case u, ok := <-resultCh:
				if !ok {
					return
				}
				userCh <- toMessageUser(u)
			}
		}
	}()

	return userCh, nil
}

func (s *mongoDB) User(ctx context.Context, f server.Filter) (server.User, error) {
	resultCh, err := find[user](ctx, s, userCollection, toMongoFilter(f))
	if err != nil {
		return server.User{}, err
	}

	select {
	case <-ctx.Done():
		return server.User{}, ctx.Err()
	case u, ok := <-resultCh:
		if !ok {
			return server.User{}, errNotFound
		}
		return toserverUser(u), nil
	}
}
