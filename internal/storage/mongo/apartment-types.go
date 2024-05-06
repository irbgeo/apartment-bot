package mongo

import "time"

type apartment struct {
	ID             int64     `bson:"_id"`
	AdType         int64     `bson:"ad_type"`
	BuildingStatus int64     `bson:"building_status"`
	Price          float64   `bson:"price"`
	Rooms          float64   `bson:"rooms"`
	Bedrooms       int64     `bson:"bedrooms"`
	Area           float64   `bson:"area"`
	Floor          int64     `bson:"floor"`
	Phone          string    `bson:"phone"`
	District       string    `bson:"district"`
	City           string    `bson:"city"`
	Coordinates    *location `bson:"location"`
	Comment        string    `bson:"comment"`
	IsOwner        bool      `bson:"is_owner"`
	OrderDate      time.Time `bson:"order_date"`

	URL       string   `bson:"url"`
	PhotoURLs []string `bson:"photo_urls"`

	Date int64 `bson:"date"`
}

type location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
