package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go"
)

type PipelineSchedules = map[int64][]*gl.PipelineSchedule

func (c *Client) ListPipelineSchedules(ctx context.Context, projects []*gl.Project) (PipelineSchedules, error) {
	result := make(PipelineSchedules, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching pipeline schedules", "project", p.PathWithNamespace)
		opts := &gl.ListPipelineSchedulesOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		var schedules []*gl.PipelineSchedule
		for {
			page, resp, err := c.api.PipelineSchedules.ListPipelineSchedules(p.ID, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing pipeline schedules for project %d: %w", p.ID, err)
			}
			schedules = append(schedules, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		// Fetch detail for each schedule to get variables.
		var detailed []*gl.PipelineSchedule
		for _, s := range schedules {
			slog.Debug("fetching pipeline schedule detail", "project", p.PathWithNamespace, "schedule", s.ID)
			d, _, err := c.api.PipelineSchedules.GetPipelineSchedule(p.ID, s.ID, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("getting pipeline schedule %d for project %d: %w", s.ID, p.ID, err)
			}
			detailed = append(detailed, d)
		}

		if len(detailed) > 0 {
			result[p.ID] = detailed
		}
	}
	return result, nil
}
