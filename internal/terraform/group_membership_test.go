package terraform

import (
	"bytes"
	"testing"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWriteGroupMembershipVariable(t *testing.T) {
	groups := []*gl.Group{
		{ID: 10, Path: "my-group", FullPath: "my-group"},
		{ID: 20, Path: "sub-group", FullPath: "my-group/sub-group"},
	}
	members := gitlab.GroupMembers{
		10: {
			{ID: 42, Username: "jdoe", AccessLevel: gl.DeveloperPermissions},
			{ID: 43, Username: "asmith", AccessLevel: gl.MaintainerPermissions},
		},
	}

	var buf bytes.Buffer
	if err := WriteGroupMembershipVariable(groups, members, &buf); err != nil {
		t.Fatalf("WriteGroupMembershipVariable error: %v", err)
	}

	compareGolden(t, "group_membership_variable.tf", buf.String())
}

func TestWriteGroupMembershipResource(t *testing.T) {
	group := &gl.Group{ID: 10, Path: "my-group", FullPath: "my-group"}

	var buf bytes.Buffer
	if err := WriteGroupMembershipResource(group, &buf); err != nil {
		t.Fatalf("WriteGroupMembershipResource error: %v", err)
	}

	compareGolden(t, "group_membership_resource.tf", buf.String())
}
