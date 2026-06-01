package gitlab

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestFilterDeletedProjects(t *testing.T) {
	marked := &gl.ISOTime{}
	projects := []*gl.Project{
		{ID: 1, PathWithNamespace: "g/keep"},
		{ID: 2, PathWithNamespace: "g/deleting", MarkedForDeletionOn: marked},
		{ID: 3, PathWithNamespace: "g/keep2"},
	}

	got := filterDeletedProjects(projects)

	if len(got) != 2 {
		t.Fatalf("expected 2 projects after filtering, got %d", len(got))
	}
	for _, p := range got {
		if p.MarkedForDeletionOn != nil {
			t.Errorf("project %q is marked for deletion and should have been filtered", p.PathWithNamespace)
		}
	}
}
