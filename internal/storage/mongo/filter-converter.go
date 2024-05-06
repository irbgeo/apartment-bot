package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/irbgeo/apartment-bot/internal/server"
)

func toMongoFilter(in server.Filter) filter {
	out := filter{
		ID:             in.ID,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		ApartmentID:    in.ApartmentID,
		Name:           in.Name,
		District:       in.District,
		CityName:       in.City,
		MinPrice:       in.MinPrice,
		MaxPrice:       in.MaxPrice,
		MinRooms:       in.MinRooms,
		MaxRooms:       in.MaxRooms,
		MinArea:        in.MinArea,
		MaxArea:        in.MaxArea,
		IsOwner:        in.IsOwner,
		MaxDistance:    in.MaxDistance,
		FromTimestamp:  in.FromTimestamp,

		PauseTimestamp: in.PauseTimestamp,
	}

	if in.User != nil {
		out.UserID = &in.User.ID
	}

	if in.Coordinates != nil {
		out.Coordinates = &coordinates{
			Lat: in.Coordinates.Lat,
			Lng: in.Coordinates.Lng,
		}
	}
	return out
}

func toFilter(in filter) server.Filter {
	out := server.Filter{
		ID:             in.ID,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Name:           in.Name,
		District:       in.District,
		City:           in.CityName,
		MinPrice:       in.MinPrice,
		MaxPrice:       in.MaxPrice,
		MinRooms:       in.MinRooms,
		MaxRooms:       in.MaxRooms,
		MinArea:        in.MinArea,
		MaxArea:        in.MaxArea,
		IsOwner:        in.IsOwner,
		MaxDistance:    in.MaxDistance,

		PauseTimestamp: in.PauseTimestamp,
	}

	if in.UserID != nil {
		out.User = &server.User{
			ID: *in.UserID,
		}
	}

	if in.Coordinates != nil {
		out.Coordinates = &server.Coordinates{
			Lat: in.Coordinates.Lat,
			Lng: in.Coordinates.Lng,
		}
	}

	return out
}

func (s *filter) forCollection(collectionName string) any {
	switch collectionName {
	case userCollection:
		return s.user()
	case filterCollection:
		return s.filter()
	case cityCollection:
		return s.city()
	}
	return s.apartment()
}
func (s *filter) apartment() any {
	filter := bson.D{}

	if s.ApartmentID != nil {
		filter = append(filter, bson.E{
			Key:   "_id",
			Value: *s.ApartmentID,
		})
	}

	if s.AdType != nil {
		filter = append(filter, bson.E{
			Key:   "ad_type",
			Value: *s.AdType,
		})
	}

	if s.BuildingStatus != nil {
		filter = append(filter, bson.E{
			Key:   "building_status",
			Value: *s.BuildingStatus,
		})
	}

	if len(s.District) != 0 {
		districts := make([]string, 0, len(s.District)+1)
		for d := range s.District {
			districts = append(districts, d)
		}
		districts = append(districts, "")

		filter = append(filter, bson.E{
			Key:   "district",
			Value: bson.D{{Key: "$in", Value: districts}},
		})
	}

	if s.CityName != nil {
		filter = append(filter, bson.E{
			Key: "city",
			Value: bson.D{
				{
					Key:   "$in",
					Value: []string{*s.CityName, ""},
				},
			},
		})
	}

	price := bson.D{}
	if s.MinPrice != nil {
		price = append(price, bson.E{Key: "$gte", Value: *s.MinPrice})
	}

	if s.MaxPrice != nil {
		price = append(price, bson.E{Key: "$lte", Value: *s.MaxPrice})
	}

	if len(price) != 0 {
		filter = append(filter, bson.E{Key: "price", Value: price})
	}

	rooms := bson.D{}
	if s.MinRooms != nil {
		rooms = append(rooms, bson.E{Key: "$gte", Value: *s.MinRooms})
	}

	if s.MaxRooms != nil {
		rooms = append(rooms, bson.E{Key: "$lte", Value: *s.MaxRooms})
	}

	if len(rooms) != 0 {
		filter = append(filter, bson.E{Key: "rooms", Value: rooms})
	}

	area := bson.D{}
	if s.MinArea != nil {
		area = append(area, bson.E{Key: "$gte", Value: *s.MinArea})
	}

	if s.MaxArea != nil {
		area = append(area, bson.E{Key: "$lte", Value: *s.MaxArea})
	}

	if len(area) != 0 {
		filter = append(filter, bson.E{Key: "area", Value: area})
	}

	if s.IsOwner != nil {
		filter = append(filter, bson.E{Key: "is_owner", Value: *s.IsOwner})
	}

	if s.Coordinates != nil && s.MaxDistance != nil {
		filter = append(
			filter,
			bson.D{
				{Key: "location", Value: bson.D{
					{Key: "$near", Value: bson.D{
						{Key: "$geometry", Value: bson.D{
							{Key: "type", Value: "Point"},
							{Key: "coordinates", Value: primitive.A{s.Coordinates.Lng, s.Coordinates.Lat}},
						}},
						{Key: "$maxDistance", Value: s.MaxDistance},
					}},
				}},
			}...,
		)
	}

	date := bson.D{}
	if s.FromTimestamp != nil {
		date = append(date, bson.E{Key: "$gte", Value: *s.FromTimestamp})
	}

	if len(date) != 0 {
		filter = append(filter, bson.E{Key: "date", Value: date})
	}

	return filter
}

func (s *filter) filter() any {
	filter := bson.D{}

	if len(s.ID) != 0 {
		filter = append(filter, bson.E{Key: "_id", Value: s.ID})
	}

	if s.UserID != nil {
		filter = append(filter, bson.E{Key: "user_id", Value: *s.UserID})
	}

	if s.Name != nil {
		filter = append(filter, bson.E{Key: "name", Value: *s.Name})
	}

	return filter
}

func (s *filter) user() any {
	filter := bson.D{}

	filter = append(filter, bson.E{Key: "tg_id", Value: s.UserID})

	return filter
}

func (s *filter) city() any {
	filter := bson.D{}

	if s.CityName != nil {
		filter = append(filter, bson.E{Key: "city_name", Value: *s.CityName})
	}

	return filter
}
