package gitlab

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/testing"
	"go.uber.org/mock/gomock"
)

func TestGetDefaultBranch(t *testing.T) {
	t.Run("returns default branch", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockProjects.EXPECT().
			GetProject("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Project{DefaultBranch: "develop"}, nil, nil)

		branch, err := c.GetDefaultBranch(context.Background(), "mygroup/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "develop" {
			t.Errorf("got branch %q, want %q", branch, "develop")
		}
	})

	t.Run("falls back to main when empty", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockProjects.EXPECT().
			GetProject("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Project{DefaultBranch: ""}, nil, nil)

		branch, err := c.GetDefaultBranch(context.Background(), "mygroup/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("got branch %q, want %q", branch, "main")
		}
	})
}

func TestFindExistingDriftMR(t *testing.T) {
	t.Run("finds drift MR", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockMergeRequests.EXPECT().
			ListProjectMergeRequests("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return([]*gl.BasicMergeRequest{
				{IID: 1, SourceBranch: "feature/something"},
				{IID: 2, SourceBranch: "drift/update-2025-01-01", WebURL: "https://gitlab.com/mr/2"},
			}, &gl.Response{}, nil)

		mr, err := c.FindExistingDriftMR(context.Background(), "mygroup/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mr == nil {
			t.Fatal("expected MR, got nil")
		}
		if mr.IID != 2 {
			t.Errorf("got IID %d, want 2", mr.IID)
		}
	})

	t.Run("returns nil when none found", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockMergeRequests.EXPECT().
			ListProjectMergeRequests("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return([]*gl.BasicMergeRequest{
				{IID: 1, SourceBranch: "feature/something"},
			}, &gl.Response{}, nil)

		mr, err := c.FindExistingDriftMR(context.Background(), "mygroup/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mr != nil {
			t.Errorf("expected nil, got MR with IID %d", mr.IID)
		}
	})

	t.Run("paginates to find drift MR", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		// First page: no drift MR, has next page
		tc.MockMergeRequests.EXPECT().
			ListProjectMergeRequests("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return([]*gl.BasicMergeRequest{
				{IID: 1, SourceBranch: "feature/a"},
			}, &gl.Response{NextPage: 2}, nil)

		// Second page: has drift MR
		tc.MockMergeRequests.EXPECT().
			ListProjectMergeRequests("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return([]*gl.BasicMergeRequest{
				{IID: 3, SourceBranch: "drift/update-2025-02-01", WebURL: "https://gitlab.com/mr/3"},
			}, &gl.Response{}, nil)

		mr, err := c.FindExistingDriftMR(context.Background(), "mygroup/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mr == nil {
			t.Fatal("expected MR, got nil")
		}
		if mr.IID != 3 {
			t.Errorf("got IID %d, want 3", mr.IID)
		}
	})
}

