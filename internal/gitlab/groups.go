package gitlab

import (
	"context"
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func (c *Client) ListGroups(ctx context.Context) ([]*gl.Group, error) {
	var allGroups []*gl.Group

	opts := &gl.ListGroupsOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		groups, resp, err := c.api.Groups.ListGroups(opts, gl.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("listing groups: %w", err)
		}

		allGroups = append(allGroups, groups...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}
	return allGroups, nil
}
