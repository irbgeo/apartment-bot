package ssge

import (
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/irbgeo/apartment-bot/internal/server"
)

var (
	individualUserEntityType = "individual"
	apartmentURLPrefix       = "https://home.ss.ge/en/real-estate/"
)

var dealTypeMap = map[int64]int64{
	rentRealEstateDealType: server.RentAdType,
	saleRealEstateDealType: server.SaleAdType,
}

var realEstateStatusIDMap = map[int64]int64{
	2:   server.NewBuildingStatus,
	3:   server.UnderConstructionBuildingStatus,
	453: server.OldBuildingStatus,
}

func toServerApartment(in apartment) server.Apartment {
	out := server.Apartment{
		ID:             in.ApplicationID,
		AdType:         dealTypeMap[in.RealEstateDealTypeID],
		BuildingStatus: realEstateStatusIDMap[in.RealEstateStatusID],
		Price:          in.Price.PriceUSD,
		Bedrooms:       in.Bedrooms,
		District:       prepareTitle(in.Address.SubdistrictTitle),
		City:           prepareTitle(in.Address.CityTitle),
		Comment:        in.Description.En,
		IsOwner:        strings.ToLower(in.UserEntityType) == individualUserEntityType,
		OrderDate:      in.OrderDate,
	}

	out.Rooms, _ = strconv.ParseFloat(in.Rooms, 64)
	out.Area, _ = strconv.ParseFloat(in.TotalArea, 64)
	out.Floor, _ = strconv.ParseInt(in.Floor, 10, 64)

	if len(in.ApplicationPhones) > 0 {
		out.Phone = in.ApplicationPhones[0].PhoneNumber
	}

	out.PhotoURLs = make([]string, 0, len(in.AppImages))

	for _, p := range in.AppImages {
		out.PhotoURLs = append(out.PhotoURLs, p.FileName)
	}

	if in.LocationLatitude != 0 && in.LocationLongitude != 0 {
		out.Coordinates = &server.Coordinates{
			Lat: in.LocationLatitude,
			Lng: in.LocationLongitude,
		}
	}

	out.URL = apartmentURLPrefix + strconv.FormatInt(in.ApplicationID, 10)

	return out
}

func prepareTitle(title string) string {
	return cases.Title(language.Und).String(strings.ToLower(title))
}
