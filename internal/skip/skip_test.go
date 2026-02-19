package skip

import (
	"slices"
	"testing"
)

func TestSetHas(t *testing.T) {
	s := Set{"memberships": true, "hooks": true}
	if !s.Has("memberships") {
		t.Error("expected Has(memberships) = true")
	}
	if s.Has("labels") {
		t.Error("expected Has(labels) = false")
	}
}

func TestNilSetHas(t *testing.T) {
	var s Set
	if s.Has("memberships") {
		t.Error("nil set should return false")
	}
}

func TestParseEmpty(t *testing.T) {
	set, warnings := Parse(nil)
	if set != nil {
		t.Errorf("expected nil set, got %v", set)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestParseResourceTypes(t *testing.T) {
	set, warnings := Parse([]string{"memberships", "hooks"})
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if !set.Has("memberships") {
		t.Error("expected memberships in set")
	}
	if !set.Has("hooks") {
		t.Error("expected hooks in set")
	}
	if set.Has("labels") {
		t.Error("expected labels NOT in set")
	}
}

func TestParseGroup(t *testing.T) {
	set, warnings := Parse([]string{"premium"})
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	for _, rt := range Groups["premium"] {
		if !set.Has(rt) {
			t.Errorf("expected %s in set from premium group", rt)
		}
	}
	if set.Has("memberships") {
		t.Error("memberships should not be in premium group")
	}
}

func TestParseGroupAndResourceType(t *testing.T) {
	set, warnings := Parse([]string{"premium", "memberships"})
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if !set.Has("memberships") {
		t.Error("expected memberships in set")
	}
	if !set.Has("hooks") {
		t.Error("expected hooks in set from premium group")
	}
}

func TestParseUnknown(t *testing.T) {
	set, warnings := Parse([]string{"foobar"})
	if set != nil {
		t.Errorf("expected nil set with only unknowns, got %v", set)
	}
	if !slices.Contains(warnings, "foobar") {
		t.Errorf("expected foobar in warnings, got %v", warnings)
	}
}

func TestParseMixed(t *testing.T) {
	set, warnings := Parse([]string{"memberships", "nonexistent", "premium"})
	if !slices.Contains(warnings, "nonexistent") {
		t.Errorf("expected nonexistent in warnings, got %v", warnings)
	}
	if !set.Has("memberships") {
		t.Error("expected memberships in set")
	}
	if !set.Has("hooks") {
		t.Error("expected hooks from premium group")
	}
}
