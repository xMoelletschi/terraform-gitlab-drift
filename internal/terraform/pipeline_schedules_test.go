package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestWritePipelineSchedules(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}

	schedules := []*gl.PipelineSchedule{
		{
			ID:           10,
			Description:  "Nightly build",
			Ref:          "main",
			Cron:         "0 2 * * *",
			CronTimezone: "Europe/Vienna",
			Active:       true,
			Variables: []*gl.PipelineVariable{
				{Key: "DEPLOY_ENV", Value: "staging", VariableType: "env_var"},
			},
		},
		{
			ID:           20,
			Description:  "Weekly cleanup",
			Ref:          "main",
			Cron:         "0 0 * * 0",
			CronTimezone: "UTC",
			Active:       false,
		},
	}

	var buf bytes.Buffer
	if err := WritePipelineSchedules(project, schedules, &buf); err != nil {
		t.Fatalf("WritePipelineSchedules error: %v", err)
	}

	compareGolden(t, "pipeline_schedules.tf", buf.String())
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Nightly build", "nightly_build"},
		{"deploy/staging", "deploy_staging"},
		{"my-pipeline", "my_pipeline"},
		{"some.job.name", "some_job_name"},
		{"a--b//c  d..e", "a_b_c_d_e"},
		{"trailing-", "trailing"},
		{"DEPLOY_ENV", "deploy_env"},
	}
	for _, tt := range tests {
		got := normalizeName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
