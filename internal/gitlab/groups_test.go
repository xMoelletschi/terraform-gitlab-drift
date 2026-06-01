package gitlab

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestFilterDeletedGroups(t *testing.T) {
	marked := &gl.ISOTime{}
	groups := []*gl.Group{
		{ID: 1, FullPath: "keep"},
		{ID: 2, FullPath: "deleting", MarkedForDeletionOn: marked},
		{ID: 3, FullPath: "keep2"},
	}

	got := filterDeletedGroups(groups)

	if len(got) != 2 {
		t.Fatalf("expected 2 groups after filtering, got %d", len(got))
	}
	for _, g := range got {
		if g.MarkedForDeletionOn != nil {
			t.Errorf("group %q is marked for deletion and should have been filtered", g.FullPath)
		}
	}
}
