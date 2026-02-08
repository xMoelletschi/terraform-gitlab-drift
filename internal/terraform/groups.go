package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go"
)

// WriteGroups writes GitLab groups as Terraform HCL resources.
func WriteGroups(groups []*gl.Group, w io.Writer, groupRefs groupRefMap) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for i, g := range groups {
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_group", normalizeToTerraformName(g.Path)})
		body := block.Body()

		// Required
		body.SetAttributeValue("name", cty.StringVal(g.Name))
		body.SetAttributeValue("path", cty.StringVal(g.Path))

		// Optional - only set if non-default
		if g.Description != "" {
			body.SetAttributeValue("description", cty.StringVal(g.Description))
		}
		if g.Visibility != gl.PublicVisibility {
			body.SetAttributeValue("visibility_level", cty.StringVal(string(g.Visibility)))
		}
		if g.ParentID != 0 {
			setGroupIDAttribute(body, "parent_id", g.ParentID, groupRefs)
		}
		if !g.LFSEnabled {
			body.SetAttributeValue("lfs_enabled", cty.BoolVal(g.LFSEnabled))
		}
		if !g.RequestAccessEnabled {
			body.SetAttributeValue("request_access_enabled", cty.BoolVal(g.RequestAccessEnabled))
		}
		if g.MembershipLock {
			body.SetAttributeValue("membership_lock", cty.BoolVal(g.MembershipLock))
		}
		if g.ShareWithGroupLock {
			body.SetAttributeValue("share_with_group_lock", cty.BoolVal(g.ShareWithGroupLock))
		}
		if g.RequireTwoFactorAuth {
			body.SetAttributeValue("require_two_factor_authentication", cty.BoolVal(g.RequireTwoFactorAuth))
		}
		if g.TwoFactorGracePeriod != 48 {
			body.SetAttributeValue("two_factor_grace_period", cty.NumberIntVal(g.TwoFactorGracePeriod))
		}
		if g.ProjectCreationLevel != gl.DeveloperProjectCreation {
			body.SetAttributeValue("project_creation_level", cty.StringVal(string(g.ProjectCreationLevel)))
		}
		if g.SubGroupCreationLevel != gl.MaintainerSubGroupCreationLevelValue {
			body.SetAttributeValue("subgroup_creation_level", cty.StringVal(string(g.SubGroupCreationLevel)))
		}
		if g.AutoDevopsEnabled {
			body.SetAttributeValue("auto_devops_enabled", cty.BoolVal(g.AutoDevopsEnabled))
		}
		if !g.EmailsEnabled {
			body.SetAttributeValue("emails_enabled", cty.BoolVal(g.EmailsEnabled))
		}
		if g.MentionsDisabled {
			body.SetAttributeValue("mentions_disabled", cty.BoolVal(g.MentionsDisabled))
		}
		if g.PreventForkingOutsideGroup {
			body.SetAttributeValue("prevent_forking_outside_group", cty.BoolVal(g.PreventForkingOutsideGroup))
		}
		if g.SharedRunnersSetting != gl.EnabledSharedRunnersSettingValue {
			body.SetAttributeValue("shared_runners_setting", cty.StringVal(string(g.SharedRunnersSetting)))
		}
		if g.DefaultBranch != "" {
			body.SetAttributeValue("default_branch", cty.StringVal(g.DefaultBranch))
		}
		if g.WikiAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("wiki_access_level", cty.StringVal(string(g.WikiAccessLevel)))
		}
		if g.IPRestrictionRanges != "" {
			body.SetAttributeValue("ip_restriction_ranges", cty.StringVal(g.IPRestrictionRanges))
		}
		if g.MaxArtifactsSize != 0 {
			body.SetAttributeValue("max_artifacts_size", cty.NumberIntVal(g.MaxArtifactsSize))
		}
		if g.RepositoryStorage != "" {
			body.SetAttributeValue("repository_storage", cty.StringVal(g.RepositoryStorage))
		}
		if g.OnlyAllowMergeIfPipelineSucceeds {
			body.SetAttributeValue("only_allow_merge_if_pipeline_succeeds", cty.BoolVal(g.OnlyAllowMergeIfPipelineSucceeds))
		}
		if g.AllowMergeOnSkippedPipeline {
			body.SetAttributeValue("allow_merge_on_skipped_pipeline", cty.BoolVal(g.AllowMergeOnSkippedPipeline))
		}
		if g.OnlyAllowMergeIfAllDiscussionsAreResolved {
			body.SetAttributeValue("only_allow_merge_if_all_discussions_are_resolved", cty.BoolVal(g.OnlyAllowMergeIfAllDiscussionsAreResolved))
		}

		// Nested block: default_branch_protection_defaults
		if !isDefaultBranchProtection(g.DefaultBranchProtectionDefaults) {
			body.AppendNewline()
			dbpd := g.DefaultBranchProtectionDefaults
			dbpdBlock := body.AppendNewBlock("default_branch_protection_defaults", nil)
			dbpdBody := dbpdBlock.Body()

			if len(dbpd.AllowedToPush) > 0 {
				levels := make([]cty.Value, len(dbpd.AllowedToPush))
				for i, al := range dbpd.AllowedToPush {
					if al.AccessLevel != nil {
						levels[i] = cty.NumberIntVal(int64(*al.AccessLevel))
					}
				}
				dbpdBody.SetAttributeValue("allowed_to_push", cty.ListVal(levels))
			}
			if len(dbpd.AllowedToMerge) > 0 {
				levels := make([]cty.Value, len(dbpd.AllowedToMerge))
				for i, al := range dbpd.AllowedToMerge {
					if al.AccessLevel != nil {
						levels[i] = cty.NumberIntVal(int64(*al.AccessLevel))
					}
				}
				dbpdBody.SetAttributeValue("allowed_to_merge", cty.ListVal(levels))
			}
			if dbpd.AllowForcePush {
				dbpdBody.SetAttributeValue("allow_force_push", cty.BoolVal(dbpd.AllowForcePush))
			}
			if dbpd.DeveloperCanInitialPush {
				dbpdBody.SetAttributeValue("developer_can_initial_push", cty.BoolVal(dbpd.DeveloperCanInitialPush))
			}
		}

		if i < len(groups)-1 {
			rootBody.AppendNewline()
		}
	}

	_, err := w.Write(f.Bytes())
	return err
}

// isDefaultBranchProtection checks if the branch protection settings are the GitLab defaults.
// Defaults: allowed_to_push=[40], allowed_to_merge=[40], allow_force_push=false, developer_can_initial_push=false
func isDefaultBranchProtection(dbpd *gl.BranchProtectionDefaults) bool {
	if dbpd == nil {
		return true
	}

	if len(dbpd.AllowedToPush) != 1 {
		return false
	}
	if dbpd.AllowedToPush[0].AccessLevel == nil || *dbpd.AllowedToPush[0].AccessLevel != gl.MaintainerPermissions {
		return false
	}

	if len(dbpd.AllowedToMerge) != 1 {
		return false
	}
	if dbpd.AllowedToMerge[0].AccessLevel == nil || *dbpd.AllowedToMerge[0].AccessLevel != gl.MaintainerPermissions {
		return false
	}

	if dbpd.AllowForcePush {
		return false
	}
	if dbpd.DeveloperCanInitialPush {
		return false
	}

	return true
}
