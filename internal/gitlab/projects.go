package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func (c *Client) ListProjects(ctx context.Context) ([]*gl.Project, error) {
	var allProjects []*gl.Project

	if c.group != "" {
		opts := &gl.ListGroupProjectsOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
			IncludeSubGroups: gl.Ptr(true),
		}
		for {
			projects, resp, err := c.api.Groups.ListGroupProjects(c.group, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing group projects: %w", err)
			}
			allProjects = append(allProjects, projects...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		return filterDeletedProjects(allProjects), nil
	}

	opts := &gl.ListProjectsOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	for {
		projects, resp, err := c.api.Projects.ListProjects(opts, gl.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("listing projects: %w", err)
		}
		allProjects = append(allProjects, projects...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return filterDeletedProjects(allProjects), nil
}

func filterDeletedProjects(projects []*gl.Project) []*gl.Project {
	filtered := make([]*gl.Project, 0, len(projects))
	for _, p := range projects {
		if p.MarkedForDeletionOn != nil {
			slog.Debug("skipping project pending deletion", "project", p.PathWithNamespace)
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}
