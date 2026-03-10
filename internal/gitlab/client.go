package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	api   *gl.Client
	group string
}

type Resources struct {
	Groups        []*gl.Group
	Projects      []*gl.Project
	GroupMembers  GroupMembers
	GroupLabels   GroupLabels
	ProjectLabels ProjectLabels
}

func NewClientFromAPI(api *gl.Client, group string) *Client {
	return &Client{api: api, group: group}
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
	slog.Info("fetched groups", "count", len(groups))

	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	slog.Info("fetched projects", "count", len(projects))

	var groupMembers GroupMembers
	if !skipSet.Has("memberships") {
		groupMembers, err = c.ListGroupMembers(ctx, groups)
		if err != nil {
			return nil, fmt.Errorf("listing group members: %w", err)
		}
		slog.Info("fetched group members", "count", len(groupMembers))
	}

	var groupLabels GroupLabels
	var projectLabels ProjectLabels
	if !skipSet.Has("labels") {
		groupLabels, err = c.ListGroupLabels(ctx, groups)
		if err != nil {
			return nil, fmt.Errorf("listing group labels: %w", err)
		}
		slog.Info("fetched group labels", "count", len(groupLabels))

		projectLabels, err = c.ListProjectLabels(ctx, projects)
		if err != nil {
			return nil, fmt.Errorf("listing project labels: %w", err)
		}
		slog.Info("fetched project labels", "count", len(projectLabels))
	}

	return &Resources{
		Groups:        groups,
		Projects:      projects,
		GroupMembers:  groupMembers,
		GroupLabels:   groupLabels,
		ProjectLabels: projectLabels,
	}, nil
}
