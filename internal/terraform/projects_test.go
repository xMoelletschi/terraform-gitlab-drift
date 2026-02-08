package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWriteProjectsDefaultsOmitted(t *testing.T) {
	projects := []*gl.Project{
		{
			ID:   1,
			Name: "Minimal Project",
			Path: "minimal-project",
			Namespace: &gl.ProjectNamespace{
				ID: 7,
			},
			Visibility:                             gl.PublicVisibility,
			ContainerRegistryAccessLevel:           gl.EnabledAccessControl,
			IssuesAccessLevel:                      gl.EnabledAccessControl,
			RepositoryAccessLevel:                  gl.EnabledAccessControl,
			MergeRequestsAccessLevel:               gl.EnabledAccessControl,
			ForkingAccessLevel:                     gl.EnabledAccessControl,
			WikiAccessLevel:                        gl.EnabledAccessControl,
			BuildsAccessLevel:                      gl.EnabledAccessControl,
			SnippetsAccessLevel:                    gl.EnabledAccessControl,
			PagesAccessLevel:                       gl.PrivateAccessControl,
			ReleasesAccessLevel:                    gl.EnabledAccessControl,
			AnalyticsAccessLevel:                   gl.EnabledAccessControl,
			OperationsAccessLevel:                  "",
			EnvironmentsAccessLevel:                gl.EnabledAccessControl,
			FeatureFlagsAccessLevel:                gl.EnabledAccessControl,
			InfrastructureAccessLevel:              gl.EnabledAccessControl,
			MonitorAccessLevel:                     gl.EnabledAccessControl,
			RequirementsAccessLevel:                gl.EnabledAccessControl,
			SecurityAndComplianceAccessLevel:       gl.PrivateAccessControl,
			ModelExperimentsAccessLevel:            gl.EnabledAccessControl,
			ModelRegistryAccessLevel:               gl.EnabledAccessControl,
			MergeMethod:                            gl.NoFastForwardMerge,
			SquashOption:                           gl.SquashOptionDefaultOff,
			BuildGitStrategy:                       "fetch",
			AutoCancelPendingPipelines:             "enabled",
			AutoDevopsDeployStrategy:               "continuous",
			CIDefaultGitDepth:                      20,
			BuildTimeout:                           3600,
			CIIdTokenSubClaimComponents:            []string{"project_path", "ref_type", "ref"},
			CIPipelineVariablesMinimumOverrideRole: gl.CIPipelineVariablesDeveloperRole,
			ResourceGroupDefaultProcessMode:        gl.Unordered,
			SharedRunnersEnabled:                   true,
			GroupRunnersEnabled:                    true,
			PackagesEnabled:                        true,
			ServiceDeskEnabled:                     true,
			LFSEnabled:                             true,
			RequestAccessEnabled:                   true,
			AutocloseReferencedIssues:              true,
			KeepLatestArtifact:                     true,
			PrintingMergeRequestLinkEnabled:        true,
			CIForwardDeploymentEnabled:             true,
			CIForwardDeploymentRollbackAllowed:     true,
			CISeparatedCaches:                      true,
			EnforceAuthChecksOnUploads:             true,
			PublicJobs:                             true,
			EmailsEnabled:                          true,
			RemoveSourceBranchAfterMerge:           true,
		},
	}

	var buf bytes.Buffer
	if err := WriteProjects(projects, &buf); err != nil {
		t.Fatalf("WriteProjects error: %v", err)
	}

	compareGolden(t, "projects_minimal.tf", buf.String())
}

