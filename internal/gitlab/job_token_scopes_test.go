package gitlab

import (
	"context"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/v2/testing"
	"go.uber.org/mock/gomock"
)

// The provider strips a single self entry on Read; we mirror that so the emitted
// HCL matches state (see the long comment in job_token_scopes.go).
func TestListJobTokenScopes_SelfDedup(t *testing.T) {
	t.Run("self appearing once is stripped", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockJobTokenScope.EXPECT().
			GetProjectJobTokenAccessSettings(gomock.Any(), gomock.Any()).
			Return(&gl.JobTokenAccessSettings{InboundEnabled: true}, &gl.Response{}, nil)
		tc.MockJobTokenScope.EXPECT().
			GetProjectJobTokenInboundAllowList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*gl.Project{{ID: 1}, {ID: 2}}, &gl.Response{}, nil)
		tc.MockJobTokenScope.EXPECT().
			GetJobTokenAllowlistGroups(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, &gl.Response{}, nil)

		result, err := c.ListJobTokenScopes(context.Background(), []*gl.Project{{ID: 1, PathWithNamespace: "mygroup/proj"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		scope := result[1]
		if scope == nil {
			t.Fatal("expected a scope for project 1")
		}
		if len(scope.AllowedProjects) != 1 || scope.AllowedProjects[0].ID != 2 {
			t.Errorf("self appearing once should be stripped, leaving only project 2; got %v", scope.AllowedProjects)
		}
	})

	t.Run("self appearing twice is kept once", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockJobTokenScope.EXPECT().
			GetProjectJobTokenAccessSettings(gomock.Any(), gomock.Any()).
			Return(&gl.JobTokenAccessSettings{InboundEnabled: true}, &gl.Response{}, nil)
		tc.MockJobTokenScope.EXPECT().
			GetProjectJobTokenInboundAllowList(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*gl.Project{{ID: 1}, {ID: 1}, {ID: 2}}, &gl.Response{}, nil)
		tc.MockJobTokenScope.EXPECT().
			GetJobTokenAllowlistGroups(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, &gl.Response{}, nil)

		result, err := c.ListJobTokenScopes(context.Background(), []*gl.Project{{ID: 1, PathWithNamespace: "mygroup/proj"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		scope := result[1]
		if scope == nil {
			t.Fatal("expected a scope for project 1")
		}
		var hasSelf, hasTwo bool
		for _, ap := range scope.AllowedProjects {
			switch ap.ID {
			case 1:
				hasSelf = true
			case 2:
				hasTwo = true
			}
		}
		if len(scope.AllowedProjects) != 2 || !hasSelf || !hasTwo {
			t.Errorf("self appearing twice should be kept once alongside project 2; got %v", scope.AllowedProjects)
		}
	})
}
