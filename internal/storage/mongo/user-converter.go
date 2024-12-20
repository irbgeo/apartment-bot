package mongo

import (
	"github.com/irbgeo/apartment-bot/internal/server"
)

func toMongoUser(in server.User) user {
	return user{
		ID:          in.ID,
		ClientID:    in.ClientID,
		IsSuperuser: in.IsSuperuser,
	}
}

func toserverUser(in user) server.User {
	return server.User{
		ID:          in.ID,
		ClientID:    in.ClientID,
		IsSuperuser: in.IsSuperuser,
	}
}
