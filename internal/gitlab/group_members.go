package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

// GroupMembers maps group IDs to their direct members.
type GroupMembers = map[int64][]*gl.GroupMember

func (c *Client) ListGroupMembers(ctx context.Context, groups []*gl.Group) (GroupMembers, error) {
	result := make(GroupMembers, len(groups))

	for _, g := range groups {
		if g == nil {
			continue
		}
		slog.Debug("fetching group members", "group", g.FullPath)
		opts := &gl.ListGroupMembersOptions{
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		}
		members, err := paginate(&opts.ListOptions, "group members", g.FullPath, func() ([]*gl.GroupMember, *gl.Response, error) {
			return c.api.Groups.ListGroupMembers(g.ID, opts, gl.WithContext(ctx))
		})
		if err != nil {
			return nil, fmt.Errorf("listing members for group %d: %w", g.ID, err)
		}
		if len(members) > 0 {
			result[g.ID] = members
		}
	}

	return result, nil
}
