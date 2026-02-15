package terraform

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

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
	f := hclwrite.NewEmptyFile()
	name := normalizeToTerraformName(group.Path)
	block := f.Body().AppendNewBlock("resource", []string{"gitlab_group_membership", name})
	body := block.Body()

	body.SetAttributeTraversal("for_each", hcl.Traversal{
		hcl.TraverseRoot{Name: "var"},
		hcl.TraverseAttr{Name: "gitlab_group_membership"},
		hcl.TraverseIndex{Key: cty.StringVal(group.FullPath)},
	})

	body.SetAttributeTraversal("group_id", hcl.Traversal{
		hcl.TraverseRoot{Name: "gitlab_group"},
		hcl.TraverseAttr{Name: name},
		hcl.TraverseAttr{Name: "id"},
	})

	body.SetAttributeRaw("user_id", tokensForIndexedTraversal(
		hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_user"},
			hcl.TraverseAttr{Name: "main"},
		},
		hcl.Traversal{
			hcl.TraverseRoot{Name: "each"},
			hcl.TraverseAttr{Name: "key"},
		},
		"id",
	))

	body.SetAttributeTraversal("access_level", hcl.Traversal{
		hcl.TraverseRoot{Name: "each"},
		hcl.TraverseAttr{Name: "value"},
	})

	_, err := w.Write(f.Bytes())
	return err
}
