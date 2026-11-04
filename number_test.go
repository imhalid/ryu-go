package ryugo

import "testing"

func TestNumberText(t *testing.T) {
	for _, tc := range []struct {
		in       string
		at       int
		neg      bool
		decimals int
		want     string
		caret    int
	}{
		{"12a", 3, false, 0, "12", 2},
		{"1a2", 2, false, 0, "12", 1},
		{"1.5", 3, false, 0, "15", 2},    // no point without decimals
		{"1,25", 4, false, 2, "1.25", 4}, // a comma is a point
		{"1.256", 5, false, 2, "1.25", 4},
		{"1.2.3", 5, false, 2, "1.23", 4},
		{"-3", 2, true, 0, "-3", 2},
		{"-3", 2, false, 0, "3", 1},
		{"3-", 2, true, 0, "3", 1},
		{"−7", 2, true, 0, "-7", 2}, // a typographic minus
		{" 4 2 ", 5, false, 0, "42", 2},
	} {
		got, caret := numberText(tc.in, tc.at, tc.neg, tc.decimals)
		if got != tc.want || caret != tc.caret {
			t.Errorf("numberText(%q, %d, %v, %d) = %q, %d; want %q, %d", tc.in, tc.at, tc.neg, tc.decimals, got, caret, tc.want, tc.caret)
		}
	}
}
