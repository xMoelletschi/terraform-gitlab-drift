package terraform

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func WriteGroupMembershipVariable(groups []*gl.Group, groupMembers gitlab.GroupMembers, w io.Writer) error {
	var b strings.Builder
	b.WriteString("variable \"gitlab_group_membership\" {\n")
	b.WriteString("  description = \"Assign gitlab users to groups.\"\n")
	b.WriteString("  default = {\n")

	for _, g := range groups {
		if g == nil {
			continue
		}
		members := groupMembers[g.ID]
		if len(members) == 0 {
			fmt.Fprintf(&b, "    \"%s\" = {}\n", g.FullPath)
		} else {
			fmt.Fprintf(&b, "    \"%s\" = {\n", g.FullPath)
			for _, m := range members {
				fmt.Fprintf(&b, "      \"%s\" = \"%s\"\n", m.Username, accessLevelToString(m.AccessLevel))
			}
			b.WriteString("    }\n")
		}
	}

	b.WriteString("  }\n")
	b.WriteString("}\n")

	_, err := w.Write(hclwrite.Format([]byte(b.String())))
	return err
}

func WriteGroupMembershipResource(group *gl.Group, w io.Writer) error {
	name := normalizeToTerraformName(group.Path)
	_, err := fmt.Fprintf(w, `resource "gitlab_group_membership" "%s" {
  for_each     = var.gitlab_group_membership["%s"]
  group_id     = gitlab_group.%s.id
  user_id      = gitlab_user.main[each.key].id
  access_level = each.value
}
`, name, group.FullPath, name)
	return err
}
