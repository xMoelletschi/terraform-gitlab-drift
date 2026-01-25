package gitlab

import (
	"context"
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	api *gl.Client
}

type Resources struct {
	Groups   []*gl.Group
	Projects []*gl.Project
	Users    []*gl.User
}

func NewClient(token, baseURL string) (*Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}
	return &Client{api: client}, nil
}

func (c *Client) FetchAll(ctx context.Context) (*Resources, error) {
	groups, err := c.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}

	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	return &Resources{
		Groups:   groups,
		Projects: projects,
		Users:    users,
	}, nil
}
