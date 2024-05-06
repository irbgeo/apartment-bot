package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDistance(t *testing.T) {
	testCases := []struct {
		testCaseName string
		lat1         float64
		lon1         float64
		lat2         float64
		lon2         float64
		expected     float64
	}{
		{
			testCaseName: "distance between different points",
			lat1:         37.7749,
			lon1:         -122.4194,
			lat2:         34.0522,
			lon2:         -118.2437,
			expected:     559100,
		},
	}

	for _, tc := range testCases {
		actual := distance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
		require.InEpsilon(t, tc.expected, actual, accuracy)
	}
}
