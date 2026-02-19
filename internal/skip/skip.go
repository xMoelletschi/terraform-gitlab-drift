package skip

import "slices"

// Set tracks which resource types should be skipped.
type Set map[string]bool

// Has returns true if the given resource type is in the skip set.
func (s Set) Has(name string) bool { return s[name] }

// Known resource types that can be skipped.
var ResourceTypes = []string{
	"memberships",
	"hooks",
	"labels",
	"variables",
	"approval_rules",
	"mr_approvals",
	"pipeline_schedules",
	"branch_protection",
	"service_accounts",
}

// Groups map a single name to multiple resource types.
var Groups = map[string][]string{
	"premium": {"hooks", "approval_rules", "mr_approvals", "service_accounts"},
}

// Parse resolves group names, validates resource type names, and returns
// the resulting Set plus any unknown names as warnings.
func Parse(input []string) (Set, []string) {
	if len(input) == 0 {
		return nil, nil
	}

	set := make(Set)
	var warnings []string

	for _, name := range input {
		if members, ok := Groups[name]; ok {
			for _, m := range members {
				set[m] = true
			}
			continue
		}
		if slices.Contains(ResourceTypes, name) {
			set[name] = true
			continue
		}
		warnings = append(warnings, name)
	}

	if len(set) == 0 {
		return nil, warnings
	}
	return set, warnings
}
