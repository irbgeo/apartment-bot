package apartment

import "time"

// Config contains configuration for creating a new service
type Config struct {
	MaxFetchPages int64
	ApartmentTTL  time.Duration
}
