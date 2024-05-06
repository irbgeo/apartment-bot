package mongo

type filter struct {
	ID             string              `bson:"_id"`
	AdType         *int64              `bson:"ad_type"`
	BuildingStatus *int64              `bson:"building_status"`
	ApartmentID    *int64              `bson:"-"`
	Name           *string             `bson:"name"`
	UserID         *int64              `bson:"user_id"`
	District       map[string]struct{} `bson:"district"`
	CityName       *string             `bson:"city"`
	MinPrice       *float64            `bson:"min_price"`
	MaxPrice       *float64            `bson:"max_price"`
	MinRooms       *float64            `bson:"min_rooms"`
	MaxRooms       *float64            `bson:"max_rooms"`
	MinArea        *float64            `bson:"min_area"`
	MaxArea        *float64            `bson:"max_area"`
	IsOwner        *bool               `bson:"is_owner"`
	Coordinates    *coordinates        `bson:"location_coordinates"`
	MaxDistance    *float64            `bson:"max_distance"`
	PauseTimestamp *int64              `bson:"pause_timestamp"`
	FromTimestamp  *int64              `bson:"-"`
}

type coordinates struct {
	Lat float64 `bson:"lat"`
	Lng float64 `bson:"lng"`
}
