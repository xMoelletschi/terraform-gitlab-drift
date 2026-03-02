package terraform

import (
	"bytes"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWriteGroupLabelVariable(t *testing.T) {
	groups := []*gl.Group{
		{ID: 10, Path: "my-group", FullPath: "my-group"},
		{ID: 20, Path: "sub-group", FullPath: "my-group/sub-group"},
	}
	labels := gitlab.GroupLabels{
		10: {
			{Name: "bug", Color: "#FF0000", Description: "Bug reports"},
			{Name: "feature", Color: "#00FF00", Description: "Feature requests"},
		},
	}

	var buf bytes.Buffer
	if err := WriteGroupLabelVariable(groups, labels, &buf); err != nil {
		t.Fatalf("WriteGroupLabelVariable error: %v", err)
	}

	compareGolden(t, "group_label_variable.tf", buf.String())
}

func TestWriteGroupLabelResource(t *testing.T) {
	group := &gl.Group{ID: 10, Path: "my-group", FullPath: "my-group"}

	var buf bytes.Buffer
	if err := WriteGroupLabelResource(group, &buf); err != nil {
		t.Fatalf("WriteGroupLabelResource error: %v", err)
	}

	compareGolden(t, "group_label_resource.tf", buf.String())
}

func TestWriteProjectLabelVariable(t *testing.T) {
	projects := []*gl.Project{
		{
			ID:   1,
			Path: "my-project",
			Namespace: &gl.ProjectNamespace{
				FullPath: "my-group",
			},
			PathWithNamespace: "my-group/my-project",
		},
		{
			ID:   2,
			Path: "other-project",
			Namespace: &gl.ProjectNamespace{
				FullPath: "my-group",
			},
			PathWithNamespace: "my-group/other-project",
		},
	}
	labels := gitlab.ProjectLabels{
		1: {
			{Name: "urgent", Color: "#FF0000", Description: "Urgent issues", IsProjectLabel: true},
		},
	}

	var buf bytes.Buffer
	if err := WriteProjectLabelVariable(projects, labels, &buf); err != nil {
		t.Fatalf("WriteProjectLabelVariable error: %v", err)
	}

	compareGolden(t, "project_label_variable.tf", buf.String())
}

func TestWriteProjectLabelResource(t *testing.T) {
	project := &gl.Project{
		ID:   1,
		Path: "my-project",
		Namespace: &gl.ProjectNamespace{
			FullPath: "my-group",
		},
		PathWithNamespace: "my-group/my-project",
	}

	var buf bytes.Buffer
	if err := WriteProjectLabelResource(project, &buf); err != nil {
		t.Fatalf("WriteProjectLabelResource error: %v", err)
	}

	compareGolden(t, "project_label_resource.tf", buf.String())
}
