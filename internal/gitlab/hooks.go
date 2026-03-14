package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	gl "gitlab.com/gitlab-org/api/client-go"
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
		var hooks []*gl.ProjectHook
		for {
			page, resp, err := c.api.Projects.ListProjectHooks(p.ID, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing hooks for project %d: %w", p.ID, err)
			}
			hooks = append(hooks, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
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
		var hooks []*gl.GroupHook
		for {
			page, resp, err := c.api.Groups.ListGroupHooks(g.ID, opts, gl.WithContext(ctx))
			if err != nil {
				var errResp *gl.ErrorResponse
				if errors.As(err, &errResp) && errResp.HasStatusCode(http.StatusForbidden) {
					slog.Warn("group hooks require Premium/Ultimate, skipping", "group", g.FullPath)
					break
				}
				return nil, fmt.Errorf("listing hooks for group %d: %w", g.ID, err)
			}
			hooks = append(hooks, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		if len(hooks) > 0 {
			result[g.ID] = hooks
		}
	}
	return result, nil
}
