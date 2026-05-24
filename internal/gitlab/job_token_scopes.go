package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

type JobTokenScope struct {
	InboundEnabled  bool
	AllowedProjects []*gl.Project
	AllowedGroups   []*gl.Group
}

type JobTokenScopes = map[int64]*JobTokenScope

func (c *Client) ListJobTokenScopes(ctx context.Context, projects []*gl.Project) (JobTokenScopes, error) {
	result := make(JobTokenScopes, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching job token scopes", "project", p.PathWithNamespace)

		settings, _, err := c.api.JobTokenScope.GetProjectJobTokenAccessSettings(p.ID, gl.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("getting job token access settings for project %d: %w", p.ID, err)
		}

		projectOpts := &gl.GetJobTokenInboundAllowListOptions{
			ListOptions: gl.ListOptions{Page: 1, PerPage: 100},
		}
		var allowedProjects []*gl.Project
		for {
			page, resp, err := c.api.JobTokenScope.GetProjectJobTokenInboundAllowList(p.ID, projectOpts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing job token inbound allowlist projects for project %d: %w", p.ID, err)
			}
			allowedProjects = append(allowedProjects, page...)
			if resp.NextPage == 0 {
				break
			}
			projectOpts.Page = resp.NextPage
		}

		groupOpts := &gl.GetJobTokenAllowlistGroupsOptions{
			ListOptions: gl.ListOptions{Page: 1, PerPage: 100},
		}
		var allowedGroups []*gl.Group
		for {
			page, resp, err := c.api.JobTokenScope.GetJobTokenAllowlistGroups(p.ID, groupOpts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing job token allowlist groups for project %d: %w", p.ID, err)
			}
			allowedGroups = append(allowedGroups, page...)
			if resp.NextPage == 0 {
				break
			}
			groupOpts.Page = resp.NextPage
		}

		// GitLab always includes the project itself in the inbound allowlist
		// implicitly; filter it out so we don't emit redundant self-references.
		filteredProjects := make([]*gl.Project, 0, len(allowedProjects))
		for _, ap := range allowedProjects {
			if ap.ID == p.ID {
				continue
			}
			filteredProjects = append(filteredProjects, ap)
		}

		result[p.ID] = &JobTokenScope{
			InboundEnabled:  settings.InboundEnabled,
			AllowedProjects: filteredProjects,
			AllowedGroups:   allowedGroups,
		}
	}

	return result, nil
}
