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
	}

	dir := t.TempDir()
	if err := WriteAll(resources, dir, "xdeveloperic"); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	// gitlab_groups.tf must NOT exist anymore (groups are in namespace files now).
	if _, err := os.Stat(filepath.Join(dir, "gitlab_groups.tf")); err == nil {
		t.Fatal("expected gitlab_groups.tf to NOT exist, but it does")
	}

	// One file per namespace.
	if _, err := os.Stat(filepath.Join(dir, "xdeveloperic.tf")); err != nil {
		t.Fatalf("expected xdeveloperic.tf to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub_group.tf")); err != nil {
		t.Fatalf("expected sub_group.tf to exist: %v", err)
	}

	// The old monolithic file must NOT exist.
	if _, err := os.Stat(filepath.Join(dir, "gitlab_projects.tf")); err == nil {
		t.Fatal("expected gitlab_projects.tf to NOT exist, but it does")
	}

	// Verify content: xdeveloperic.tf should contain the group + projects A and B.
	data, err := os.ReadFile(filepath.Join(dir, "xdeveloperic.tf"))
	if err != nil {
		t.Fatalf("reading xdeveloperic.tf: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `gitlab_group" "xdeveloperic"`) {
		t.Error("xdeveloperic.tf should contain the xdeveloperic group")
	}
	if !strings.Contains(content, `"xdeveloperic_project_a"`) {
		t.Error("xdeveloperic.tf should contain xdeveloperic_project_a")
	}
	if !strings.Contains(content, `"xdeveloperic_project_b"`) {
		t.Error("xdeveloperic.tf should contain xdeveloperic_project_b")
	}
	if strings.Contains(content, `"sub_group_project_c"`) {
		t.Error("xdeveloperic.tf should NOT contain sub_group_project_c")
	}

	// Verify content: sub_group.tf should contain the group + project C.
	data, err = os.ReadFile(filepath.Join(dir, "sub_group.tf"))
	if err != nil {
		t.Fatalf("reading sub_group.tf: %v", err)
	}
	content = string(data)
	if !strings.Contains(content, `gitlab_group" "sub_group"`) {
		t.Error("sub_group.tf should contain the sub_group group")
	}
	if !strings.Contains(content, `parent_id`) {
		t.Error("sub_group.tf should contain parent_id reference")
	}
	if !strings.Contains(content, `"sub_group_project_c"`) {
		t.Error("sub_group.tf should contain sub_group_project_c")
	}
	// Verify group is before projects (group should appear first in file)
	groupIdx := strings.Index(content, `gitlab_group" "sub_group"`)
	projectIdx := strings.Index(content, `gitlab_project" "sub_group_project_c"`)
	if groupIdx == -1 || projectIdx == -1 || groupIdx > projectIdx {
		t.Error("sub_group.tf: group should appear before projects")
	}
}
