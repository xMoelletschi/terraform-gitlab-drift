package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// ProjectHooks maps project IDs to their hooks.
type ProjectHooks = map[int64][]*gl.ProjectHook

// GroupHooks maps group IDs to their hooks.
type GroupHooks = map[int64][]*gl.GroupHook

func (c *Client) ListProjectHooks(ctx context.Context, projects []*gl.Project) (ProjectHooks, error) {
	result := make(ProjectHooks, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching project hooks", "project", p.PathWithNamespace)
		opts := &gl.ListProjectHooksOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		hooks, err := paginate(&opts.ListOptions, "project hooks", p.PathWithNamespace, func() ([]*gl.ProjectHook, *gl.Response, error) {
			return c.api.Projects.ListProjectHooks(p.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, fmt.Errorf("listing hooks for project %d: %w", p.ID, err)
		}
		if len(hooks) > 0 {
			result[p.ID] = hooks
		}
	}
	return result, nil
}

func (c *Client) ListGroupHooks(ctx context.Context, groups []*gl.Group) (GroupHooks, error) {
	result := make(GroupHooks, len(groups))

	for _, g := range groups {
		if g == nil {
			continue
		}
		slog.Debug("fetching group hooks", "group", g.FullPath)
		opts := &gl.ListGroupHooksOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		// Group hooks require Premium/Ultimate; a 403 here is expected on Free and
		// is skipped (via paginate's skipInaccessible) like any inaccessible group.
		hooks, err := paginate(&opts.ListOptions, "group hooks", g.FullPath, func() ([]*gl.GroupHook, *gl.Response, error) {
			return c.api.Groups.ListGroupHooks(g.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, fmt.Errorf("listing hooks for group %d: %w", g.ID, err)
		}
		if len(hooks) > 0 {
			result[g.ID] = hooks
		}
	}
	return result, nil
}
