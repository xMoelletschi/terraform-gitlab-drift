package gitlab

import (
	"context"
	"errors"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestFetchPerProject(t *testing.T) {
	projects := []*gl.Project{{ID: 1}, nil, {ID: 2}, {ID: 3}}

	got, err := fetchPerProject(context.Background(), 2, projects,
		func(_ context.Context, p *gl.Project) (int64, bool, error) {
			if p.ID == 3 {
				return 0, false, nil // drop: keep == false
			}
			return p.ID * 10, true, nil
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[1] != 10 || got[2] != 20 {
		t.Fatalf("got %v, want {1:10, 2:20}", got)
	}
	if _, ok := got[3]; ok {
		t.Error("project 3 returned keep=false and must be dropped")
	}
	// nil project must be skipped without panicking.
}

func TestFetchPerProject_PropagatesError(t *testing.T) {
	projects := []*gl.Project{{ID: 1}, {ID: 2}}

	_, err := fetchPerProject(context.Background(), 2, projects,
		func(_ context.Context, p *gl.Project) (int64, bool, error) {
			if p.ID == 2 {
				return 0, false, errors.New("boom")
			}
			return 10, true, nil
		})
	if err == nil {
		t.Error("expected the per-project fetch error to propagate")
	}
}
