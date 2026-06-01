package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// ProjectVariables maps project IDs to their CI/CD variables.
type ProjectVariables = map[int64][]*gl.ProjectVariable

// GroupVariables maps group IDs to their CI/CD variables.
type GroupVariables = map[int64][]*gl.GroupVariable

func (c *Client) ListProjectVariables(ctx context.Context, projects []*gl.Project) (ProjectVariables, error) {
	result := make(ProjectVariables, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching project variables", "project", p.PathWithNamespace)
		opts := &gl.ListProjectVariablesOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		all, err := paginate(&opts.ListOptions, "project variables", p.PathWithNamespace, func() ([]*gl.ProjectVariable, *gl.Response, error) {
			return c.api.ProjectVariables.ListVariables(p.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, fmt.Errorf("listing variables for project %d: %w", p.ID, err)
		}
		var vars []*gl.ProjectVariable
		for _, v := range all {
			if v.Masked {
				slog.Debug("skipping masked variable", "key", v.Key, "project", p.PathWithNamespace)
				continue
			}
			if v.Hidden {
				slog.Debug("skipping hidden variable", "key", v.Key, "project", p.PathWithNamespace)
				continue
			}
			if v.VariableType == gl.FileVariableType {
				slog.Debug("skipping file variable", "key", v.Key, "project", p.PathWithNamespace)
				continue
			}
			vars = append(vars, v)
		}
		if len(vars) > 0 {
			result[p.ID] = vars
		}
	}

	return result, nil
}

func (c *Client) ListGroupVariables(ctx context.Context, groups []*gl.Group) (GroupVariables, error) {
	result := make(GroupVariables, len(groups))

	for _, g := range groups {
		if g == nil {
			continue
		}
		slog.Debug("fetching group variables", "group", g.FullPath)
		opts := &gl.ListGroupVariablesOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		all, err := paginate(&opts.ListOptions, "group variables", g.FullPath, func() ([]*gl.GroupVariable, *gl.Response, error) {
			return c.api.GroupVariables.ListVariables(g.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, fmt.Errorf("listing variables for group %d: %w", g.ID, err)
		}
		var vars []*gl.GroupVariable
		for _, v := range all {
			if v.Masked {
				slog.Debug("skipping masked variable", "key", v.Key, "group", g.FullPath)
				continue
			}
			if v.Hidden {
				slog.Debug("skipping hidden variable", "key", v.Key, "group", g.FullPath)
				continue
			}
			if v.VariableType == gl.FileVariableType {
				slog.Debug("skipping file variable", "key", v.Key, "group", g.FullPath)
				continue
			}
			vars = append(vars, v)
		}
		if len(vars) > 0 {
			result[g.ID] = vars
		}
	}

	return result, nil
}
