package gitlab

import (
	"context"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/v2/testing"
	"go.uber.org/mock/gomock"
)

func TestListProjectVariables_FiltersMaskedHiddenFile(t *testing.T) {
	tc := gitlabtesting.NewTestClient(t)
	c := NewClientFromAPI(tc.Client, "mygroup")

	tc.MockProjectVariables.EXPECT().
		ListVariables(gomock.Any(), gomock.Any(), gomock.Any()).
		Return([]*gl.ProjectVariable{
			{Key: "NORMAL"},
			{Key: "MASKED", Masked: true},
			{Key: "HIDDEN", Hidden: true},
			{Key: "FILE", VariableType: gl.FileVariableType},
		}, &gl.Response{}, nil)

	projects := []*gl.Project{{ID: 1, PathWithNamespace: "mygroup/proj"}}
	result, err := c.ListProjectVariables(context.Background(), projects)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result[1]) != 1 {
		t.Fatalf("expected only the non-masked/hidden/file variable, got %d: %v", len(result[1]), result[1])
	}
	if result[1][0].Key != "NORMAL" {
		t.Errorf("kept variable = %q, want NORMAL", result[1][0].Key)
	}
}
