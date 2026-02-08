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
			{Name: "xdeveloperic", Path: "xdeveloperic"},
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
	if err := WriteAll(resources, dir); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	// gitlab_groups.tf must exist.
	if _, err := os.Stat(filepath.Join(dir, "gitlab_groups.tf")); err != nil {
		t.Fatalf("expected gitlab_groups.tf to exist: %v", err)
	}

	// One file per namespace.
	if _, err := os.Stat(filepath.Join(dir, "xdeveloperic.tf")); err != nil {
		t.Fatalf("expected xdeveloperic.tf to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "xdeveloperic_sub_group.tf")); err != nil {
		t.Fatalf("expected xdeveloperic_sub_group.tf to exist: %v", err)
	}

	// The old monolithic file must NOT exist.
	if _, err := os.Stat(filepath.Join(dir, "gitlab_projects.tf")); err == nil {
		t.Fatal("expected gitlab_projects.tf to NOT exist, but it does")
	}

	// Verify content: xdeveloperic.tf should contain both projects A and B.
	data, err := os.ReadFile(filepath.Join(dir, "xdeveloperic.tf"))
	if err != nil {
		t.Fatalf("reading xdeveloperic.tf: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, `"project_a"`) {
		t.Error("xdeveloperic.tf should contain project_a")
	}
	if !strings.Contains(content, `"project_b"`) {
		t.Error("xdeveloperic.tf should contain project_b")
	}
	if strings.Contains(content, `"project_c"`) {
		t.Error("xdeveloperic.tf should NOT contain project_c")
	}

	// Verify content: xdeveloperic_sub_group.tf should contain project C only.
	data, err = os.ReadFile(filepath.Join(dir, "xdeveloperic_sub_group.tf"))
	if err != nil {
		t.Fatalf("reading xdeveloperic_sub_group.tf: %v", err)
	}
	content = string(data)
	if !strings.Contains(content, `"project_c"`) {
		t.Error("xdeveloperic_sub_group.tf should contain project_c")
	}
}
