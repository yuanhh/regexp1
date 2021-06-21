package regexp2

import (
	"testing"
)

func TestRegexp(t *testing.T) {
	tests := []struct {
		re     string
		inputs []string
	}{
		{
			re: "(ab)|(cd)",
			inputs: []string{
				"ab",
				"cd",
			},
		},
		{
			re: "a(b*)",
			inputs: []string{
				"a",
				"abbb",
			},
		},
		{
			re: "a(bb)+a",
			inputs: []string{
				"abbbba",
			},
		},
		{
			re: "a(b|c)*d",
			inputs: []string{
				"abbbbd",
				"accd",
			},
		},
		{
			re: "((a|b|c)*(d))",
			inputs: []string{
				"aaad",
				"acd",
				"d",
				"cd",
				"bbbbbbd",
			},
		},
	}

	for _, test := range tests {
		post := re2post(test.re)
		prog := post2nfa(post)
		for _, input := range test.inputs {
			if !execute(prog, input) {
				t.Errorf("match(%v) failed, post: %v", input, post)
			}
		}
	}
}
