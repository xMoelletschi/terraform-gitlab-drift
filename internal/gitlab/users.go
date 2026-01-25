package gitlab

import (
	"context"
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func (c *Client) ListUsers(ctx context.Context) ([]*gl.User, error) {
	var allUsers []*gl.User

	opts := &gl.ListUsersOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for {
		users, resp, err := c.api.Users.ListUsers(opts, gl.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("listing users: %w", err)
		}

		allUsers = append(allUsers, users...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allUsers, nil
}
