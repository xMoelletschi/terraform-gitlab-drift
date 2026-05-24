package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func (c *Client) ListGroups(ctx context.Context) ([]*gl.Group, error) {
	var allGroups []*gl.Group

	if c.group != "" {
		slog.Debug("fetching groups", "group", c.group)
		// First, fetch the main group itself
		mainGroup, _, err := c.api.Groups.GetGroup(c.group, nil, gl.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("getting main group: %w", err)
		}
		allGroups = append(allGroups, mainGroup)

		// Then fetch all descendant groups
		opts := &gl.ListDescendantGroupsOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		for {
			groups, resp, err := c.api.Groups.ListDescendantGroups(c.group, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing descendant groups: %w", err)
			}
			allGroups = append(allGroups, groups...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		return filterDeletedGroups(allGroups), nil
	}

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
	return filterDeletedGroups(allGroups), nil
}

func filterDeletedGroups(groups []*gl.Group) []*gl.Group {
	filtered := make([]*gl.Group, 0, len(groups))
	for _, g := range groups {
		if g.MarkedForDeletionOn != nil {
			slog.Debug("skipping group pending deletion", "group", g.FullPath)
			continue
		}
		filtered = append(filtered, g)
	}
	return filtered
}
