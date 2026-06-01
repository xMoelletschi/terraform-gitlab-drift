package terraform

import (
	"fmt"
	"io"
	"strings"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
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

// pipelineScheduleResourceNames returns deterministic, collision-free terraform
// resource names for one project's pipeline schedules.
func pipelineScheduleResourceNames(p *gl.Project, schedules []*gl.PipelineSchedule) []string {
	bases := make([]string, len(schedules))
	for i, s := range schedules {
		bases[i] = pipelineScheduleResourceName(p, s)
	}
	return dedupeResourceNames(bases)
}

// pipelineScheduleVariableResourceNames returns deterministic, collision-free
// terraform resource names for one schedule's variables, prefixed with the
// schedule's already-deduped resource name.
func pipelineScheduleVariableResourceNames(schedName string, vars []*gl.PipelineVariable) []string {
	bases := make([]string, len(vars))
	for i, v := range vars {
		bases[i] = schedName + "_" + normalizeName(v.Key)
	}
	return dedupeResourceNames(bases)
}

func WritePipelineSchedules(p *gl.Project, schedules []*gl.PipelineSchedule, w io.Writer) error {
	projName := projectResourceName(p)
	schedNames := pipelineScheduleResourceNames(p, schedules)
	for i, s := range schedules {
		schedName := schedNames[i]
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

		varNames := pipelineScheduleVariableResourceNames(schedName, s.Variables)
		for j, v := range s.Variables {
			if _, err := fmt.Fprintf(w, `
resource "gitlab_pipeline_schedule_variable" "%s" {
  project              = gitlab_project.%s.id
  pipeline_schedule_id = gitlab_pipeline_schedule.%s.pipeline_schedule_id
  key                  = "%s"
  value                = "%s"
  variable_type        = "%s"
}
`, varNames[j], projName, schedName, v.Key, v.Value, string(v.VariableType)); err != nil {
				return err
			}
		}
	}
	return nil
}
