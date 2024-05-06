package mongo

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/irbgeo/apartment-bot/internal/server"
)

func toMongoCity(in server.City) bson.M {
	city := make(bson.M)
	city["city_name"] = in.Name
	for d := range in.District {
		city["district."+d] = struct{}{}
	}

	return city
}

func toCity(in city) server.City {
	return server.City{
		Name:     in.CityName,
		District: in.District,
	}
}
