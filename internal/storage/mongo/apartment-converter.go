package mongo

import (
	"github.com/irbgeo/apartment-bot/internal/server"
)

func toMongoApartment(in server.Apartment) apartment {
	out := apartment{
		ID:             in.ID,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Price:          in.Price,
		Rooms:          in.Rooms,
		Bedrooms:       in.Bedrooms,
		Area:           in.Area,
		Floor:          in.Floor,
		Phone:          in.Phone,
		City:           in.City,
		District:       in.District,
		Comment:        in.Comment,
		IsOwner:        in.IsOwner,
		OrderDate:      in.OrderDate,

		URL:       in.URL,
		PhotoURLs: in.PhotoURLs,
	}

	if in.Coordinates != nil {
		out.Coordinates = &location{
			Type:        "Point",
			Coordinates: []float64{in.Coordinates.Lng, in.Coordinates.Lat},
		}
	}

	return out
}

func toApartment(in apartment) server.Apartment {
	out := server.Apartment{
		ID:             in.ID,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Price:          in.Price,
		Rooms:          in.Rooms,
		Bedrooms:       in.Bedrooms,
		Area:           in.Area,
		Floor:          in.Floor,
		Phone:          in.Phone,
		District:       in.District,
		City:           in.City,
		Comment:        in.Comment,
		IsOwner:        in.IsOwner,
		OrderDate:      in.OrderDate,

		PhotoURLs: in.PhotoURLs,
		URL:       in.URL,
	}

	if in.Coordinates != nil {
		out.Coordinates = &server.Coordinates{
			Lat: in.Coordinates.Coordinates[1],
			Lng: in.Coordinates.Coordinates[0],
		}
	}

	return out
}
