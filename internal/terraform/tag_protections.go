package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func tagProtectionResourceName(p *gl.Project, t *gl.ProtectedTag) string {
	return projectResourceName(p) + "_" + normalizeBranchName(t.Name)
}

func isBaseTagAccessLevel(l *gl.TagAccessDescription) bool {
	return l.UserID == 0 && l.GroupID == 0 && l.DeployKeyID == 0
}

// baseTagAccessLevel returns the access level from the entry that represents
// the overall create access level (no specific user/group/deploy key).
func baseTagAccessLevel(levels []*gl.TagAccessDescription) gl.AccessLevelValue {
	for _, l := range levels {
		if isBaseTagAccessLevel(l) {
			return l.AccessLevel
		}
	}
	if len(levels) > 0 {
		return levels[0].AccessLevel
	}
	return gl.MaintainerPermissions
}

func WriteTagProtections(p *gl.Project, tags []*gl.ProtectedTag, premium bool, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	projName := projectResourceName(p)

	for i, t := range tags {
		if i > 0 {
			rootBody.AppendNewline()
		}
		name := tagProtectionResourceName(p, t)
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_tag_protection", name})
		body := block.Body()

		body.SetAttributeTraversal("project", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_project"},
			hcl.TraverseAttr{Name: projName},
			hcl.TraverseAttr{Name: "id"},
		})
		body.SetAttributeValue("tag", cty.StringVal(t.Name))

		createLevel := baseTagAccessLevel(t.CreateAccessLevels)
		body.SetAttributeValue("create_access_level", cty.StringVal(branchProtectionAccessLevel(createLevel)))

		if premium {
			writeTagAccessBlocks(body, "allowed_to_create", t.CreateAccessLevels)
		}
	}

	_, err := w.Write(f.Bytes())
	return err
}

func writeTagAccessBlocks(body *hclwrite.Body, blockName string, levels []*gl.TagAccessDescription) {
	for _, l := range levels {
		if isBaseTagAccessLevel(l) {
			continue
		}
		nested := body.AppendNewBlock(blockName, nil)
		nb := nested.Body()
		if l.UserID != 0 {
			nb.SetAttributeValue("user_id", cty.NumberIntVal(l.UserID))
		}
		if l.GroupID != 0 {
			nb.SetAttributeValue("group_id", cty.NumberIntVal(l.GroupID))
		}
		if l.DeployKeyID != 0 {
			nb.SetAttributeValue("deploy_key_id", cty.NumberIntVal(l.DeployKeyID))
		}
	}
}
