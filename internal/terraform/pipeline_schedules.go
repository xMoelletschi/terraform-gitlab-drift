package terraform

import (
	"fmt"
	"io"
	"strings"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func normalizeName(s string) string {
	// Lowercase, then replace every character that is not valid in a terraform
	// identifier ([a-z0-9_]) with an underscore. This covers the common
	// separators (/ - space .) as well as anything else a free-text GitLab field
	// might contain (quotes, $, {, (, :, ', ...), which would otherwise produce
	// an invalid resource label.
	normalized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			return r
		default:
			return '_'
		}
	}, strings.ToLower(s))
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
	return buildResourceNames(schedules, func(s *gl.PipelineSchedule) string {
		return pipelineScheduleResourceName(p, s)
	})
}

// pipelineScheduleVariableResourceNames returns deterministic, collision-free
// terraform resource names for one schedule's variables, prefixed with the
// schedule's already-deduped resource name.
func pipelineScheduleVariableResourceNames(schedName string, vars []*gl.PipelineVariable) []string {
	return buildResourceNames(vars, func(v *gl.PipelineVariable) string {
		return schedName + "_" + normalizeName(v.Key)
	})
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
  description   = %s
  ref           = %s
  cron          = %s
  cron_timezone = %s
  active        = %t
}
`, schedName, projName, hclString(s.Description), hclString(s.Ref), hclString(s.Cron), hclString(s.CronTimezone), s.Active); err != nil {
			return err
		}

		varNames := pipelineScheduleVariableResourceNames(schedName, s.Variables)
		for j, v := range s.Variables {
			if _, err := fmt.Fprintf(w, `
resource "gitlab_pipeline_schedule_variable" "%s" {
  project              = gitlab_project.%s.id
  pipeline_schedule_id = gitlab_pipeline_schedule.%s.pipeline_schedule_id
  key                  = %s
  value                = %s
  variable_type        = %s
}
`, varNames[j], projName, schedName, hclString(v.Key), hclString(v.Value), hclString(string(v.VariableType))); err != nil {
				return err
			}
		}
	}
	return nil
}
