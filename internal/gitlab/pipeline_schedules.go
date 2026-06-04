package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/sync/errgroup"
)

type PipelineSchedules = map[int64][]*gl.PipelineSchedule

func (c *Client) ListPipelineSchedules(ctx context.Context, projects []*gl.Project) (PipelineSchedules, error) {
	result := make(PipelineSchedules, len(projects))
	var mu sync.Mutex

	// Pipeline schedules are the heaviest fetcher: the list endpoint omits
	// variables, so each schedule needs an extra detail call. To stop this from
	// becoming the long pole of the scan, fetch projects concurrently. The
	// client's rate limiter still bounds the overall request rate.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(resolveConcurrency(c.concurrency))

	for _, p := range projects {
		if p == nil {
			continue
		}
		g.Go(func() error {
			slog.Debug("fetching pipeline schedules", "project", p.PathWithNamespace)
			opts := &gl.ListPipelineSchedulesOptions{
				ListOptions: gl.ListOptions{Page: 1, PerPage: 100},
			}
			schedules, err := paginate(&opts.ListOptions, "pipeline schedules", p.PathWithNamespace, func() ([]*gl.PipelineSchedule, *gl.Response, error) {
				return c.api.PipelineSchedules.ListPipelineSchedules(p.ID, opts, gl.WithContext(gctx))
			})
			if err != nil {
				return fmt.Errorf("listing pipeline schedules for project %d: %w", p.ID, err)
			}

			// Fetch detail for each schedule to get variables.
			var detailed []*gl.PipelineSchedule
			for _, s := range schedules {
				slog.Debug("fetching pipeline schedule detail", "project", p.PathWithNamespace, "schedule", s.ID)
				d, _, err := c.api.PipelineSchedules.GetPipelineSchedule(p.ID, s.ID, gl.WithContext(gctx))
				if err != nil {
					if skipInaccessible(err, "pipeline schedule detail", p.PathWithNamespace) {
						continue
					}
					return fmt.Errorf("getting pipeline schedule %d for project %d: %w", s.ID, p.ID, err)
				}
				detailed = append(detailed, d)
			}

			if len(detailed) > 0 {
				mu.Lock()
				result[p.ID] = detailed
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
