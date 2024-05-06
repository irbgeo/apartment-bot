package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s1, err := NewService(nil, nil)
	require.NoError(t, err)

	_, err = NewService(nil, nil)
	require.EqualError(t, err, "client already exist")
	s1.Stop()

	s, err := NewService(nil, nil)
	require.NoError(t, err)
	s.Stop()
}

func TestAvailableCities(t *testing.T) {
	firstCities := []string{"City1", "City2", "City3"}

	// Create a new service instance
	s, err := NewService(nil, firstCities)
	require.NoError(t, err)
	defer s.Stop()

	allCities := []string{"City1", "City2", "City3", "City4", "City5", "City6", "City7", "City8", "City9", "City10"}

	for _, city := range allCities {
		s.cities.Store(city, []string{})
	}

	availableCities := s.AvailableCities()
	require.Equal(t, firstCities, availableCities[:len(firstCities)])
	for _, city := range allCities {
		require.Contains(t, availableCities, city)
	}
}
