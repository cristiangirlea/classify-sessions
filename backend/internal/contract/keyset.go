package contract

import "sort"

// KeySet returns every object key path in a decoded JSON value, with array elements
// folded to "[]". Exported so the live-summary drift test in PR1 can reuse it.
func KeySet(v any) []string {
	set := map[string]bool{}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, child := range t {
				p := prefix + "." + k
				set[p] = true
				walk(p, child)
			}
		case []any:
			for _, child := range t {
				walk(prefix+"[]", child)
			}
		}
	}
	walk("", v)
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
