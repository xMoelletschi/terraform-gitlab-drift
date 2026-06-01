package terraform

import (
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func normalizeBranchName(s string) string {
	s = strings.ReplaceAll(s, "*", "wildcard")
	return normalizeName(s)
}

func branchProtectionResourceName(p *gl.Project, b *gl.ProtectedBranch) string {
	return projectResourceName(p) + "_" + normalizeBranchName(b.Name)
}

// branchProtectionResourceNames returns deterministic, collision-free terraform
// resource names for one project's protected branches.
func branchProtectionResourceNames(p *gl.Project, branches []*gl.ProtectedBranch) []string {
	bases := make([]string, len(branches))
	for i, b := range branches {
		bases[i] = branchProtectionResourceName(p, b)
	}
	return dedupeResourceNames(bases)
}

const adminAccessLevel gl.AccessLevelValue = 60

// protectionAccessLevel maps an access level to the string used by branch and
// tag protection resources (push/merge/unprotect and create access levels).
func protectionAccessLevel(level gl.AccessLevelValue) string {
	switch level {
	case gl.NoPermissions:
		return "no one"
	case gl.DeveloperPermissions:
		return "developer"
	case gl.MaintainerPermissions:
		return "maintainer"
	case adminAccessLevel:
		return "admin"
	default:
		return "maintainer"
	}
}

func isBaseAccessLevel(l *gl.BranchAccessDescription) bool {
	return l.UserID == 0 && l.GroupID == 0 && l.DeployKeyID == 0
}

// baseAccessLevel returns the access level from the entry that represents
// the overall branch access level (no specific user/group/deploy key).
func baseAccessLevel(levels []*gl.BranchAccessDescription) gl.AccessLevelValue {
	for _, l := range levels {
		if isBaseAccessLevel(l) {
			return l.AccessLevel
		}
	}
	if len(levels) > 0 {
		return levels[0].AccessLevel
	}
	return gl.MaintainerPermissions
}

func WriteBranchProtections(p *gl.Project, branches []*gl.ProtectedBranch, premium bool, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	projName := projectResourceName(p)
	names := branchProtectionResourceNames(p, branches)

	for i, b := range branches {
		if i > 0 {
			rootBody.AppendNewline()
		}
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_branch_protection", names[i]})
		body := block.Body()

		body.SetAttributeTraversal("project", hcl.Traversal{
			hcl.TraverseRoot{Name: "gitlab_project"},
			hcl.TraverseAttr{Name: projName},
			hcl.TraverseAttr{Name: "id"},
		})
		body.SetAttributeValue("branch", cty.StringVal(b.Name))

		pushLevel := baseAccessLevel(b.PushAccessLevels)
		mergeLevel := baseAccessLevel(b.MergeAccessLevels)

		body.SetAttributeValue("push_access_level", cty.StringVal(protectionAccessLevel(pushLevel)))
		body.SetAttributeValue("merge_access_level", cty.StringVal(protectionAccessLevel(mergeLevel)))

		if b.AllowForcePush {
			body.SetAttributeValue("allow_force_push", cty.True)
		}

		if premium {
			unprotectLevel := baseAccessLevel(b.UnprotectAccessLevels)
			body.SetAttributeValue("unprotect_access_level", cty.StringVal(protectionAccessLevel(unprotectLevel)))

			if b.CodeOwnerApprovalRequired {
				body.SetAttributeValue("code_owner_approval_required", cty.True)
			}

			writeAccessBlocks(body, "allowed_to_push", b.PushAccessLevels)
			writeAccessBlocks(body, "allowed_to_merge", b.MergeAccessLevels)
			writeAccessBlocks(body, "allowed_to_unprotect", b.UnprotectAccessLevels)
		}
	}

	_, err := w.Write(f.Bytes())
	return err
}

func writeAccessBlocks(body *hclwrite.Body, blockName string, levels []*gl.BranchAccessDescription) {
	for _, l := range levels {
		if isBaseAccessLevel(l) {
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
