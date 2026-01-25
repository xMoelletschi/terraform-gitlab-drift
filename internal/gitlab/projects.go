package gitlab

import (
	"context"
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func (c *Client) ListProjects(ctx context.Context) ([]*gl.Project, error) {
	var allProjects []*gl.Project

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

	return allProjects, nil
}
