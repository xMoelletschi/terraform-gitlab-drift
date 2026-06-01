package terraform

import (
	"bytes"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func assertParseableHCL(t *testing.T, src string) {
	t.Helper()
	_, diags := hclsyntax.ParseConfig([]byte(src), "gen.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("generated HCL does not parse:\n%s\ndiagnostics: %v", src, diags)
	}
}

func TestWritePipelineSchedules_EscapesSpecialChars(t *testing.T) {
	project := &gl.Project{
		ID:                1,
		Path:              "my-project",
		Namespace:         &gl.ProjectNamespace{FullPath: "my-group"},
		PathWithNamespace: "my-group/my-project",
	}
	schedules := []*gl.PipelineSchedule{
		{
			ID:           10,
			Description:  "Nightly build", // safe: also used as the resource name
			Ref:          "main",
			Cron:         "0 2 * * *",
			CronTimezone: "Europe/Vienna",
			Active:       true,
			Variables: []*gl.PipelineVariable{
				// A CI variable value is free text and not part of any name.
				{Key: "MSG", Value: `say "hi"` + "\nline2 ${x}", VariableType: "env_var"},
			},
		},
	}

	var buf bytes.Buffer
	if err := WritePipelineSchedules(project, schedules, &buf); err != nil {
		t.Fatalf("WritePipelineSchedules error: %v", err)
	}
	assertParseableHCL(t, buf.String())
}

func TestWriteGroupLabelVariable_EscapesSpecialChars(t *testing.T) {
	groups := []*gl.Group{{ID: 10, Path: "my-group", FullPath: "my-group"}}
	groupLabels := gitlab.GroupLabels{
		10: {{Name: `urgent "p1"`, Color: "#ff0000", Description: `fix ${now}`}},
	}

	var buf bytes.Buffer
	if err := WriteGroupLabelVariable(groups, groupLabels, &buf); err != nil {
		t.Fatalf("WriteGroupLabelVariable error: %v", err)
	}
	assertParseableHCL(t, buf.String())
}
