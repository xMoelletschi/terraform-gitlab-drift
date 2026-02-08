package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type groupRefMap map[int64]string

func buildGroupRefMap(groups []*gl.Group) groupRefMap {
	if len(groups) == 0 {
		return nil
	}
	refs := make(groupRefMap, len(groups))
	for _, g := range groups {
		if g == nil || g.ID == 0 {
			continue
		}
		refs[g.ID] = normalizeToTerraformName(g.Path)
	}
	return refs
}

func setGroupIDAttribute(body *hclwrite.Body, attr string, id int64, refs groupRefMap) {
	if id == 0 {
		return
	}
	if refs != nil {
		if name, ok := refs[id]; ok && name != "" {
			body.SetAttributeTraversal(attr, hcl.Traversal{
				hcl.TraverseRoot{Name: "gitlab_group"},
				hcl.TraverseAttr{Name: name},
				hcl.TraverseAttr{Name: "id"},
			})
			return
		}
	}
	body.SetAttributeValue(attr, cty.NumberIntVal(int64(id)))
}
