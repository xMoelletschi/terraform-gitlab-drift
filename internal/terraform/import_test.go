package terraform

import (
	"bytes"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestGenerateImportCommandsNewGroup(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{ID: 10, Path: "my-group", FullPath: "my-group"},
		},
	}

	cmds := GenerateImportCommands(resources, nil, "my-group", nil)

	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Address != "gitlab_group.my_group" {
		t.Errorf("address = %q, want %q", cmds[0].Address, "gitlab_group.my_group")
	}
	if cmds[0].ID != "10" {
		t.Errorf("id = %q, want %q", cmds[0].ID, "10")
	}
}

func TestGenerateImportCommandsExistingGroupSkipped(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{ID: 10, Path: "my-group", FullPath: "my-group"},
		},
	}

	existing := map[string]bool{
		"gitlab_group.my_group": true,
	}

	cmds := GenerateImportCommands(resources, existing, "my-group", nil)

	if len(cmds) != 0 {
		t.Fatalf("expected 0 commands for existing resource, got %d", len(cmds))
	}
}

func TestGenerateImportCommandsNewProject(t *testing.T) {
	resources := &gitlab.Resources{
		Projects: []*gl.Project{
			{
				ID:   1,
				Path: "my-project",
				Namespace: &gl.ProjectNamespace{
					FullPath: "parent-group",
				},
			},
		},
	}

	cmds := GenerateImportCommands(resources, nil, "parent-group", nil)

	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Address != "gitlab_project.parent_group_my_project" {
		t.Errorf("address = %q, want %q", cmds[0].Address, "gitlab_project.parent_group_my_project")
	}
	if cmds[0].ID != "1" {
		t.Errorf("id = %q, want %q", cmds[0].ID, "1")
	}
}

func TestGenerateImportCommandsNewGroupMembership(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{ID: 10, Path: "my-group", FullPath: "my-group"},
		},
		GroupMembers: map[int64][]*gl.GroupMember{
			10: {
				{ID: 100, Username: "alice", AccessLevel: gl.DeveloperPermissions},
				{ID: 200, Username: "bob", AccessLevel: gl.MaintainerPermissions},
			},
		},
	}

	// Group itself exists, but membership resource does not.
	existing := map[string]bool{
		"gitlab_group.my_group": true,
	}

	cmds := GenerateImportCommands(resources, existing, "my-group", nil)

	// Expect 2 membership imports (one per member).
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
	if cmds[0].Address != `gitlab_group_membership.my_group["alice"]` {
		t.Errorf("address = %q, want %q", cmds[0].Address, `gitlab_group_membership.my_group["alice"]`)
	}
	if cmds[0].ID != "10:100" {
		t.Errorf("id = %q, want %q", cmds[0].ID, "10:100")
	}
	if cmds[1].Address != `gitlab_group_membership.my_group["bob"]` {
		t.Errorf("address = %q, want %q", cmds[1].Address, `gitlab_group_membership.my_group["bob"]`)
	}
	if cmds[1].ID != "10:200" {
		t.Errorf("id = %q, want %q", cmds[1].ID, "10:200")
	}
}

func TestGenerateImportCommandsNewProjectShareGroup(t *testing.T) {
	resources := &gitlab.Resources{
		Projects: []*gl.Project{
			{
				ID:   1,
				Path: "my-project",
				Namespace: &gl.ProjectNamespace{
					FullPath: "parent",
				},
				SharedWithGroups: []gl.ProjectSharedWithGroup{
					{GroupID: 20, GroupFullPath: "parent/shared"},
				},
			},
		},
	}

	// Project exists, but share group resource does not.
	existing := map[string]bool{
		"gitlab_project.parent_my_project": true,
	}

	cmds := GenerateImportCommands(resources, existing, "parent", nil)

	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Address != `gitlab_project_share_group.parent_my_project["parent/shared"]` {
		t.Errorf("address = %q, want %q", cmds[0].Address, `gitlab_project_share_group.parent_my_project["parent/shared"]`)
	}
	if cmds[0].ID != "1:20" {
		t.Errorf("id = %q, want %q", cmds[0].ID, "1:20")
	}
}

func TestGenerateImportCommandsSkipMemberships(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{ID: 10, Path: "grp", FullPath: "grp"},
		},
		Projects: []*gl.Project{
			{
				ID:   1,
				Path: "proj",
				Namespace: &gl.ProjectNamespace{
					FullPath: "grp",
				},
				SharedWithGroups: []gl.ProjectSharedWithGroup{
					{GroupID: 99, GroupFullPath: "grp/other"},
				},
			},
		},
		GroupMembers: map[int64][]*gl.GroupMember{
			10: {{ID: 100, Username: "user1"}},
		},
	}

	skipSet := skip.Set{"memberships": true}
	cmds := GenerateImportCommands(resources, nil, "grp", skipSet)

	// Should only have group + project imports, no memberships or share groups.
	for _, cmd := range cmds {
		if cmd.Address == `gitlab_group_membership.grp["user1"]` {
			t.Error("should not generate group membership import when memberships skipped")
		}
		if cmd.Address == `gitlab_project_share_group.grp_proj["grp/other"]` {
			t.Error("should not generate project share group import when memberships skipped")
		}
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands (group + project), got %d", len(cmds))
	}
}

func TestGenerateImportCommandsAllExisting(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{ID: 10, Path: "grp", FullPath: "grp"},
		},
		Projects: []*gl.Project{
			{
				ID:   1,
				Path: "proj",
				Namespace: &gl.ProjectNamespace{
					FullPath: "grp",
				},
			},
		},
		GroupMembers: map[int64][]*gl.GroupMember{
			10: {{ID: 100, Username: "user1"}},
		},
	}

	existing := map[string]bool{
		"gitlab_group.grp":            true,
		"gitlab_project.grp_proj":     true,
		"gitlab_group_membership.grp": true,
		// No gitlab_project_share_group needed since project has no shared groups.
	}

	cmds := GenerateImportCommands(resources, existing, "grp", nil)

	if len(cmds) != 0 {
		t.Errorf("expected 0 commands when all resources exist, got %d", len(cmds))
	}
}

func TestPrintImportCommands(t *testing.T) {
	cmds := []ImportCommand{
		{Address: "gitlab_group.my_group", ID: "10"},
		{Address: `gitlab_group_membership.my_group["alice"]`, ID: "10:100"},
	}

	var buf bytes.Buffer
	PrintImportCommands(&buf, cmds)

	want := "terraform import 'gitlab_group.my_group' '10'\n" +
		"terraform import 'gitlab_group_membership.my_group[\"alice\"]' '10:100'\n"

	if buf.String() != want {
		t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}
