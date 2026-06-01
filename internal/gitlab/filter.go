package gitlab

import (
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// FilterConfig controls which groups and projects a scan includes, based on
// their lifecycle state. It is populated from CLI flags today; keeping it as a
// single struct means a future config file can populate the same fields without
// touching the call sites.
type FilterConfig struct {
	// ShowPendingDeletion includes groups and projects that are in the deletion
	// grace period (MarkedForDeletionOn). They are hidden by default.
	ShowPendingDeletion bool
	// HideArchived excludes archived projects. They are shown by default.
	// (Groups have no archived state in the GitLab API.)
	HideArchived bool
}

func filterGroups(groups []*gl.Group, cfg FilterConfig) []*gl.Group {
	filtered := make([]*gl.Group, 0, len(groups))
	for _, g := range groups {
		if !cfg.ShowPendingDeletion && g.MarkedForDeletionOn != nil {
			slog.Debug("skipping group pending deletion", "group", g.FullPath)
			continue
		}
		filtered = append(filtered, g)
	}
	return filtered
}

func filterProjects(projects []*gl.Project, cfg FilterConfig) []*gl.Project {
	filtered := make([]*gl.Project, 0, len(projects))
	for _, p := range projects {
		if !cfg.ShowPendingDeletion && p.MarkedForDeletionOn != nil {
			slog.Debug("skipping project pending deletion", "project", p.PathWithNamespace)
			continue
		}
		if cfg.HideArchived && p.Archived {
			slog.Debug("skipping archived project", "project", p.PathWithNamespace)
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}
