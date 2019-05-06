package godirwalk

import "testing"

func ensureStringSlicesMatch(tb testing.TB, actual, expected []string) {
	tb.Helper()
	if got, want := len(actual), len(expected); got != want {
		tb.Errorf("GOT: %v; WANT: %v", got, want)
	}
	la := len(actual)
	le := len(expected)
	for i := 0; i < la || i < le; i++ {
		if i < la {
			if i < le {
				if got, want := actual[i], expected[i]; got != want {
					tb.Errorf("GOT: %q; WANT: %q", got, want)
				}
			} else {
				tb.Errorf("GOT: %q (extra)", actual[i])
			}
		} else if i < le {
			tb.Errorf("WANT: %q (missing)", expected[i])
		}
	}
}
