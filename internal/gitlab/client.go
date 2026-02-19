package gitlab

import (
	"context"
	"fmt"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	api   *gl.Client
	group string
}

type Resources struct {
	Groups       []*gl.Group
	Projects     []*gl.Project
	GroupMembers GroupMembers
}

func NewClient(token, baseURL, group string) (*Client, error) {
	client, err := gl.NewClient(token, gl.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("creating GitLab client: %w", err)
	}
	return &Client{api: client, group: group}, nil
}

func (c *Client) FetchAll(ctx context.Context, skipSet skip.Set) (*Resources, error) {
	groups, err := c.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}

	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	var groupMembers GroupMembers
	if !skipSet.Has("memberships") {
		groupMembers, err = c.ListGroupMembers(ctx, groups)
		if err != nil {
			return nil, fmt.Errorf("listing group members: %w", err)
		}
	}

	return &Resources{
		Groups:       groups,
		Projects:     projects,
		GroupMembers: groupMembers,
	}, nil
}
