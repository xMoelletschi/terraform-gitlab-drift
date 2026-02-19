package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestNormalizeToTerraformName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "mygroup",
			expected: "mygroup",
		},
		{
			name:     "with dashes",
			input:    "my-group",
			expected: "my_group",
		},
		{
			name:     "with slashes",
			input:    "my-group/my-project",
			expected: "my_group_my_project",
		},
		{
			name:     "uppercase",
			input:    "My-Group",
			expected: "my_group",
		},
		{
			name:     "mixed",
			input:    "Parent-Group/Sub-Group/My-Project",
			expected: "parent_group_sub_group_my_project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeToTerraformName(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeToTerraformName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestWriteAllSplitsProjectsByNamespace(t *testing.T) {
	resources := &gitlab.Resources{
		Groups: []*gl.Group{
			{
				ID:       10,
				Name:     "xdeveloperic",
				Path:     "xdeveloperic",
				FullPath: "xdeveloperic",
			},
			{
				ID:       20,
				Name:     "Sub Group",
				Path:     "sub-group",
				FullPath: "xdeveloperic/sub-group",
				ParentID: 10,
			},
		},
		Projects: []*gl.Project{
			{
				ID:   1,
				Name: "Project A",
				Path: "project-a",
				Namespace: &gl.ProjectNamespace{
					ID:       10,
					FullPath: "xdeveloperic",
				},
				SharedWithGroups: []gl.ProjectSharedWithGroup{
					{
						GroupID:          20,
						GroupName:        "sub-group",
						GroupFullPath:    "xdeveloperic/sub-group",
						GroupAccessLevel: 30,
					},
				},
			},
			{
				ID:   2,
				Name: "Project B",
				Path: "project-b",
				Namespace: &gl.ProjectNamespace{
					ID:       10,
					FullPath: "xdeveloperic",
				},
			},
			{
				ID:   3,
				Name: "Project C",
				Path: "project-c",
				Namespace: &gl.ProjectNamespace{
					ID:       20,
					FullPath: "xdeveloperic/sub-group",
				},
			},
		},
		GroupMembers: map[int64][]*gl.GroupMember{
			10: {
				{
					ID:          100,
					Username:    "admin",
					AccessLevel: gl.OwnerPermissions,
				},
			},
		},
	}

	dir := t.TempDir()
	if err := WriteAll(resources, dir, "xdeveloperic", nil); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	// group_membership.tf: variable only, no resources
	data, err := os.ReadFile(filepath.Join(dir, "group_membership.tf"))
	if err != nil {
		t.Fatalf("reading group_membership.tf: %v", err)
	}
	gmContent := string(data)
	if !strings.Contains(gmContent, `variable "gitlab_group_membership"`) {
		t.Error("group_membership.tf should contain gitlab_group_membership variable")
	}
	if !strings.Contains(gmContent, `"admin" = "owner"`) {
		t.Error("group_membership.tf should contain admin user entry")
	}
	if strings.Contains(gmContent, `resource`) {
		t.Error("group_membership.tf should NOT contain resource blocks")
	}

	// project_membership.tf: variable + helpers only, no resources
	data, err = os.ReadFile(filepath.Join(dir, "project_membership.tf"))
	if err != nil {
		t.Fatalf("reading project_membership.tf: %v", err)
	}
	pmContent := string(data)
	if !strings.Contains(pmContent, `variable "gitlab_project_membership"`) {
		t.Error("project_membership.tf should contain gitlab_project_membership variable")
	}
	if !strings.Contains(pmContent, `"xdeveloperic/sub-group" = "developer"`) {
		t.Error("project_membership.tf should contain shared group entry")
	}
	if !strings.Contains(pmContent, `data "gitlab_group" "by_projects"`) {
		t.Error("project_membership.tf should contain data source for group lookup")
	}
	if strings.Contains(pmContent, `resource`) {
		t.Error("project_membership.tf should NOT contain resource blocks")
	}

	// One file per namespace.
	if _, err := os.Stat(filepath.Join(dir, "xdeveloperic.tf")); err != nil {
		t.Fatalf("expected xdeveloperic.tf to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub_group.tf")); err != nil {
		t.Fatalf("expected sub_group.tf to exist: %v", err)
	}

	// Verify xdeveloperic.tf: group + membership resource + projects + share group resources
	data, err = os.ReadFile(filepath.Join(dir, "xdeveloperic.tf"))
	if err != nil {
		t.Fatalf("reading xdeveloperic.tf: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `gitlab_group" "xdeveloperic"`) {
		t.Error("xdeveloperic.tf should contain the xdeveloperic group")
	}
	if !strings.Contains(content, `gitlab_group_membership" "xdeveloperic"`) {
		t.Error("xdeveloperic.tf should contain group membership resource")
	}
	if !strings.Contains(content, `"xdeveloperic_project_a"`) {
		t.Error("xdeveloperic.tf should contain xdeveloperic_project_a")
	}
	if !strings.Contains(content, `gitlab_project_share_group" "xdeveloperic_project_a"`) {
		t.Error("xdeveloperic.tf should contain project share group resource")
	}

	// Verify ordering: group → membership → project → share group
	groupIdx := strings.Index(content, `gitlab_group" "xdeveloperic"`)
	membershipIdx := strings.Index(content, `gitlab_group_membership" "xdeveloperic"`)
	projectIdx := strings.Index(content, `gitlab_project" "xdeveloperic_project_a"`)
	shareIdx := strings.Index(content, `gitlab_project_share_group" "xdeveloperic_project_a"`)
	if groupIdx > membershipIdx {
		t.Error("group should appear before membership resource")
	}
	if membershipIdx > projectIdx {
		t.Error("membership resource should appear before projects")
	}
	if projectIdx > shareIdx {
		t.Error("projects should appear before share group resources")
	}

	// Verify sub_group.tf
	data, err = os.ReadFile(filepath.Join(dir, "sub_group.tf"))
	if err != nil {
		t.Fatalf("reading sub_group.tf: %v", err)
	}
	content = string(data)
	if !strings.Contains(content, `gitlab_group" "sub_group"`) {
		t.Error("sub_group.tf should contain the sub_group group")
	}
	if !strings.Contains(content, `gitlab_group_membership" "sub_group"`) {
		t.Error("sub_group.tf should contain group membership resource")
	}
	if !strings.Contains(content, `"sub_group_project_c"`) {
		t.Error("sub_group.tf should contain sub_group_project_c")
	}
}
