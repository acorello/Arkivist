package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TryRun(t *testing.T) {
	config := populateConfig()
	linkToCleanPath(config)
}

func TestHasUnorthodoxRune(t *testing.T) {
	cases := []struct {
		filename          string
		invalidSubstrings []invalidSubstring
	}{
		{
			filename: "Grønbaek",
		},
		{
			filename: "9781101152140 • Drive",
			invalidSubstrings: []invalidSubstring{
				{position: 17, value: "\u00a0"},
			},
		},
		{
			filename: "9781101152140 • Drive • by Daniel H. Pink • Riverhead Books (z-lib.org)",
		},
		{
			filename: "Handbook of Graph Grammars — (Z-Library).pdf",
			invalidSubstrings: []invalidSubstring{
				{position: 27, value: "—"}},
		},
		{
			filename: "Mastering Visual Studio Code A Beginner’s Guide.pdf",
		},
	}
	for n, tc := range cases {
		t.Run(fmt.Sprintf("%0.2d:%q", n, tc.filename), func(t *testing.T) {
			actual := invalidSubstrings(tc.filename)
			expected := tc.invalidSubstrings
			assert.Equal(t, expected, actual)
		})
	}

}

func TestAppendNil(t *testing.T) {
	errs := []error{fmt.Errorf("one")}
	for _, err := range []error{nil, nil} {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 1 {
		t.Errorf("Append added something: %#v", errs)
	}
}
