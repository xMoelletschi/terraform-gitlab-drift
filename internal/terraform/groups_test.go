package terraform

import (
	"bytes"
	"strings"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWriteGroupsDefaultsOmitted(t *testing.T) {
	groups := []*gl.Group{
		{
			Name:          "My Group",
			Path:          "my-group",
			EmailsEnabled: true,
		},
	}

	var buf bytes.Buffer
	if err := WriteGroups(groups, &buf, buildGroupRefMap(groups)); err != nil {
		t.Fatalf("WriteGroups error: %v", err)
	}

	compareGolden(t, "groups_minimal.tf", buf.String())
}

func TestWriteGroupsAllOptions(t *testing.T) {
	level := gl.MaintainerPermissions
	groups := []*gl.Group{
		{
			ID:            42,
			Name:          "Parent Group",
			Path:          "parent-group",
			EmailsEnabled: true,
		},
		{
			ID:                         100,
			Name:                       "Full Group",
			Path:                       "full-group",
			Description:                "Full group description",
			Visibility:                 gl.PrivateVisibility,
			ParentID:                   42,
			LFSEnabled:                 true,
			RequestAccessEnabled:       true,
			MembershipLock:             true,
			ShareWithGroupLock:         true,
			RequireTwoFactorAuth:       true,
			TwoFactorGracePeriod:       7,
			ProjectCreationLevel:       gl.OwnerProjectCreation,
			SubGroupCreationLevel:      gl.MaintainerSubGroupCreationLevelValue,
			AutoDevopsEnabled:          true,
			EmailsEnabled:              false,
			MentionsDisabled:           true,
			PreventForkingOutsideGroup: true,
			SharedRunnersSetting:       gl.DisabledAndOverridableSharedRunnersSettingValue,
			DefaultBranch:                                 "main",
			WikiAccessLevel:                               gl.PrivateAccessControl,
			IPRestrictionRanges:                           "10.0.0.0/24",
			MaxArtifactsSize:                              100,
			RepositoryStorage:                             "default",
			OnlyAllowMergeIfPipelineSucceeds:              true,
			AllowMergeOnSkippedPipeline:                   true,
			OnlyAllowMergeIfAllDiscussionsAreResolved:     true,
			DefaultBranchProtectionDefaults: &gl.BranchProtectionDefaults{
				AllowedToPush: []*gl.GroupAccessLevel{
					{AccessLevel: &level},
				},
				AllowedToMerge: []*gl.GroupAccessLevel{
					{AccessLevel: &level},
				},
				AllowForcePush:          true,
				DeveloperCanInitialPush: true,
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteGroups(groups, &buf, buildGroupRefMap(groups)); err != nil {
		t.Fatalf("WriteGroups error: %v", err)
	}

	compareGolden(t, "groups_full.tf", buf.String())
}

func TestWriteGroupsDefaultBranchProtectionOmitted(t *testing.T) {
	level := gl.MaintainerPermissions
	groups := []*gl.Group{
		{
			Name: "Group With Defaults",
			Path: "group-with-defaults",
			DefaultBranchProtectionDefaults: &gl.BranchProtectionDefaults{
				AllowedToPush: []*gl.GroupAccessLevel{
					{AccessLevel: &level}, // 40 = default
				},
				AllowedToMerge: []*gl.GroupAccessLevel{
					{AccessLevel: &level}, // 40 = default
				},
				AllowForcePush:          false, // default
				DeveloperCanInitialPush: false, // default
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteGroups(groups, &buf, buildGroupRefMap(groups)); err != nil {
		t.Fatalf("WriteGroups error: %v", err)
	}

	// Should not contain default_branch_protection_defaults block
	if strings.Contains(buf.String(), "default_branch_protection_defaults") {
		t.Error("default_branch_protection_defaults should be omitted when set to defaults")
	}
}
