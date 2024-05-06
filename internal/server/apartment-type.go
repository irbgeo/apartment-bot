package server

import "time"

const (
	RentAdType int64 = iota + 1
	SaleAdType
)

const (
	NewBuildingStatus = iota + 1
	UnderConstructionBuildingStatus
	OldBuildingStatus
)

type Apartment struct {
	ID             int64
	AdType         int64
	BuildingStatus int64
	Price          float64
	Rooms          float64
	Bedrooms       int64
	Floor          int64
	Area           float64
	Phone          string
	District       string
	City           string
	Coordinates    *Coordinates
	Comment        string
	OrderDate      time.Time
	URL            string
	PhotoURLs      []string
	IsOwner        bool

	Filter map[int64][]string
}
