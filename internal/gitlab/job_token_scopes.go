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

		// Mirror the provider's Read behavior so HCL matches state.
		//
		// The provider's getProjectCIJobScopes removes self with a `break`,
		// stripping a single self entry from the raw API response before
		// storing the rest in a Set (which dedups by ID). When the GitLab
		// API returns self more than once (a known data quirk reachable e.g.
		// by manual UI/API additions), one self instance survives the
		// `break` and ends up in state — and the API rejects DELETE on self
		// with 400 "Source project cannot be removed from the job token
		// scope", so emitting self-less HCL makes apply fail.
		//
		// We replicate this: count self occurrences in the raw response, then
		// dedup. If self appeared <= 1 times we strip it (matches state =
		// without self). If >= 2 times we keep it (matches state = with
		// self). We never POST self ourselves, so this never propagates the
		// duplicate to projects that don't already have it.
		selfCount := 0
		seen := make(map[int64]bool, len(allowedProjects))
		deduped := make([]*gl.Project, 0, len(allowedProjects))
		for _, ap := range allowedProjects {
			if ap.ID == p.ID {
				selfCount++
			}
			if seen[ap.ID] {
				continue
			}
			seen[ap.ID] = true
			deduped = append(deduped, ap)
		}

		allowed := deduped
		if selfCount <= 1 {
			filtered := make([]*gl.Project, 0, len(deduped))
			for _, ap := range deduped {
				if ap.ID == p.ID {
					continue
				}
				filtered = append(filtered, ap)
			}
			allowed = filtered
		}
		if len(allowed) == 0 {
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
