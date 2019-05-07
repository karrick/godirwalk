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
			panic(fmt.Errorf("cannot find key: %s", s)) // panic because this function is broken
		}
		switch v {
		case -1:
			tb.Errorf("GOT: %q (extra)", s)
		case 0:
			// both slices have this key
		case 1:
			tb.Errorf("WANT: %q (missing)", s)
		default:
			panic(fmt.Errorf("key has invalid value: %s: %d", s, v)) // panic because this function is broken
		}
	}
}
