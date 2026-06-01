package terraform

import "fmt"

// dedupeResourceNames returns collision-free terraform resource names for a list
// of sibling resources, preserving order. Distinct source values can normalize to
// the same label (e.g. "v1.0" and "v1-0" both become "v1_0"); the second and later
// entries that share a base get a numeric suffix (_1, _2, ...) so every resource
// address stays unique. Names that do not collide are returned unchanged, so
// existing terraform addresses never shift.
//
// The writer and the import generator must build base names the same way and pass
// the same slice so the names they emit agree.
func dedupeResourceNames(bases []string) []string {
	names := make([]string, len(bases))
	used := make(map[string]bool, len(bases))
	for i, base := range bases {
		name := base
		for n := 1; used[name]; n++ {
			name = fmt.Sprintf("%s_%d", base, n)
		}
		used[name] = true
		names[i] = name
	}
	return names
}
