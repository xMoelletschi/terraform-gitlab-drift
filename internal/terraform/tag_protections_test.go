package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func TestTagProtectionResourceName(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	tag := &gl.ProtectedTag{Name: "v1.0.0"}
	got := tagProtectionResourceName(project, tag)
	want := "my_group_my_project_v1_0_0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTagProtectionResourceNameWildcard(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	tag := &gl.ProtectedTag{Name: "v*"}
	got := tagProtectionResourceName(project, tag)
	want := "my_group_my_project_vwildcard"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTagProtectionResourceNamesCollision(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	// All three tag names normalize to the same base label and must stay unique.
	tags := []*gl.ProtectedTag{
		{Name: "v1.0"},
		{Name: "v1-0"},
		{Name: "v1_0"},
	}
	got := tagProtectionResourceNames(project, tags)
	want := []string{
		"my_group_my_project_v1_0",
		"my_group_my_project_v1_0_1",
		"my_group_my_project_v1_0_2",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestWriteTagProtections_Free(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	tags := []*gl.ProtectedTag{
		{
			Name: "v1.0.0",
			CreateAccessLevels: []*gl.TagAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
			},
		},
		{
			Name: "release-*",
			CreateAccessLevels: []*gl.TagAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTagProtections(project, tags, false, &buf); err != nil {
		t.Fatalf("WriteTagProtections error: %v", err)
	}

	compareGolden(t, "tag_protections_free.tf", buf.String())
}

func TestWriteTagProtections_Premium(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	tags := []*gl.ProtectedTag{
		{
			Name: "v1.0.0",
			CreateAccessLevels: []*gl.TagAccessDescription{
				{AccessLevel: gl.MaintainerPermissions},
				{AccessLevel: gl.DeveloperPermissions, UserID: 5},
				{AccessLevel: gl.DeveloperPermissions, UserID: 10},
			},
		},
		{
			Name: "release-*",
			CreateAccessLevels: []*gl.TagAccessDescription{
				{AccessLevel: gl.DeveloperPermissions},
				{AccessLevel: gl.DeveloperPermissions, GroupID: 42},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTagProtections(project, tags, true, &buf); err != nil {
		t.Fatalf("WriteTagProtections error: %v", err)
	}

	compareGolden(t, "tag_protections_premium.tf", buf.String())
}
