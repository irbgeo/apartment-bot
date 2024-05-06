package server

import (
	"strings"
)

type Filter struct {
	IsUpdate bool

	ID   string
	User *User

	AdType         *int64
	BuildingStatus *int64
	Name           *string
	District       map[string]struct{}
	City           *string
	MinPrice       *float64
	MaxPrice       *float64
	MinRooms       *float64
	MaxRooms       *float64
	MinArea        *float64
	MaxArea        *float64
	IsOwner        *bool
	Coordinates    *Coordinates
	MaxDistance    *float64

	TillTimestamp  *int64
	FromTimestamp  *int64
	PauseTimestamp *int64

	ApartmentID *int64
}

type Coordinates struct {
	Lat float64
	Lng float64
}

func (s *Filter) CheckDistance(a *Apartment) bool {
	if s.MaxDistance == nil || s.Coordinates == nil {
		return true
	}

	if a.Coordinates == nil {
		return false
	}

	dist := distance(
		a.Coordinates.Lat, a.Coordinates.Lng,
		s.Coordinates.Lat, s.Coordinates.Lng,
	)
	return dist <= *s.MaxDistance+accuracy
}

func (s *Filter) CheckDistrict(a *Apartment) bool {
	if len(s.District) == 0 {
		return true
	}

	for district := range s.District {
		if strings.Contains(strings.ToLower(a.District), strings.ToLower(district)) {
			return true
		}
	}
	return false
}

func (s *Filter) IsFit(a *Apartment) bool {
	if s.PauseTimestamp != nil {
		return false
	}

	isFit := true

	isFit = isFit && s.CheckDistrict(a)
	isFit = isFit && s.CheckDistance(a)

	if s.AdType != nil {
		isFit = isFit && *s.AdType == a.AdType
	}

	if s.BuildingStatus != nil {
		isFit = isFit && *s.BuildingStatus == a.BuildingStatus
	}

	if a.City != "" && s.City != nil {
		isFit = isFit && strings.Contains(a.City, *s.City)
	}

	if s.MinPrice != nil {
		isFit = isFit && *s.MinPrice <= a.Price
	}
	if s.MaxPrice != nil {
		isFit = isFit && *s.MaxPrice >= a.Price
	}

	if s.MinRooms != nil {
		isFit = isFit && *s.MinRooms <= a.Rooms
	}
	if s.MaxRooms != nil {
		isFit = isFit && *s.MaxRooms >= a.Rooms
	}

	if s.MinArea != nil {
		isFit = isFit && *s.MinArea <= a.Area
	}
	if s.MaxArea != nil {
		isFit = isFit && *s.MaxArea >= a.Area
	}

	if s.IsOwner != nil {
		isFit = isFit && *s.IsOwner == a.IsOwner
	}

	return isFit
}
