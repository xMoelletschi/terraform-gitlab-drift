package gitlab

import (
	"context"
	"net/http"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/v2/testing"
	"go.uber.org/mock/gomock"
)

func TestListProtectedBranches_SkipsInaccessibleProject(t *testing.T) {
	tc := gitlabtesting.NewTestClient(t)
	c := NewClientFromAPI(tc.Client, "mygroup")

	gomock.InOrder(
		// First project is forbidden — must be skipped, not abort the whole scan.
		tc.MockProtectedBranches.EXPECT().
			ListProtectedBranches(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil, &gl.ErrorResponse{StatusCode: http.StatusForbidden}),
		// Second project must still be fetched.
		tc.MockProtectedBranches.EXPECT().
			ListProtectedBranches(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*gl.ProtectedBranch{{Name: "main"}}, &gl.Response{}, nil),
	)

	projects := []*gl.Project{
		{ID: 1, PathWithNamespace: "mygroup/forbidden"},
		{ID: 2, PathWithNamespace: "mygroup/ok"},
	}
	result, err := c.ListProtectedBranches(context.Background(), projects)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result[1]; ok {
		t.Errorf("forbidden project should have no entry, got %v", result[1])
	}
	if len(result[2]) != 1 || result[2][0].Name != "main" {
		t.Errorf("expected project 2 to have branch 'main', got %v", result[2])
	}
}
