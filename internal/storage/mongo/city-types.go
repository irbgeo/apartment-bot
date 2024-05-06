package mongo

type city struct {
	CityName string              `bson:"city_name"`
	District map[string]struct{} `bson:"district"`
}
