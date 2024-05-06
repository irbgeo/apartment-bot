package client

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/irbgeo/apartment-bot/internal/server"
)

func TestCheckFilter(t *testing.T) {
	testCases := []struct {
		testCaseName  string
		filter        *server.Filter
		expectedError error
	}{
		{
			testCaseName:  "no set name",
			filter:        &server.Filter{},
			expectedError: ErrUnknownFilterName,
		},
		// TODO: остальные кейсы
	}

	for _, tc := range testCases {
		err := checkFilter(tc.filter)
		require.Equal(t, tc.expectedError, err)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
