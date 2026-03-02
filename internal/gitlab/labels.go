package gitlab

import (
	"context"
	"fmt"

	gl "gitlab.com/gitlab-org/api/client-go"
)

// GroupLabels maps group IDs to their labels.
type GroupLabels = map[int64][]*gl.GroupLabel

// ProjectLabels maps project IDs to their labels.
type ProjectLabels = map[int64][]*gl.Label

func (c *Client) ListGroupLabels(ctx context.Context, groups []*gl.Group) (GroupLabels, error) {
	result := make(GroupLabels, len(groups))
	seen := make(map[int64]bool) // track label IDs already attributed to a group

	for _, g := range groups {
		if g == nil {
			continue
		}
		opts := &gl.ListGroupLabelsOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
			OnlyGroupLabels: gl.Ptr(true),
		}
		var labels []*gl.GroupLabel
		for {
			page, resp, err := c.api.GroupLabels.ListGroupLabels(g.ID, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing labels for group %d: %w", g.ID, err)
			}
			labels = append(labels, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		var owned []*gl.GroupLabel
		for _, l := range labels {
			if !seen[l.ID] {
				seen[l.ID] = true
				owned = append(owned, l)
			}
		}
		if len(owned) > 0 {
			result[g.ID] = owned
		}
	}

	return result, nil
}

func (c *Client) ListProjectLabels(ctx context.Context, projects []*gl.Project) (ProjectLabels, error) {
	result := make(ProjectLabels, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		opts := &gl.ListLabelsOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		var labels []*gl.Label
		for {
			page, resp, err := c.api.Labels.ListLabels(p.ID, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing labels for project %d: %w", p.ID, err)
			}
			for _, l := range page {
				if l.IsProjectLabel {
					labels = append(labels, l)
				}
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		if len(labels) > 0 {
			result[p.ID] = labels
		}
	}

	return result, nil
}
