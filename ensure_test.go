package godirwalk

import (
	"fmt"
	"sort"
	"testing"
)

func ensureStringSlicesMatch(tb testing.TB, actual, expected []string) {
	tb.Helper()

	results := make(map[string]int)

	for _, s := range actual {
		results[s] -= 1
	}
	for _, s := range expected {
		results[s] += 1
	}

	keys := make([]string, 0, len(results))
	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, s := range keys {
		v, ok := results[s]
		if !ok {
			panic(fmt.Errorf("cannot find key: %s", s))
		}
		switch v {
		case -1:
			tb.Errorf("GOT: %q (extra)", s)
		case 0:
			// t.Errorf("actual extra key: %s", s)
		case 1:
			tb.Errorf("WANT: %q (missing)", s)
		default:
			tb.Errorf("key has invalid value: %s: %d", s, v)
		}
	}
}