func TestWriteProjectsAllOptions(t *testing.T) {
	full := &gl.Project{
		ID:   10,
		Name: "Full Project",
		Path: "full-project",
		Namespace: &gl.ProjectNamespace{
			ID: 123,
		},
		Description:                               "Full project description",
		Visibility:                                gl.InternalVisibility,
		ContainerRegistryAccessLevel:              gl.DisabledAccessControl,
		IssuesAccessLevel:                         gl.DisabledAccessControl,
		RepositoryAccessLevel:                     gl.PrivateAccessControl,
		MergeRequestsAccessLevel:                  gl.DisabledAccessControl,
		ForkingAccessLevel:                        gl.PrivateAccessControl,
		WikiAccessLevel:                           gl.DisabledAccessControl,
		BuildsAccessLevel:                         gl.PrivateAccessControl,
		SnippetsAccessLevel:                       gl.DisabledAccessControl,
		PagesAccessLevel:                          gl.PublicAccessControl,
		ReleasesAccessLevel:                       gl.PrivateAccessControl,
		AnalyticsAccessLevel:                      gl.PrivateAccessControl,
		OperationsAccessLevel:                     gl.PrivateAccessControl,
		EnvironmentsAccessLevel:                   gl.PrivateAccessControl,
		FeatureFlagsAccessLevel:                   gl.PrivateAccessControl,
		InfrastructureAccessLevel:                 gl.PrivateAccessControl,
		MonitorAccessLevel:                        gl.PrivateAccessControl,
		RequirementsAccessLevel:                   gl.PrivateAccessControl,
		SecurityAndComplianceAccessLevel:          gl.DisabledAccessControl,
		ModelExperimentsAccessLevel:               gl.PrivateAccessControl,
		ModelRegistryAccessLevel:                  gl.PrivateAccessControl,
		DefaultBranch:                             "main",
		Topics:                                    []string{"topic-a", "topic-b"},
		MergeMethod:                               gl.RebaseMerge,
		SquashOption:                              gl.SquashOptionAlways,
		MergeCommitTemplate:                       "merge {{title}}",
		SquashCommitTemplate:                      "squash {{title}}",
		SuggestionCommitMessage:                   "suggest {{title}}",
		IssueBranchTemplate:                       "issue-{{iid}}",
		IssuesTemplate:                            "issues.md",
		MergeRequestsTemplate:                     "merge_requests.md",
		BuildGitStrategy:                          "clone",
		AutoCancelPendingPipelines:                "disabled",
		AutoDevopsDeployStrategy:                  "timed_incremental",
		CIConfigPath:                              ".gitlab-ci.yml",
		CIDefaultGitDepth:                         50,
		CIDeletePipelinesInSeconds:                3600,
		CIIdTokenSubClaimComponents:               []string{"project_path", "ref_type"},
		CIRestrictPipelineCancellationRole:        gl.PrivateAccessControl,
		CIPipelineVariablesMinimumOverrideRole:    gl.CiPipelineVariablesMaintainerRole,
		RepositoryStorage:                         "default",
		ImportURL:                                 "https://example.com/full.git",
		ExternalAuthorizationClassificationLabel:  "secret",
		BuildTimeout:                              600,
		SharedRunnersEnabled:                      false,
		GroupRunnersEnabled:                       false,
		PackagesEnabled:                           false,
		LFSEnabled:                                false,
		RequestAccessEnabled:                      false,
		AutocloseReferencedIssues:                 false,
		ServiceDeskEnabled:                        false,
		MergePipelinesEnabled:                     true,
		MergeTrainsEnabled:                        true,
		MergeTrainsSkipTrainAllowed:               true,
		Mirror:                                    true,
		MirrorUserID:                              99,
		MirrorTriggerBuilds:                       true,
		OnlyMirrorProtectedBranches:               true,
		MirrorOverwritesDivergedBranches:          true,
		ResourceGroupDefaultProcessMode:           gl.NewestFirst,
		KeepLatestArtifact:                        false,
		MaxArtifactsSize:                          123,
		MergeRequestDefaultTargetSelf:             true,
		PreventMergeWithoutJiraIssue:              true,
		AllowPipelineTriggerApproveDeployment:     true,
		AutoDuoCodeReviewEnabled:                  true,
		PrintingMergeRequestLinkEnabled:           false,
		CIForwardDeploymentEnabled:                false,
		CIForwardDeploymentRollbackAllowed:        false,
		CIPushRepositoryForJobTokenAllowed:        true,
		CISeparatedCaches:                         false,
		EnforceAuthChecksOnUploads:                false,
		PublicJobs:                                false,
		AllowMergeOnSkippedPipeline:               true,
		OnlyAllowMergeIfPipelineSucceeds:          true,
		OnlyAllowMergeIfAllDiscussionsAreResolved: true,
		RemoveSourceBranchAfterMerge:              false,
		ResolveOutdatedDiffDiscussions:            true,
		AutoDevopsEnabled:                         true,
		EmailsEnabled:                             false,
		ContainerExpirationPolicy: &gl.ContainerExpirationPolicy{
			Cadence:         "1d",
			KeepN:           5,
			OlderThan:       "7d",
			NameRegexDelete: ".*",
			NameRegexKeep:   "keep.*",
			Enabled:         true,
		},
	}

	var buf bytes.Buffer
	if err := WriteProjects([]*gl.Project{full}, &buf); err != nil {
		t.Fatalf("WriteProjects error: %v", err)
	}

	compareGolden(t, "projects_full.tf", buf.String())
}
