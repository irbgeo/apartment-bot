package server

import "math"

const (
	earthRadius = 6371000 // Earth's radius in meters
	accuracy    = 100
)

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	radlat1 := toRadians(lat1)
	radlat2 := toRadians(lat2)
	radlon1 := toRadians(lon1)
	radlon2 := toRadians(lon2)

	// Haversine formula
	distance := earthRadius * math.Sqrt(math.Pow(radlat2-radlat1, 2)+math.Pow(radlon2-radlon1, 2))

	return distance
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
