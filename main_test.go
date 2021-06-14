package regexp1

import (
	"testing"
)

func TestRe2post(t *testing.T) {
	tests := []struct {
		re   string
		post string
	}{
		{re: "(ab)|(cd)", post: "ab.cd.|"},
		// {re: "ab|cd", post: "ab.cd.|"},
		{re: "a(b*)", post: "ab*."},
		{re: "ab*", post: "ab*."},
		{re: "a(bb)+a", post: "abb.+.a."},
		{re: "a(b|c)*d", post: "abc|*.d."},
		{re: "((a|b|c)*(d))", post: "abc||*d."},
	}

	for _, test := range tests {
		post := re2post(test.re)
		if post != test.post {
			t.Errorf("re2post(%v) = %v, want %v", test.re, post, test.post)
		}
	}
}

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
		start := post2nfa(post)
		for _, input := range test.inputs {
			if !match(start, input) {
				t.Errorf("match(%v) failed, post: %v", input, post)
			}
		}
	}
}
