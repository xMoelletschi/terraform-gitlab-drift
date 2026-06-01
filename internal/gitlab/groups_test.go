package gitlab

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestFilterGroups(t *testing.T) {
	marked := &gl.ISOTime{}
	groups := []*gl.Group{
		{ID: 1, FullPath: "keep"},
		{ID: 2, FullPath: "deleting", MarkedForDeletionOn: marked},
		{ID: 3, FullPath: "keep2"},
	}

	t.Run("default hides pending-deletion", func(t *testing.T) {
		got := filterGroups(groups, FilterConfig{})
		if len(got) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(got))
		}
		for _, g := range got {
			if g.MarkedForDeletionOn != nil {
				t.Errorf("group %q pending deletion should be hidden", g.FullPath)
			}
		}
	})

	t.Run("ShowPendingDeletion keeps them", func(t *testing.T) {
		got := filterGroups(groups, FilterConfig{ShowPendingDeletion: true})
		if len(got) != 3 {
			t.Fatalf("expected all 3 groups, got %d", len(got))
		}
	})
}
