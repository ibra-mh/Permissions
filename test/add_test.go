package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	type testCase struct {
		name        string
		a           int
		b           int
		checkResult func(result int)
	}

	testCases := []testCase{
		{
			name: "success",
			a:    1,
			b:    1,
			checkResult: func(result int) {
				assert.Equal(t, 2, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := add(tc.a, tc.b)
			tc.checkResult(result)
		})

	}
}

