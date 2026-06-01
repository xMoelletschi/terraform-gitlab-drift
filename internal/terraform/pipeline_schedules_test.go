package terraform

import (
	"bytes"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
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

func TestPipelineScheduleResourceNamesCollision(t *testing.T) {
	project := &gl.Project{
		Path:      "my-project",
		Namespace: &gl.ProjectNamespace{FullPath: "my-group"},
	}
	// All three descriptions normalize to the same base label and must stay unique.
	schedules := []*gl.PipelineSchedule{
		{Description: "Nightly build"},
		{Description: "nightly-build"},
		{Description: "nightly.build"},
	}
	got := pipelineScheduleResourceNames(project, schedules)
	want := []string{
		"my_group_my_project_nightly_build",
		"my_group_my_project_nightly_build_1",
		"my_group_my_project_nightly_build_2",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPipelineScheduleVariableResourceNamesCollision(t *testing.T) {
	// Two variable keys within one schedule normalize to the same label.
	vars := []*gl.PipelineVariable{
		{Key: "DEPLOY_ENV"},
		{Key: "deploy-env"},
	}
	got := pipelineScheduleVariableResourceNames("my_group_my_project_nightly_build", vars)
	want := []string{
		"my_group_my_project_nightly_build_deploy_env",
		"my_group_my_project_nightly_build_deploy_env_1",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, got[i], want[i])
		}
	}
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
