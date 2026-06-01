package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// ProtectedBranches maps project IDs to their protected branches.
type ProtectedBranches = map[int64][]*gl.ProtectedBranch

func (c *Client) ListProtectedBranches(ctx context.Context, projects []*gl.Project) (ProtectedBranches, error) {
	result := make(ProtectedBranches, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching protected branches", "project", p.PathWithNamespace)
		opts := &gl.ListProtectedBranchesOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		var branches []*gl.ProtectedBranch
		for {
			page, resp, err := c.api.ProtectedBranches.ListProtectedBranches(p.ID, opts, gl.WithContext(ctx))
			if err != nil {
				if skipInaccessible(err, "protected branches", p.PathWithNamespace) {
					break
				}
				return nil, fmt.Errorf("listing protected branches for project %d: %w", p.ID, err)
			}
			branches = append(branches, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		if len(branches) > 0 {
			result[p.ID] = branches
		}
	}
	return result, nil
}
