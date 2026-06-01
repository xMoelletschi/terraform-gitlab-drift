package terraform

import (
	"slices"
	"testing"
)

func TestDedupeResourceNames_NoCollisionsUnchanged(t *testing.T) {
	// The hard constraint: names that don't collide must be returned byte-for-byte
	// unchanged, so existing terraform addresses never shift.
	in := []string{"alpha", "beta", "gamma"}
	got := dedupeResourceNames(in)
	want := []string{"alpha", "beta", "gamma"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDedupeResourceNames_CollisionsSuffixed(t *testing.T) {
	in := []string{"v1_0", "v1_0", "v1_0"}
	got := dedupeResourceNames(in)
	want := []string{"v1_0", "v1_0_1", "v1_0_2"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDedupeResourceNames_MixedCollisions(t *testing.T) {
	in := []string{"a", "b", "a", "a", "b"}
	got := dedupeResourceNames(in)
	want := []string{"a", "b", "a_1", "a_2", "b_1"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
