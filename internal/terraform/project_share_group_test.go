package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWriteProjectMembershipVariable(t *testing.T) {
	projects := []*gl.Project{
		{
			ID:   1,
			Path: "project-a",
			Namespace: &gl.ProjectNamespace{
				FullPath: "my-group",
			},
			SharedWithGroups: []gl.ProjectSharedWithGroup{
				{GroupID: 20, GroupName: "sub-group", GroupFullPath: "my-group/sub-group", GroupAccessLevel: 30},
			},
		},
		{
			ID:   2,
			Path: "project-b",
			Namespace: &gl.ProjectNamespace{
				FullPath: "my-group",
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteProjectMembershipVariable(projects, &buf); err != nil {
		t.Fatalf("WriteProjectMembershipVariable error: %v", err)
	}

	compareGolden(t, "project_membership_variable.tf", buf.String())
}

func TestWriteProjectShareGroupResource(t *testing.T) {
	project := &gl.Project{
		ID:   1,
		Path: "project-a",
		Namespace: &gl.ProjectNamespace{
			FullPath: "my-group",
		},
	}

	var buf bytes.Buffer
	if err := WriteProjectShareGroupResource(project, &buf); err != nil {
		t.Fatalf("WriteProjectShareGroupResource error: %v", err)
	}

	compareGolden(t, "project_share_group_resource.tf", buf.String())
}

func TestWriteProjectMembershipHelpers(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteProjectMembershipHelpers(&buf); err != nil {
		t.Fatalf("WriteProjectMembershipHelpers error: %v", err)
	}

	compareGolden(t, "project_membership_helpers.tf", buf.String())
}
