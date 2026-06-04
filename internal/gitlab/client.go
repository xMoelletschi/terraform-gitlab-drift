package gitlab

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	api         *gl.Client
	group       string
	concurrency int
}

// defaultConcurrency bounds how many resource fetchers run at once when the
// caller does not specify a value.
const defaultConcurrency = 10

// ClientConfig tunes how the client talks to the GitLab API.
type ClientConfig struct {
	// RateLimit caps API requests per second via the client's proactive
	// limiter. 0 disables the cap.
	RateLimit float64
	// Concurrency bounds how many resource fetchers run at once. <= 0 uses
	// defaultConcurrency.
	Concurrency int
}

func resolveConcurrency(n int) int {
	if n <= 0 {
		return defaultConcurrency
	}
	return n
}

type Resources struct {
	Groups            []*gl.Group
	Projects          []*gl.Project
	GroupMembers      GroupMembers
	GroupLabels       GroupLabels
	ProjectLabels     ProjectLabels
	PipelineSchedules PipelineSchedules
	ProjectHooks      ProjectHooks
	GroupHooks        GroupHooks
	ProjectVariables  ProjectVariables
	GroupVariables    GroupVariables
	ProtectedBranches ProtectedBranches
	ProtectedTags     ProtectedTags
	JobTokenScopes    JobTokenScopes
}

func NewClientFromAPI(api *gl.Client, group string) *Client {
	return &Client{api: api, group: group, concurrency: defaultConcurrency}
}

func NewClient(token, baseURL, group string, cfg ClientConfig) (*Client, error) {
	client, err := gl.NewClient(token,
		gl.WithBaseURL(baseURL),
		gl.WithCustomLimiter(newRateLimiter(cfg.RateLimit)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating GitLab client: %w", err)
	}
	return &Client{api: client, group: group, concurrency: resolveConcurrency(cfg.Concurrency)}, nil
}

func (c *Client) FetchAll(ctx context.Context, skipSet skip.Set, cfg FilterConfig) (*Resources, error) {
	// Groups and projects are needed by every subresource fetcher, so fetch
	// them first — concurrently with each other.
	var (
		groups   []*gl.Group
		projects []*gl.Project
	)
	pre, preCtx := errgroup.WithContext(ctx)
	pre.Go(func() error {
		var err error
		if groups, err = c.ListGroups(preCtx, cfg); err != nil {
			return fmt.Errorf("listing groups: %w", err)
		}
		return nil
	})
	pre.Go(func() error {
		var err error
		if projects, err = c.ListProjects(preCtx, cfg); err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}
		return nil
	})
	if err := pre.Wait(); err != nil {
		return nil, err
	}
	slog.Info("fetched groups", "count", len(groups))
	slog.Info("fetched projects", "count", len(projects))

	// The subresource fetchers are independent and each writes its own result
	// variable, so they run concurrently, bounded by c.concurrency. The client's
	// rate limiter keeps the combined request rate under the API limit.
	var (
		groupMembers      GroupMembers
		groupLabels       GroupLabels
		projectLabels     ProjectLabels
		pipelineSchedules PipelineSchedules
		projectHooks      ProjectHooks
		groupHooks        GroupHooks
		projectVariables  ProjectVariables
		groupVariables    GroupVariables
		protectedBranches ProtectedBranches
		protectedTags     ProtectedTags
		jobTokenScopes    JobTokenScopes
	)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(resolveConcurrency(c.concurrency))

	if !skipSet.Has("memberships") {
		g.Go(func() error {
			var err error
			if groupMembers, err = c.ListGroupMembers(gctx, groups); err != nil {
				return fmt.Errorf("listing group members: %w", err)
			}
			slog.Info("fetched group members", "count", len(groupMembers))
			return nil
		})
	}
	if !skipSet.Has("labels") {
		g.Go(func() error {
			var err error
			if groupLabels, err = c.ListGroupLabels(gctx, groups); err != nil {
				return fmt.Errorf("listing group labels: %w", err)
			}
			slog.Info("fetched group labels", "count", len(groupLabels))
			return nil
		})
		g.Go(func() error {
			var err error
			if projectLabels, err = c.ListProjectLabels(gctx, projects); err != nil {
				return fmt.Errorf("listing project labels: %w", err)
			}
			slog.Info("fetched project labels", "count", len(projectLabels))
			return nil
		})
	}
	if !skipSet.Has("schedules") {
		g.Go(func() error {
			var err error
			if pipelineSchedules, err = c.ListPipelineSchedules(gctx, projects); err != nil {
				return fmt.Errorf("listing pipeline schedules: %w", err)
			}
			slog.Info("fetched pipeline schedules", "count", len(pipelineSchedules))
			return nil
		})
	}
	if !skipSet.Has("variables") {
		g.Go(func() error {
			var err error
			if groupVariables, err = c.ListGroupVariables(gctx, groups); err != nil {
				return fmt.Errorf("listing group variables: %w", err)
			}
			slog.Info("fetched group variables", "count", len(groupVariables))
			return nil
		})
		g.Go(func() error {
			var err error
			if projectVariables, err = c.ListProjectVariables(gctx, projects); err != nil {
				return fmt.Errorf("listing project variables: %w", err)
			}
			slog.Info("fetched project variables", "count", len(projectVariables))
			return nil
		})
	}
	if !skipSet.Has("branch_protection") {
		g.Go(func() error {
			var err error
			if protectedBranches, err = c.ListProtectedBranches(gctx, projects); err != nil {
				return fmt.Errorf("listing protected branches: %w", err)
			}
			slog.Info("fetched protected branches", "count", len(protectedBranches))
			return nil
		})
	}
	if !skipSet.Has("tag_protection") {
		g.Go(func() error {
			var err error
			if protectedTags, err = c.ListProtectedTags(gctx, projects); err != nil {
				return fmt.Errorf("listing protected tags: %w", err)
			}
			slog.Info("fetched protected tags", "count", len(protectedTags))
			return nil
		})
	}
	if !skipSet.Has("job_token_scopes") {
		g.Go(func() error {
			var err error
			if jobTokenScopes, err = c.ListJobTokenScopes(gctx, projects); err != nil {
				return fmt.Errorf("listing job token scopes: %w", err)
			}
			slog.Info("fetched job token scopes", "count", len(jobTokenScopes))
			return nil
		})
	}
	if !skipSet.Has("hooks") {
		g.Go(func() error {
			var err error
			if projectHooks, err = c.ListProjectHooks(gctx, projects); err != nil {
				return fmt.Errorf("listing project hooks: %w", err)
			}
			slog.Info("fetched project hooks", "count", len(projectHooks))
			return nil
		})
		g.Go(func() error {
			var err error
			if groupHooks, err = c.ListGroupHooks(gctx, groups); err != nil {
				return fmt.Errorf("listing group hooks: %w", err)
			}
			slog.Info("fetched group hooks", "count", len(groupHooks))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &Resources{
		Groups:            groups,
		Projects:          projects,
		GroupMembers:      groupMembers,
		GroupLabels:       groupLabels,
		ProjectLabels:     projectLabels,
		PipelineSchedules: pipelineSchedules,
		ProjectHooks:      projectHooks,
		GroupHooks:        groupHooks,
		ProjectVariables:  projectVariables,
		GroupVariables:    groupVariables,
		ProtectedBranches: protectedBranches,
		ProtectedTags:     protectedTags,
		JobTokenScopes:    jobTokenScopes,
	}, nil
}
