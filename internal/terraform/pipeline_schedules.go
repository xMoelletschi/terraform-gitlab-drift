package terraform

import (
	"fmt"
	"io"
	"strings"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func normalizeName(s string) string {
	normalized := strings.ToLower(s)
	for _, ch := range []string{"/", "-", " ", "."} {
		normalized = strings.ReplaceAll(normalized, ch, "_")
	}
	// Collapse consecutive underscores.
	for strings.Contains(normalized, "__") {
		normalized = strings.ReplaceAll(normalized, "__", "_")
	}
	normalized = strings.TrimRight(normalized, "_")
	return normalized
}

func pipelineScheduleResourceName(p *gl.Project, s *gl.PipelineSchedule) string {
	return projectResourceName(p) + "_" + normalizeName(s.Description)
}

func pipelineScheduleVariableResourceName(p *gl.Project, s *gl.PipelineSchedule, v *gl.PipelineVariable) string {
	return pipelineScheduleResourceName(p, s) + "_" + normalizeName(v.Key)
}

func WritePipelineSchedules(p *gl.Project, schedules []*gl.PipelineSchedule, w io.Writer) error {
	projName := projectResourceName(p)
	for i, s := range schedules {
		schedName := pipelineScheduleResourceName(p, s)
		if i > 0 {
			if _, err := fmt.Fprint(w, "\n"); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, `resource "gitlab_pipeline_schedule" "%s" {
  project       = gitlab_project.%s.id
  description   = "%s"
  ref           = "%s"
  cron          = "%s"
  cron_timezone = "%s"
  active        = %t
}
`, schedName, projName, s.Description, s.Ref, s.Cron, s.CronTimezone, s.Active); err != nil {
			return err
		}

		for _, v := range s.Variables {
			varName := pipelineScheduleVariableResourceName(p, s, v)
			if _, err := fmt.Fprintf(w, `
resource "gitlab_pipeline_schedule_variable" "%s" {
  project              = gitlab_project.%s.id
  pipeline_schedule_id = gitlab_pipeline_schedule.%s.pipeline_schedule_id
  key                  = "%s"
  value                = "%s"
  variable_type        = "%s"
}
`, varName, projName, schedName, v.Key, v.Value, string(v.VariableType)); err != nil {
				return err
			}
		}
	}
	return nil
}
