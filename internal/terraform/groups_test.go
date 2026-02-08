package terraform

import (
	"bytes"
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
	if err := WriteGroups(groups, &buf); err != nil {
		t.Fatalf("WriteGroups error: %v", err)
	}

	compareGolden(t, "groups_minimal.tf", buf.String())
}

func TestWriteGroupsAllOptions(t *testing.T) {
	level := gl.MaintainerPermissions
	groups := []*gl.Group{
		{
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
			DefaultBranch:              "main",
			WikiAccessLevel:            gl.PrivateAccessControl,
			IPRestrictionRanges:        "10.0.0.0/24",
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
	if err := WriteGroups(groups, &buf); err != nil {
		t.Fatalf("WriteGroups error: %v", err)
	}

	compareGolden(t, "groups_full.tf", buf.String())
}
