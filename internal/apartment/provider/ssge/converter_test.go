package ssge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareTitle(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "hello world",
			expected: "Hello World",
		},
		{
			input:    "this is a test",
			expected: "This Is A Test",
		},
		{
			input:    "another example",
			expected: "Another Example",
		},
	}

	for _, tc := range testCases {
		actual := prepareTitle(tc.input)
		assert.Equal(t, tc.expected, actual)
	}
}
