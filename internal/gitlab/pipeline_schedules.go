package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

type PipelineSchedules = map[int64][]*gl.PipelineSchedule

// ListPipelineSchedules is the heaviest fetcher: the list endpoint omits
// variables, so each schedule needs an extra detail call. fetchPerProject runs
// the projects concurrently so it doesn't become the scan's long pole; the
// client's rate limiter still bounds the overall request rate.
func (c *Client) ListPipelineSchedules(ctx context.Context, projects []*gl.Project) (PipelineSchedules, error) {
	return fetchPerProject(ctx, c.concurrency, projects, func(ctx context.Context, p *gl.Project) ([]*gl.PipelineSchedule, bool, error) {
		slog.Debug("fetching pipeline schedules", "project", p.PathWithNamespace)
		opts := &gl.ListPipelineSchedulesOptions{
			ListOptions: gl.ListOptions{Page: 1, PerPage: 100},
		}
		schedules, err := paginate(&opts.ListOptions, "pipeline schedules", p.PathWithNamespace, func() ([]*gl.PipelineSchedule, *gl.Response, error) {
			return c.api.PipelineSchedules.ListPipelineSchedules(p.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, false, fmt.Errorf("listing pipeline schedules for project %d: %w", p.ID, err)
		}

		// Fetch detail for each schedule to get variables.
		var detailed []*gl.PipelineSchedule
		for _, s := range schedules {
			slog.Debug("fetching pipeline schedule detail", "project", p.PathWithNamespace, "schedule", s.ID)
			d, _, err := c.api.PipelineSchedules.GetPipelineSchedule(p.ID, s.ID, gl.WithContext(ctx))
			if err != nil {
				if skipInaccessible(err, "pipeline schedule detail", p.PathWithNamespace) {
					continue
				}
				return nil, false, fmt.Errorf("getting pipeline schedule %d for project %d: %w", s.ID, p.ID, err)
			}
			detailed = append(detailed, d)
		}

		return detailed, len(detailed) > 0, nil
	})
}
