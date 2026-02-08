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
			EmailsEnabled: true,
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
		Visibility:                                gl.PrivateVisibility,
		ContainerRegistryAccessLevel:              gl.EnabledAccessControl,
		IssuesAccessLevel:                         gl.EnabledAccessControl,
		RepositoryAccessLevel:                     gl.EnabledAccessControl,
		MergeRequestsAccessLevel:                  gl.EnabledAccessControl,
		ForkingAccessLevel:                        gl.PrivateAccessControl,
		WikiAccessLevel:                           gl.EnabledAccessControl,
		BuildsAccessLevel:                         gl.EnabledAccessControl,
		SnippetsAccessLevel:                       gl.EnabledAccessControl,
		PagesAccessLevel:                          gl.EnabledAccessControl,
		ReleasesAccessLevel:                       gl.EnabledAccessControl,
		AnalyticsAccessLevel:                      gl.EnabledAccessControl,
		OperationsAccessLevel:                     gl.EnabledAccessControl,
		EnvironmentsAccessLevel:                   gl.EnabledAccessControl,
		FeatureFlagsAccessLevel:                   gl.EnabledAccessControl,
		InfrastructureAccessLevel:                 gl.EnabledAccessControl,
		MonitorAccessLevel:                        gl.EnabledAccessControl,
		RequirementsAccessLevel:                   gl.EnabledAccessControl,
		SecurityAndComplianceAccessLevel:          gl.EnabledAccessControl,
		ModelExperimentsAccessLevel:               gl.EnabledAccessControl,
		ModelRegistryAccessLevel:                  gl.EnabledAccessControl,
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
		BuildCoverageRegex:                        "cov",
		BuildGitStrategy:                          "clone",
		AutoCancelPendingPipelines:                "enabled",
		AutoDevopsDeployStrategy:                  "continuous",
		CIConfigPath:                              ".gitlab-ci.yml",
		CIDefaultGitDepth:                         20,
		CIDeletePipelinesInSeconds:                3600,
		CIIdTokenSubClaimComponents:               []string{"project_path", "ref_type"},
		CIRestrictPipelineCancellationRole:        gl.PrivateAccessControl,
		CIPipelineVariablesMinimumOverrideRole:    gl.CIPipelineVariablesDeveloperRole,
		RepositoryStorage:                         "default",
		ImportURL:                                 "https://example.com/full.git",
		ExternalAuthorizationClassificationLabel:  "secret",
		BuildTimeout:                              600,
		IssuesEnabled:                             false,
		MergeRequestsEnabled:                      false,
		WikiEnabled:                               false,
		SnippetsEnabled:                           false,
		ContainerRegistryEnabled:                  false,
		SharedRunnersEnabled:                      false,
		GroupRunnersEnabled:                       false,
		PackagesEnabled:                           false,
		LFSEnabled:                                true,
		RequestAccessEnabled:                      true,
		AutocloseReferencedIssues:                 true,
		MergePipelinesEnabled:                     true,
		MergeTrainsEnabled:                        true,
		MergeTrainsSkipTrainAllowed:               true,
		Mirror:                                    true,
		MirrorUserID:                              99,
		MirrorTriggerBuilds:                       true,
		OnlyMirrorProtectedBranches:               true,
		MirrorOverwritesDivergedBranches:          true,
		ResourceGroupDefaultProcessMode:           gl.OldestFirst,
		KeepLatestArtifact:                        true,
		MaxArtifactsSize:                          123,
		MergeRequestDefaultTargetSelf:             true,
		PreventMergeWithoutJiraIssue:              true,
		AllowPipelineTriggerApproveDeployment:     true,
		AutoDuoCodeReviewEnabled:                  true,
		PrintingMergeRequestLinkEnabled:           true,
		CIForwardDeploymentEnabled:                true,
		CIForwardDeploymentRollbackAllowed:        true,
		CIPushRepositoryForJobTokenAllowed:        true,
		CISeparatedCaches:                         true,
		EnforceAuthChecksOnUploads:                true,
		PublicJobs:                                true,
		AllowMergeOnSkippedPipeline:               true,
		OnlyAllowMergeIfPipelineSucceeds:          true,
		OnlyAllowMergeIfAllDiscussionsAreResolved: true,
		RemoveSourceBranchAfterMerge:              true,
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

	legacy := &gl.Project{
		ID:   11,
		Name: "Legacy Policy",
		Path: "legacy-policy",
		Namespace: &gl.ProjectNamespace{
			ID: 321,
		},
		ContainerExpirationPolicy: &gl.ContainerExpirationPolicy{
			Cadence:   "2d",
			NameRegex: "legacy.*",
			Enabled:   true,
		},
	}

	var buf bytes.Buffer
	if err := WriteProjects([]*gl.Project{full, legacy}, &buf); err != nil {
		t.Fatalf("WriteProjects error: %v", err)
	}

	compareGolden(t, "projects_full.tf", buf.String())
}
