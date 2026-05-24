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

		// GitLab can return duplicate entries in the allowlist; the provider's
		// state stores a deduped Set, so we dedup here too to match.
		seen := make(map[int64]bool, len(allowedProjects))
		deduped := make([]*gl.Project, 0, len(allowedProjects))
		for _, ap := range allowedProjects {
			if seen[ap.ID] {
				continue
			}
			seen[ap.ID] = true
			deduped = append(deduped, ap)
		}

		// The provider's Read function stores target_project_ids = [] when the
		// API allowlist contains only the source project itself; we mirror
		// that so HCL matches state. When the allowlist has additional
		// entries we must keep self in the list — the GitLab API rejects
		// DELETE requests for the source project ("Source project cannot be
		// removed from the job token scope"), so emitting self-less HCL
		// causes apply to fail when the provider tries to reconcile.
		allowed := deduped
		if len(deduped) == 1 && deduped[0].ID == p.ID {
			allowed = nil
		}

		result[p.ID] = &JobTokenScope{
			InboundEnabled:  settings.InboundEnabled,
			AllowedProjects: allowed,
			AllowedGroups:   allowedGroups,
		}
	}

	return result, nil
}
