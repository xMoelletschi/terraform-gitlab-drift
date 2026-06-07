package gitlab

import (
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestFilterProjects(t *testing.T) {
	marked := &gl.ISOTime{}
	projects := []*gl.Project{
		{ID: 1, PathWithNamespace: "g/keep"},
		{ID: 2, PathWithNamespace: "g/deleting", MarkedForDeletionOn: marked},
		{ID: 3, PathWithNamespace: "g/archived", Archived: true},
	}

	t.Run("default hides pending-deletion, keeps archived", func(t *testing.T) {
		got := filterProjects(projects, FilterConfig{})
		ids := projectIDs(got)
		if len(ids) != 2 || !ids[1] || !ids[3] {
			t.Fatalf("expected projects 1 and 3 (archived shown), got %v", ids)
		}
	})

	t.Run("ShowPendingDeletion keeps the deleting project", func(t *testing.T) {
		got := filterProjects(projects, FilterConfig{ShowPendingDeletion: true})
		if len(got) != 3 {
			t.Fatalf("expected all 3 projects, got %d", len(got))
		}
	})

	t.Run("HideArchived drops the archived project", func(t *testing.T) {
		got := filterProjects(projects, FilterConfig{HideArchived: true})
		ids := projectIDs(got)
		if len(ids) != 1 || !ids[1] {
			t.Fatalf("expected only project 1, got %v", ids)
		}
	})
}

func projectIDs(projects []*gl.Project) map[int64]bool {
	ids := make(map[int64]bool, len(projects))
	for _, p := range projects {
		ids[p.ID] = true
	}
	return ids
}
