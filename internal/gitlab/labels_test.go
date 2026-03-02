package gitlab

import (
	"context"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/testing"
	"go.uber.org/mock/gomock"
)

func TestListGroupLabels(t *testing.T) {
	t.Run("returns labels for single group", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockGroupLabels.EXPECT().
			ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*gl.GroupLabel{
				{ID: 1, Name: "bug", Color: "#dc143c"},
				{ID: 2, Name: "feature", Color: "#009966"},
			}, &gl.Response{}, nil)

		groups := []*gl.Group{{ID: 10, Path: "mygroup", FullPath: "mygroup"}}
		result, err := c.ListGroupLabels(context.Background(), groups)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result[10]) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(result[10]))
		}
		if result[10][0].Name != "bug" {
			t.Errorf("label name = %q, want %q", result[10][0].Name, "bug")
		}
	})

	t.Run("deduplicates inherited labels across groups", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		gomock.InOrder(
			// Parent group owns labels 1 and 2.
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 1, Name: "bug"},
					{ID: 2, Name: "feature"},
				}, &gl.Response{}, nil),

			// Child group inherits label 1 and 2, and owns label 3.
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 1, Name: "bug"},
					{ID: 2, Name: "feature"},
					{ID: 3, Name: "docs"},
				}, &gl.Response{}, nil),
		)

		groups := []*gl.Group{
			{ID: 10, Path: "mygroup", FullPath: "mygroup"},
			{ID: 20, Path: "sub", FullPath: "mygroup/sub", ParentID: 10},
		}
		result, err := c.ListGroupLabels(context.Background(), groups)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Parent should have 2 labels.
		if len(result[10]) != 2 {
			t.Fatalf("parent: expected 2 labels, got %d", len(result[10]))
		}

		// Child should only have 1 (the non-inherited label).
		if len(result[20]) != 1 {
			t.Fatalf("child: expected 1 label, got %d", len(result[20]))
		}
		if result[20][0].Name != "docs" {
			t.Errorf("child label = %q, want %q", result[20][0].Name, "docs")
		}
	})

	t.Run("child with only inherited labels gets no entry", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		gomock.InOrder(
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 1, Name: "bug"},
				}, &gl.Response{}, nil),

			// Child only has the inherited label, nothing of its own.
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 1, Name: "bug"},
				}, &gl.Response{}, nil),
		)

		groups := []*gl.Group{
			{ID: 10, Path: "mygroup", FullPath: "mygroup"},
			{ID: 20, Path: "sub", FullPath: "mygroup/sub", ParentID: 10},
		}
		result, err := c.ListGroupLabels(context.Background(), groups)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result[10]) != 1 {
			t.Fatalf("parent: expected 1 label, got %d", len(result[10]))
		}
		if _, ok := result[20]; ok {
			t.Errorf("child: expected no entry, got %d labels", len(result[20]))
		}
	})

	t.Run("paginates results", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		gomock.InOrder(
			// First page
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 1, Name: "bug"},
				}, &gl.Response{NextPage: 2}, nil),

			// Second page
			tc.MockGroupLabels.EXPECT().
				ListGroupLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.GroupLabel{
					{ID: 2, Name: "feature"},
				}, &gl.Response{}, nil),
		)

		groups := []*gl.Group{{ID: 10, Path: "mygroup", FullPath: "mygroup"}}
		result, err := c.ListGroupLabels(context.Background(), groups)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result[10]) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(result[10]))
		}
	})

	t.Run("skips nil groups", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		groups := []*gl.Group{nil}
		result, err := c.ListGroupLabels(context.Background(), groups)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d entries", len(result))
		}
	})
}

func TestListProjectLabels(t *testing.T) {
	t.Run("returns only project labels", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockLabels.EXPECT().
			ListLabels(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*gl.Label{
				{ID: 1, Name: "project-bug", IsProjectLabel: true},
				{ID: 2, Name: "inherited-label", IsProjectLabel: false},
				{ID: 3, Name: "project-feature", IsProjectLabel: true},
			}, &gl.Response{}, nil)

		projects := []*gl.Project{{ID: 1, Path: "proj"}}
		result, err := c.ListProjectLabels(context.Background(), projects)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result[1]) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(result[1]))
		}
		if result[1][0].Name != "project-bug" {
			t.Errorf("label name = %q, want %q", result[1][0].Name, "project-bug")
		}
		if result[1][1].Name != "project-feature" {
			t.Errorf("label name = %q, want %q", result[1][1].Name, "project-feature")
		}
	})

	t.Run("skips nil projects", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		projects := []*gl.Project{nil}
		result, err := c.ListProjectLabels(context.Background(), projects)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d entries", len(result))
		}
	})

	t.Run("paginates results", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		gomock.InOrder(
			tc.MockLabels.EXPECT().
				ListLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.Label{
					{ID: 1, Name: "bug", IsProjectLabel: true},
				}, &gl.Response{NextPage: 2}, nil),

			tc.MockLabels.EXPECT().
				ListLabels(gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]*gl.Label{
					{ID: 2, Name: "feature", IsProjectLabel: true},
				}, &gl.Response{}, nil),
		)

		projects := []*gl.Project{{ID: 1, Path: "proj"}}
		result, err := c.ListProjectLabels(context.Background(), projects)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result[1]) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(result[1]))
		}
	})
}
