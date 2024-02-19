package mux

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInvalid(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"empty":         {input: "", want: ""},
		"one char":      {input: "x", want: "x"},
		"prefix dash":   {input: "-test", want: "test"},
		"prefix number": {input: "13test", want: "test"},
		"suffix dash":   {input: "test-", want: "test"},
		"space":         {input: "Lukas Zapletal", want: "lukas-zapletal"},
		"fqdn":          {input: "example.com", want: "example.com"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ToHostname(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
