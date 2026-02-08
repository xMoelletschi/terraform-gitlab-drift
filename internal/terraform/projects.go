package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	gl "gitlab.com/gitlab-org/api/client-go"
)

// WriteProjects writes GitLab projects as Terraform HCL resources.
func WriteProjects(projects []*gl.Project, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for _, p := range projects {
		block := rootBody.AppendNewBlock("resource", []string{"gitlab_project", normalizeToTerraformName(p.Path)})
		body := block.Body()

		// Required
		body.SetAttributeValue("name", cty.StringVal(p.Name))
		body.SetAttributeValue("path", cty.StringVal(p.Path))
		if p.Namespace != nil && p.Namespace.ID != 0 {
			body.SetAttributeValue("namespace_id", cty.NumberIntVal(p.Namespace.ID))
		}

		// Optional - only set if non-default
		if p.Description != "" {
			body.SetAttributeValue("description", cty.StringVal(p.Description))
		}
		if p.Visibility != gl.PublicVisibility {
			body.SetAttributeValue("visibility_level", cty.StringVal(string(p.Visibility)))
		}
		if p.ContainerRegistryAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("container_registry_access_level", cty.StringVal(string(p.ContainerRegistryAccessLevel)))
		}
		if p.IssuesAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("issues_access_level", cty.StringVal(string(p.IssuesAccessLevel)))
		}
		if p.RepositoryAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("repository_access_level", cty.StringVal(string(p.RepositoryAccessLevel)))
		}
		if p.MergeRequestsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("merge_requests_access_level", cty.StringVal(string(p.MergeRequestsAccessLevel)))
		}
		if p.ForkingAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("forking_access_level", cty.StringVal(string(p.ForkingAccessLevel)))
		}
		if p.WikiAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("wiki_access_level", cty.StringVal(string(p.WikiAccessLevel)))
		}
		if p.BuildsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("builds_access_level", cty.StringVal(string(p.BuildsAccessLevel)))
		}
		if p.SnippetsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("snippets_access_level", cty.StringVal(string(p.SnippetsAccessLevel)))
		}
		if p.PagesAccessLevel != gl.PrivateAccessControl {
			body.SetAttributeValue("pages_access_level", cty.StringVal(string(p.PagesAccessLevel)))
		}
		if p.ReleasesAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("releases_access_level", cty.StringVal(string(p.ReleasesAccessLevel)))
		}
		if p.AnalyticsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("analytics_access_level", cty.StringVal(string(p.AnalyticsAccessLevel)))
		}
		if p.OperationsAccessLevel != "" {
			body.SetAttributeValue("operations_access_level", cty.StringVal(string(p.OperationsAccessLevel)))
		}
		if p.EnvironmentsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("environments_access_level", cty.StringVal(string(p.EnvironmentsAccessLevel)))
		}
		if p.FeatureFlagsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("feature_flags_access_level", cty.StringVal(string(p.FeatureFlagsAccessLevel)))
		}
		if p.InfrastructureAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("infrastructure_access_level", cty.StringVal(string(p.InfrastructureAccessLevel)))
		}
		if p.MonitorAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("monitor_access_level", cty.StringVal(string(p.MonitorAccessLevel)))
		}
		if p.RequirementsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("requirements_access_level", cty.StringVal(string(p.RequirementsAccessLevel)))
		}
		if p.SecurityAndComplianceAccessLevel != gl.PrivateAccessControl {
			body.SetAttributeValue("security_and_compliance_access_level", cty.StringVal(string(p.SecurityAndComplianceAccessLevel)))
		}
		if p.ModelExperimentsAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("model_experiments_access_level", cty.StringVal(string(p.ModelExperimentsAccessLevel)))
		}
		if p.ModelRegistryAccessLevel != gl.EnabledAccessControl {
			body.SetAttributeValue("model_registry_access_level", cty.StringVal(string(p.ModelRegistryAccessLevel)))
		}
		if p.DefaultBranch != "" {
			body.SetAttributeValue("default_branch", cty.StringVal(p.DefaultBranch))
		}
		if len(p.Topics) > 0 {
			topics := make([]cty.Value, len(p.Topics))
			for i, t := range p.Topics {
				topics[i] = cty.StringVal(t)
			}
			body.SetAttributeValue("topics", cty.ListVal(topics))
		}
		if p.MergeMethod != gl.NoFastForwardMerge {
			body.SetAttributeValue("merge_method", cty.StringVal(string(p.MergeMethod)))
		}
		if p.SquashOption != gl.SquashOptionDefaultOff {
			body.SetAttributeValue("squash_option", cty.StringVal(string(p.SquashOption)))
		}
		if p.MergeCommitTemplate != "" {
			body.SetAttributeValue("merge_commit_template", cty.StringVal(p.MergeCommitTemplate))
		}
		if p.SquashCommitTemplate != "" {
			body.SetAttributeValue("squash_commit_template", cty.StringVal(p.SquashCommitTemplate))
		}
		if p.SuggestionCommitMessage != "" {
			body.SetAttributeValue("suggestion_commit_message", cty.StringVal(p.SuggestionCommitMessage))
		}
		if p.IssueBranchTemplate != "" {
			body.SetAttributeValue("issue_branch_template", cty.StringVal(p.IssueBranchTemplate))
		}
		if p.IssuesTemplate != "" {
			body.SetAttributeValue("issues_template", cty.StringVal(p.IssuesTemplate))
		}
		if p.MergeRequestsTemplate != "" {
			body.SetAttributeValue("merge_requests_template", cty.StringVal(p.MergeRequestsTemplate))
		}
		if p.BuildGitStrategy != "fetch" {
			body.SetAttributeValue("build_git_strategy", cty.StringVal(p.BuildGitStrategy))
		}
		if p.AutoCancelPendingPipelines != "enabled" {
			body.SetAttributeValue("auto_cancel_pending_pipelines", cty.StringVal(p.AutoCancelPendingPipelines))
		}
		if p.AutoDevopsDeployStrategy != "continuous" {
			body.SetAttributeValue("auto_devops_deploy_strategy", cty.StringVal(p.AutoDevopsDeployStrategy))
		}
		if p.CIConfigPath != "" {
			body.SetAttributeValue("ci_config_path", cty.StringVal(p.CIConfigPath))
		}
		if p.CIDefaultGitDepth != 20 {
			body.SetAttributeValue("ci_default_git_depth", cty.NumberIntVal(p.CIDefaultGitDepth))
		}
		if p.CIDeletePipelinesInSeconds != 0 {
			body.SetAttributeValue("ci_delete_pipelines_in_seconds", cty.NumberIntVal(p.CIDeletePipelinesInSeconds))
		}
		if len(p.CIIdTokenSubClaimComponents) > 0 {
			defaultComponents := []string{"project_path", "ref_type", "ref"}
			isDefault := len(p.CIIdTokenSubClaimComponents) == len(defaultComponents)
			if isDefault {
				for i, c := range defaultComponents {
					if p.CIIdTokenSubClaimComponents[i] != c {
						isDefault = false
						break
					}
				}
			}
			if !isDefault {
				components := make([]cty.Value, len(p.CIIdTokenSubClaimComponents))
				for i, c := range p.CIIdTokenSubClaimComponents {
					components[i] = cty.StringVal(c)
				}
				body.SetAttributeValue("ci_id_token_sub_claim_components", cty.ListVal(components))
			}
		}
		if p.CIRestrictPipelineCancellationRole != "" {
			body.SetAttributeValue("ci_restrict_pipeline_cancellation_role", cty.StringVal(string(p.CIRestrictPipelineCancellationRole)))
		}
		if p.CIPipelineVariablesMinimumOverrideRole != gl.CIPipelineVariablesDeveloperRole {
			body.SetAttributeValue("ci_pipeline_variables_minimum_override_role", cty.StringVal(p.CIPipelineVariablesMinimumOverrideRole))
		}
		if p.RepositoryStorage != "" {
			body.SetAttributeValue("repository_storage", cty.StringVal(p.RepositoryStorage))
		}
		if p.ImportURL != "" {
			body.SetAttributeValue("import_url", cty.StringVal(p.ImportURL))
		}
		if p.ExternalAuthorizationClassificationLabel != "" {
			body.SetAttributeValue("external_authorization_classification_label", cty.StringVal(p.ExternalAuthorizationClassificationLabel))
		}
		if p.BuildTimeout != 3600 {
			body.SetAttributeValue("build_timeout", cty.NumberIntVal(p.BuildTimeout))
		}
		if !p.SharedRunnersEnabled {
			body.SetAttributeValue("shared_runners_enabled", cty.BoolVal(p.SharedRunnersEnabled))
		}
		if !p.GroupRunnersEnabled {
			body.SetAttributeValue("group_runners_enabled", cty.BoolVal(p.GroupRunnersEnabled))
		}
		if !p.PackagesEnabled {
			body.SetAttributeValue("packages_enabled", cty.BoolVal(p.PackagesEnabled))
		}
		if !p.ServiceDeskEnabled {
			body.SetAttributeValue("service_desk_enabled", cty.BoolVal(p.ServiceDeskEnabled))
		}
		if !p.LFSEnabled {
			body.SetAttributeValue("lfs_enabled", cty.BoolVal(p.LFSEnabled))
		}
		if !p.RequestAccessEnabled {
			body.SetAttributeValue("request_access_enabled", cty.BoolVal(p.RequestAccessEnabled))
		}
		if !p.AutocloseReferencedIssues {
			body.SetAttributeValue("autoclose_referenced_issues", cty.BoolVal(p.AutocloseReferencedIssues))
		}
		if p.MergePipelinesEnabled {
			body.SetAttributeValue("merge_pipelines_enabled", cty.BoolVal(p.MergePipelinesEnabled))
		}
		if p.MergeTrainsEnabled {
			body.SetAttributeValue("merge_trains_enabled", cty.BoolVal(p.MergeTrainsEnabled))
		}
		if p.MergeTrainsSkipTrainAllowed {
			body.SetAttributeValue("merge_trains_skip_train_allowed", cty.BoolVal(p.MergeTrainsSkipTrainAllowed))
		}
		if p.Mirror {
			body.SetAttributeValue("mirror", cty.BoolVal(p.Mirror))
		}
		if p.MirrorUserID != 0 {
			body.SetAttributeValue("mirror_user_id", cty.NumberIntVal(p.MirrorUserID))
		}
		if p.MirrorTriggerBuilds {
			body.SetAttributeValue("mirror_trigger_builds", cty.BoolVal(p.MirrorTriggerBuilds))
		}
		if p.OnlyMirrorProtectedBranches {
			body.SetAttributeValue("only_mirror_protected_branches", cty.BoolVal(p.OnlyMirrorProtectedBranches))
		}
		if p.MirrorOverwritesDivergedBranches {
			body.SetAttributeValue("mirror_overwrites_diverged_branches", cty.BoolVal(p.MirrorOverwritesDivergedBranches))
		}
		if p.ResourceGroupDefaultProcessMode != gl.Unordered {
			body.SetAttributeValue("resource_group_default_process_mode", cty.StringVal(string(p.ResourceGroupDefaultProcessMode)))
		}
		if !p.KeepLatestArtifact {
			body.SetAttributeValue("keep_latest_artifact", cty.BoolVal(p.KeepLatestArtifact))
		}
		if p.MaxArtifactsSize != 0 {
			body.SetAttributeValue("max_artifacts_size", cty.NumberIntVal(p.MaxArtifactsSize))
		}
		if p.MergeRequestDefaultTargetSelf {
			body.SetAttributeValue("mr_default_target_self", cty.BoolVal(p.MergeRequestDefaultTargetSelf))
		}
		if p.PreventMergeWithoutJiraIssue {
			body.SetAttributeValue("prevent_merge_without_jira_issue", cty.BoolVal(p.PreventMergeWithoutJiraIssue))
		}
		if p.AllowPipelineTriggerApproveDeployment {
			body.SetAttributeValue("allow_pipeline_trigger_approve_deployment", cty.BoolVal(p.AllowPipelineTriggerApproveDeployment))
		}
		if p.AutoDuoCodeReviewEnabled {
			body.SetAttributeValue("auto_duo_code_review_enabled", cty.BoolVal(p.AutoDuoCodeReviewEnabled))
		}
		if !p.PrintingMergeRequestLinkEnabled {
			body.SetAttributeValue("printing_merge_request_link_enabled", cty.BoolVal(p.PrintingMergeRequestLinkEnabled))
		}
		if !p.CIForwardDeploymentEnabled {
			body.SetAttributeValue("ci_forward_deployment_enabled", cty.BoolVal(p.CIForwardDeploymentEnabled))
		}
		if !p.CIForwardDeploymentRollbackAllowed {
			body.SetAttributeValue("ci_forward_deployment_rollback_allowed", cty.BoolVal(p.CIForwardDeploymentRollbackAllowed))
		}
		if p.CIPushRepositoryForJobTokenAllowed {
			body.SetAttributeValue("ci_push_repository_for_job_token_allowed", cty.BoolVal(p.CIPushRepositoryForJobTokenAllowed))
		}
		if !p.CISeparatedCaches {
			body.SetAttributeValue("ci_separated_caches", cty.BoolVal(p.CISeparatedCaches))
		}
		if !p.EnforceAuthChecksOnUploads {
			body.SetAttributeValue("enforce_auth_checks_on_uploads", cty.BoolVal(p.EnforceAuthChecksOnUploads))
		}
		if !p.PublicJobs {
			body.SetAttributeValue("public_jobs", cty.BoolVal(p.PublicJobs))
		}
		if p.AllowMergeOnSkippedPipeline {
			body.SetAttributeValue("allow_merge_on_skipped_pipeline", cty.BoolVal(p.AllowMergeOnSkippedPipeline))
		}
		if p.OnlyAllowMergeIfPipelineSucceeds {
			body.SetAttributeValue("only_allow_merge_if_pipeline_succeeds", cty.BoolVal(p.OnlyAllowMergeIfPipelineSucceeds))
		}
		if p.OnlyAllowMergeIfAllDiscussionsAreResolved {
			body.SetAttributeValue("only_allow_merge_if_all_discussions_are_resolved", cty.BoolVal(p.OnlyAllowMergeIfAllDiscussionsAreResolved))
		}
		if !p.RemoveSourceBranchAfterMerge {
			body.SetAttributeValue("remove_source_branch_after_merge", cty.BoolVal(p.RemoveSourceBranchAfterMerge))
		}
		if p.ResolveOutdatedDiffDiscussions {
			body.SetAttributeValue("resolve_outdated_diff_discussions", cty.BoolVal(p.ResolveOutdatedDiffDiscussions))
		}
		if p.AutoDevopsEnabled {
			body.SetAttributeValue("auto_devops_enabled", cty.BoolVal(p.AutoDevopsEnabled))
		}
		if !p.EmailsEnabled {
			body.SetAttributeValue("emails_enabled", cty.BoolVal(p.EmailsEnabled))
		}

		if p.ContainerExpirationPolicy != nil && p.ContainerExpirationPolicy.Enabled {
			cep := p.ContainerExpirationPolicy
			cepBlock := body.AppendNewBlock("container_expiration_policy", nil)
			cepBody := cepBlock.Body()

			if cep.Cadence != "" {
				cepBody.SetAttributeValue("cadence", cty.StringVal(cep.Cadence))
			}
			if cep.KeepN != 0 {
				cepBody.SetAttributeValue("keep_n", cty.NumberIntVal(cep.KeepN))
			}
			if cep.OlderThan != "" {
				cepBody.SetAttributeValue("older_than", cty.StringVal(cep.OlderThan))
			}
			if cep.NameRegexDelete != "" {
				cepBody.SetAttributeValue("name_regex_delete", cty.StringVal(cep.NameRegexDelete))
			}
			if cep.NameRegexKeep != "" {
				cepBody.SetAttributeValue("name_regex_keep", cty.StringVal(cep.NameRegexKeep))
			}
			if cep.Enabled {
				cepBody.SetAttributeValue("enabled", cty.BoolVal(cep.Enabled))
			}
		}

		rootBody.AppendNewline()
	}

	_, err := w.Write(f.Bytes())
	return err
}
