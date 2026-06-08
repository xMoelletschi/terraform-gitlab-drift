package gitlab

import (
	"context"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/v2/testing"
	"go.uber.org/mock/gomock"
)

// FetchAll fetches groups+projects and then the enabled subresource fetchers
// concurrently. This exercises that the results compose correctly and, under
// `go test -race`, that the concurrent writes are race-free.
func TestFetchAll_ComposesResultsConcurrently(t *testing.T) {
	tc := gitlabtesting.NewTestClient(t)
	c := NewClientFromAPI(tc.Client, "mygroup")

	mainGroup := &gl.Group{ID: 10, Path: "mygroup", FullPath: "mygroup"}
	project := &gl.Project{ID: 1, Path: "proj", PathWithNamespace: "mygroup/proj"}

	// Groups + projects (the pre-phase, run concurrently with each other).
	tc.MockGroups.EXPECT().GetGroup(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mainGroup, &gl.Response{}, nil)
	tc.MockGroups.EXPECT().ListDescendantGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, &gl.Response{}, nil)
	tc.MockGroups.EXPECT().ListGroupProjects(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.Project{project}, &gl.Response{}, nil)

	// Three subresource fetchers enabled → three concurrent goroutines.
	tc.MockGroups.EXPECT().ListGroupMembers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.GroupMember{{ID: 100, Username: "alice"}}, &gl.Response{}, nil)
	tc.MockGroupLabels.EXPECT().ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.GroupLabel{{ID: 1, Name: "bug"}}, &gl.Response{}, nil)
	tc.MockLabels.EXPECT().ListLabels(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.Label{{ID: 2, Name: "feature", IsProjectLabel: true}}, &gl.Response{}, nil)

	skipSet := skip.Set{
		"schedules": true, "variables": true, "branch_protection": true,
		"tag_protection": true, "job_token_scopes": true, "hooks": true,
	}

	res, err := c.FetchAll(context.Background(), skipSet, FilterConfig{})
	if err != nil {
		t.Fatalf("FetchAll error: %v", err)
	}
	if len(res.Groups) != 1 || len(res.Projects) != 1 {
		t.Fatalf("groups=%d projects=%d, want 1 and 1", len(res.Groups), len(res.Projects))
	}
	if len(res.GroupMembers[10]) != 1 {
		t.Errorf("group members not composed: %v", res.GroupMembers)
	}
	if len(res.GroupLabels[10]) != 1 {
		t.Errorf("group labels not composed: %v", res.GroupLabels)
	}
	if len(res.ProjectLabels[1]) != 1 {
		t.Errorf("project labels not composed: %v", res.ProjectLabels)
	}
}
