package gitlab

import (
	"context"
	"sync"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/sync/errgroup"
)

// fetchPerProject fetches a per-project resource concurrently and returns it
// keyed by project ID. fetch is called once per non-nil project (bounded by
// concurrency) and returns the value plus whether to keep it — results with
// keep == false are dropped, matching the per-resource "store only non-empty"
// convention. The first error cancels the rest and is returned.
//
// The client's rate limiter bounds the overall request rate, so this is safe to
// nest inside FetchAll's fetcher-level concurrency. Use it for any
// project-scoped fetcher that issues more than one request per project (e.g.
// job-token scopes, pipeline schedules); single-request fetchers are fine
// running serially under FetchAll's concurrency.
func fetchPerProject[R any](
	ctx context.Context,
	concurrency int,
	projects []*gl.Project,
	fetch func(context.Context, *gl.Project) (R, bool, error),
) (map[int64]R, error) {
	result := make(map[int64]R, len(projects))
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(resolveConcurrency(concurrency))

	for _, p := range projects {
		if p == nil {
			continue
		}
		g.Go(func() error {
			val, keep, err := fetch(gctx, p)
			if err != nil {
				return err
			}
			if keep {
				mu.Lock()
				result[p.ID] = val
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
