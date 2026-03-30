package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestBranchProtectionResourceName(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	branch := &gl.ProtectedBranch{Name: "main"}
	got := branchProtectionResourceName(project, branch)
	want := "my_group_my_project_main"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBranchProtectionResourceNameWildcard(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	branch := &gl.ProtectedBranch{Name: "release/*"}
	got := branchProtectionResourceName(project, branch)
	want := "my_group_my_project_release_wildcard"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteBranchProtections_Free(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	branches := []*gl.ProtectedBranch{
		{
			ID:   1,
			Name: "main",
			PushAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
			},
			MergeAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
			},
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: true, // should be ignored in free mode
		},
		{
			ID:   2,
			Name: "release/*",
			PushAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.NoPermissions},
			},
			MergeAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
			},
			AllowForcePush: true,
		},
	}

	var buf bytes.Buffer
	if err := WriteBranchProtections(project, branches, false, &buf); err != nil {
		t.Fatalf("WriteBranchProtections error: %v", err)
	}

	compareGolden(t, "branch_protections_free.tf", buf.String())
}

func TestWriteBranchProtections_Premium(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	branches := []*gl.ProtectedBranch{
		{
			ID:   1,
			Name: "main",
			PushAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
				{AccessLevel: gl.DeveloperPermissions, UserID: 5},
				{AccessLevel: gl.DeveloperPermissions, UserID: 10},
			},
			MergeAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
				{AccessLevel: gl.DeveloperPermissions, GroupID: 42},
			},
			UnprotectAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
			},
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: true,
		},
		{
			ID:   2,
			Name: "develop",
			PushAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
			},
			MergeAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
			},
			UnprotectAccessLevels: []*gl.BranchAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
			},
			AllowForcePush: true,
		},
	}

	var buf bytes.Buffer
	if err := WriteBranchProtections(project, branches, true, &buf); err != nil {
		t.Fatalf("WriteBranchProtections error: %v", err)
	}

	compareGolden(t, "branch_protections_premium.tf", buf.String())
}
