package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// ProtectedTags maps project IDs to their protected tags.
type ProtectedTags = map[int64][]*gl.ProtectedTag

func (c *Client) ListProtectedTags(ctx context.Context, projects []*gl.Project) (ProtectedTags, error) {
	result := make(ProtectedTags, len(projects))

	for _, p := range projects {
		if p == nil {
			continue
		}
		slog.Debug("fetching protected tags", "project", p.PathWithNamespace)
		opts := &gl.ListProtectedTagsOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		var tags []*gl.ProtectedTag
		for {
			page, resp, err := c.api.ProtectedTags.ListProtectedTags(p.ID, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("listing protected tags for project %d: %w", p.ID, err)
			}
			tags = append(tags, page...)
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		if len(tags) > 0 {
			result[p.ID] = tags
		}
	}
	return result, nil
}
