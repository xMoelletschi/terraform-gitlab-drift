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
	"schedules",
	"branch_protection",
	"service_accounts",
	"job_token_scopes",
}

// Groups map a single name to multiple resource types.
var Groups = map[string][]string{
	"premium": {"hooks", "approval_rules", "mr_approvals", "service_accounts"},
}

// DefaultSkipped lists resource types that are skipped unless the user
// explicitly opts in via --include. Currently contains branch_protection
// because the gitlabhq/gitlab provider's Read does not populate
// unprotect_access_level into state on Free tier, causing perpetual
// ForceNew drift on every plan.
var DefaultSkipped = []string{"branch_protection"}

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
			set[name] = true
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

// Resolve combines --skip values with default-skipped resources, with --include
// allowing opt-in for any of the defaults. Returns the resolved skip set and
// warnings for unknown names from either input.
func Resolve(skipInput, includeInput []string) (Set, []string) {
	set, warnings := Parse(skipInput)
	if set == nil {
		set = make(Set)
	}

	included := make(Set)
	for _, name := range includeInput {
		if !slices.Contains(ResourceTypes, name) {
			warnings = append(warnings, name)
			continue
		}
		included[name] = true
	}

	for _, name := range DefaultSkipped {
		if included[name] {
			continue
		}
		set[name] = true
	}

	if len(set) == 0 {
		return nil, warnings
	}
	return set, warnings
}
