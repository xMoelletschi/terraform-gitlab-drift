package gitlab

import (
	"net/http"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestPaginate_CollectsAllPages(t *testing.T) {
	opts := &gl.ListOptions{Page: 1, PerPage: 100}
	calls := 0
	fetch := func() ([]int, *gl.Response, error) {
		calls++
		if calls == 1 {
			return []int{1, 2}, &gl.Response{NextPage: 2}, nil
		}
		return []int{3}, &gl.Response{NextPage: 0}, nil
	}

	got, err := paginate(opts, "test", "path", fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", got)
	}
	if opts.Page != 2 {
		t.Errorf("opts.Page = %d, want 2 (advanced to last fetched page)", opts.Page)
	}
}

func TestPaginate_SkipsOnForbidden(t *testing.T) {
	opts := &gl.ListOptions{Page: 1, PerPage: 100}
	fetch := func() ([]int, *gl.Response, error) {
		return nil, nil, &gl.ErrorResponse{StatusCode: http.StatusForbidden}
	}

	got, err := paginate(opts, "test", "path", fetch)
	if err != nil {
		t.Fatalf("forbidden should be skipped, not returned as error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestPaginate_PropagatesOtherErrors(t *testing.T) {
	opts := &gl.ListOptions{Page: 1, PerPage: 100}
	fetch := func() ([]int, *gl.Response, error) {
		return nil, nil, &gl.ErrorResponse{StatusCode: http.StatusInternalServerError}
	}

	if _, err := paginate(opts, "test", "path", fetch); err == nil {
		t.Error("expected a non-skippable error to propagate")
	}
}