func TestEnsureBranch(t *testing.T) {
	t.Run("branch already exists", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockBranches.EXPECT().
			GetBranch("mygroup/myproject", "drift/update-2025-01-01", gomock.Any()).
			Return(&gl.Branch{Name: "drift/update-2025-01-01"}, nil, nil)

		err := c.EnsureBranch(context.Background(), "mygroup/myproject", "drift/update-2025-01-01", "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("creates branch on 404", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockBranches.EXPECT().
			GetBranch("mygroup/myproject", "drift/update-2025-01-01", gomock.Any()).
			Return(nil, nil, &gl.ErrorResponse{
				Response: &http.Response{StatusCode: http.StatusNotFound},
			})

		tc.MockBranches.EXPECT().
			CreateBranch("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Branch{Name: "drift/update-2025-01-01"}, nil, nil)

		err := c.EnsureBranch(context.Background(), "mygroup/myproject", "drift/update-2025-01-01", "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("propagates auth error", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		tc.MockBranches.EXPECT().
			GetBranch("mygroup/myproject", "drift/update-2025-01-01", gomock.Any()).
			Return(nil, nil, &gl.ErrorResponse{
				Response: &http.Response{StatusCode: http.StatusForbidden},
			})

		err := c.EnsureBranch(context.Background(), "mygroup/myproject", "drift/update-2025-01-01", "main")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCommitDriftFiles(t *testing.T) {
	t.Run("creates new files", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "groups.tf"), []byte("resource {}"), 0644); err != nil {
			t.Fatal(err)
		}

		tc.MockRepositoryFiles.EXPECT().
			GetRawFile("mygroup/myproject", "groups.tf", gomock.Any(), gomock.Any()).
			Return(nil, nil, &gl.ErrorResponse{
				Response: &http.Response{StatusCode: http.StatusNotFound},
			})

		tc.MockCommits.EXPECT().
			CreateCommit("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Commit{ID: "abc123"}, nil, nil)

		committed, err := c.CommitDriftFiles(context.Background(), "mygroup/myproject", "drift/update", dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !committed {
			t.Error("expected commit to be created")
		}
	})

	t.Run("updates changed files", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "groups.tf"), []byte("resource { new }"), 0644); err != nil {
			t.Fatal(err)
		}

		tc.MockRepositoryFiles.EXPECT().
			GetRawFile("mygroup/myproject", "groups.tf", gomock.Any(), gomock.Any()).
			Return([]byte("resource { old }"), nil, nil)

		tc.MockCommits.EXPECT().
			CreateCommit("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Commit{ID: "def456"}, nil, nil)

		committed, err := c.CommitDriftFiles(context.Background(), "mygroup/myproject", "drift/update", dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !committed {
			t.Error("expected commit to be created")
		}
	})

	t.Run("skips identical files", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		content := []byte("resource {}")
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "groups.tf"), content, 0644); err != nil {
			t.Fatal(err)
		}

		tc.MockRepositoryFiles.EXPECT().
			GetRawFile("mygroup/myproject", "groups.tf", gomock.Any(), gomock.Any()).
			Return(content, nil, nil)

		committed, err := c.CommitDriftFiles(context.Background(), "mygroup/myproject", "drift/update", dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if committed {
			t.Error("expected no commit for identical files")
		}
	})

	t.Run("uses repo path prefix", func(t *testing.T) {
		tc := gitlabtesting.NewTestClient(t)
		c := NewClientFromAPI(tc.Client, "mygroup")

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "groups.tf"), []byte("resource {}"), 0644); err != nil {
			t.Fatal(err)
		}

		tc.MockRepositoryFiles.EXPECT().
			GetRawFile("mygroup/myproject", "terraform/groups.tf", gomock.Any(), gomock.Any()).
			Return(nil, nil, &gl.ErrorResponse{
				Response: &http.Response{StatusCode: http.StatusNotFound},
			})

		tc.MockCommits.EXPECT().
			CreateCommit("mygroup/myproject", gomock.Any(), gomock.Any()).
			Return(&gl.Commit{ID: "ghi789"}, nil, nil)

		committed, err := c.CommitDriftFiles(context.Background(), "mygroup/myproject", "drift/update", dir, "terraform")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !committed {
			t.Error("expected commit to be created")
		}
	})
}

func TestCreateDriftMR(t *testing.T) {
	tc := gitlabtesting.NewTestClient(t)
	c := NewClientFromAPI(tc.Client, "mygroup")

	tc.MockMergeRequests.EXPECT().
		CreateMergeRequest("mygroup/myproject", gomock.Any(), gomock.Any()).
		Return(&gl.MergeRequest{
			BasicMergeRequest: gl.BasicMergeRequest{
				IID:    1,
				WebURL: "https://gitlab.com/mygroup/myproject/-/merge_requests/1",
			},
		}, nil, nil)

	mr, err := c.CreateDriftMR(context.Background(), "mygroup/myproject", "drift/update-2025-01-01", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.WebURL != "https://gitlab.com/mygroup/myproject/-/merge_requests/1" {
		t.Errorf("got WebURL %q, want expected URL", mr.WebURL)
	}
}
