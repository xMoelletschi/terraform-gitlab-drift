package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestNormalizeHookURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://hooks.slack.com/services/T123", "hooks_slack_com_services_t123"},
		{"http://example.com/webhook", "example_com_webhook"},
		{"https://example.com:8080/hook", "example_com_8080_hook"},
		{"example.com/plain", "example_com_plain"},
	}
	for _, tt := range tests {
		got := normalizeHookURL(tt.input)
		if got != tt.want {
			t.Errorf("normalizeHookURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestWriteProjectHooks(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	hooks := []*gl.ProjectHook{
		{
			ID:                    100,
			URL:                   "https://hooks.slack.com/services/T123",
			Name:                  "Slack notifications",
			Description:           "Posts to #dev",
			EnableSSLVerification: true,
			PushEvents:            true,
			MergeRequestsEvents:   true,
		},
		{
			ID:                    200,
			URL:                   "https://example.com/webhook",
			EnableSSLVerification: true,
			PushEvents:            false,
			TagPushEvents:         true,
			PipelineEvents:        true,
		},
	}

	var buf bytes.Buffer
	if err := WriteProjectHooks(project, hooks, &buf); err != nil {
		t.Fatalf("WriteProjectHooks error: %v", err)
	}

	compareGolden(t, "project_hooks.tf", buf.String())
}

func TestWriteGroupHooks(t *testing.T) {
	group := &gl.Group{
		ID:       10,
		Path:     "my-group",
		FullPath: "my-group",
	}

	hooks := []*gl.GroupHook{
		{
			ID:                    300,
			URL:                   "https://example.com/webhook",
			EnableSSLVerification: true,
			PushEvents:            true,
			SubGroupEvents:        true,
		},
		{
			ID:                    400,
			URL:                   "https://ci.internal.io/notify",
			Name:                  "CI notify",
			EnableSSLVerification: false,
			PushEvents:            true,
			MergeRequestsEvents:   true,
			EmojiEvents:           true,
			FeatureFlagEvents:     true,
		},
	}

	var buf bytes.Buffer
	if err := WriteGroupHooks(group, hooks, &buf); err != nil {
		t.Fatalf("WriteGroupHooks error: %v", err)
	}

	compareGolden(t, "group_hooks.tf", buf.String())
}
