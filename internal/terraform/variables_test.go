package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestWriteGroupVariables(t *testing.T) {
	group := &gl.Group{ID: 10, Path: "my-group", FullPath: "my-group"}
	vars := []*gl.GroupVariable{
		{
			Key:              "API_URL",
			Value:            "https://api.example.com",
			VariableType:     "env_var",
			Protected:        false,
			Raw:              true,
			EnvironmentScope: "*",
			Description:      "API base URL",
		},
		{
			Key:              "DB_HOST",
			Value:            "db.example.com",
			VariableType:     "env_var",
			Protected:        true,
			Raw:              false,
			EnvironmentScope: "production",
		},
	}

	var buf bytes.Buffer
	if err := WriteGroupVariables(group, vars, &buf); err != nil {
		t.Fatalf("WriteGroupVariables error: %v", err)
	}

	compareGolden(t, "group_variables.tf", buf.String())
}

func TestWriteProjectVariables(t *testing.T) {
	project := &gl.Project{
		ID:   1,
		Path: "my-project",
		Namespace: &gl.ProjectNamespace{
			FullPath: "my-group",
		},
		PathWithNamespace: "my-group/my-project",
	}
	vars := []*gl.ProjectVariable{
		{
			Key:              "SECRET_KEY",
			Value:            "abc123",
			VariableType:     "env_var",
			Protected:        false,
			Raw:              false,
			EnvironmentScope: "*",
			Description:      "Secret key",
		},
	}

	var buf bytes.Buffer
	if err := WriteProjectVariables(project, vars, &buf); err != nil {
		t.Fatalf("WriteProjectVariables error: %v", err)
	}

	compareGolden(t, "project_variables.tf", buf.String())
}

func TestGroupVariableResourceName(t *testing.T) {
	g := &gl.Group{Path: "my-group"}
	tests := []struct {
		v    *gl.GroupVariable
		want string
	}{
		{&gl.GroupVariable{Key: "API_URL", EnvironmentScope: "*"}, "my_group_api_url"},
		{&gl.GroupVariable{Key: "DB_HOST", EnvironmentScope: "production"}, "my_group_db_host_production"},
	}
	for _, tt := range tests {
		got := groupVariableResourceName(g, tt.v)
		if got != tt.want {
			t.Errorf("groupVariableResourceName(%q, %q) = %q, want %q", tt.v.Key, tt.v.EnvironmentScope, got, tt.want)
		}
	}
}

func TestGroupVariableResourceNamesCollision(t *testing.T) {
	group := &gl.Group{Path: "my-group"}
	// Both keys normalize to the same label and must stay unique.
	vars := []*gl.GroupVariable{
		{Key: "MY.KEY", EnvironmentScope: "*"},
		{Key: "MY-KEY", EnvironmentScope: "*"},
	}
	got := groupVariableResourceNames(group, vars)
	want := []string{"my_group_my_key", "my_group_my_key_1"}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestProjectVariableResourceNamesCollision(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	vars := []*gl.ProjectVariable{
		{Key: "MY.KEY", EnvironmentScope: "*"},
		{Key: "MY-KEY", EnvironmentScope: "*"},
	}
	got := projectVariableResourceNames(project, vars)
	want := []string{"my_group_my_project_my_key", "my_group_my_project_my_key_1"}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
