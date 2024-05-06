package server

import (
	"time"

	api "github.com/irbgeo/apartment-bot/internal/api/server/proto"
	"github.com/irbgeo/apartment-bot/internal/server"
)

func filterToAPI(in server.Filter) *api.Filter {
	out := &api.Filter{
		Id:             in.ID,
		UserId:         in.User.ID,
		Name:           in.Name,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Districts:      make([]string, 0, len(in.District)),
		City:           in.City,
		MinPrice:       in.MinPrice,
		MaxPrice:       in.MaxPrice,
		MinRooms:       in.MinRooms,
		MaxRooms:       in.MaxRooms,
		MinArea:        in.MinArea,
		MaxArea:        in.MaxArea,
		MaxDistance:    in.MaxDistance,
		IsOwner:        in.IsOwner,

		PauseTimestamp: in.PauseTimestamp,
	}

	if in.Coordinates != nil {
		out.LocationCoordinates = &api.Coordinates{
			Lat: in.Coordinates.Lat,
			Lng: in.Coordinates.Lng,
		}
	}

	for d := range in.District {
		out.Districts = append(out.Districts, d)
	}
	return out
}

func filterFromAPI(in *api.Filter) server.Filter {
	out := server.Filter{
		ID:             in.Id,
		User:           &server.User{ID: in.UserId},
		Name:           in.Name,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		District:       make(map[string]struct{}),
		City:           in.City,
		MinPrice:       in.MinPrice,
		MaxPrice:       in.MaxPrice,
		MinRooms:       in.MinRooms,
		MaxRooms:       in.MaxRooms,
		MinArea:        in.MinArea,
		MaxArea:        in.MaxArea,
		MaxDistance:    in.MaxDistance,
		IsOwner:        in.IsOwner,

		PauseTimestamp: in.PauseTimestamp,
	}

	if in.LocationCoordinates != nil {
		out.Coordinates = &server.Coordinates{
			Lat: in.LocationCoordinates.Lat,
			Lng: in.LocationCoordinates.Lng,
		}
	}

	for _, d := range in.Districts {
		out.District[d] = struct{}{}
	}

	return out
}

func apartmentToAPI(in server.Apartment) *api.Apartment {
	out := &api.Apartment{
		Id:             in.ID,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Price:          in.Price,
		Rooms:          in.Rooms,
		Bedrooms:       (in.Bedrooms),
		Floor:          (in.Floor),
		Area:           in.Area,
		Phone:          in.Phone,
		District:       in.District,
		City:           in.City,
		Comment:        in.Comment,
		OrderDate:      in.OrderDate.Format("2006-01-02"),
		Url:            in.URL,
		PhotoUrls:      in.PhotoURLs,
		IsOwner:        in.IsOwner,

		Filters: make([]*api.ApartmentFilter, 0, len(in.Filter)),
	}

	if in.Coordinates != nil {
		out.Coordinates = &api.Coordinates{
			Lat: in.Coordinates.Lat,
			Lng: in.Coordinates.Lng,
		}
	}

	for uid, names := range in.Filter {
		out.Filters = append(out.Filters, &api.ApartmentFilter{
			UserId:      uid,
			FilterNames: names,
		})
	}

	return out
}

func apartmentFromAPI(in *api.Apartment) server.Apartment {
	out := server.Apartment{
		ID:             in.Id,
		AdType:         in.AdType,
		BuildingStatus: in.BuildingStatus,
		Price:          in.Price,
		Rooms:          in.Rooms,
		Bedrooms:       in.Bedrooms,
		Floor:          in.Floor,
		Area:           in.Area,
		Phone:          in.Phone,
		District:       in.District,
		City:           in.City,
		Comment:        in.Comment,
		URL:            in.Url,
		PhotoURLs:      in.PhotoUrls,
		IsOwner:        in.IsOwner,

		Filter: make(map[int64][]string),
	}

	if in.Coordinates != nil {
		out.Coordinates = &server.Coordinates{
			Lat: in.Coordinates.Lat,
			Lng: in.Coordinates.Lng,
		}
	}

	out.OrderDate, _ = time.Parse("2006-01-02", in.OrderDate)

	for _, f := range in.Filters {
		out.Filter[f.UserId] = f.FilterNames
	}
	return out
}
