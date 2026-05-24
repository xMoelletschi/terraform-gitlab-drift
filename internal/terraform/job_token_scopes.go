package terraform

import (
	"fmt"
	"io"
	"sort"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func WriteJobTokenScopes(p *gl.Project, scope *gitlab.JobTokenScope, projectRefs projectRefMap, groupRefs groupRefMap, w io.Writer) error {
	if scope == nil {
		return nil
	}
	projName := projectResourceName(p)

	if _, err := fmt.Fprintf(w, "resource \"gitlab_project_job_token_scopes\" \"%s\" {\n", projName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  project = gitlab_project.%s.id\n", projName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  enabled = %t\n", scope.InboundEnabled); err != nil {
		return err
	}

	if len(scope.AllowedProjects) == 0 {
		if _, err := fmt.Fprintln(w, "  target_project_ids = []"); err != nil {
			return err
		}
	} else {
		projects := make([]*gl.Project, len(scope.AllowedProjects))
		copy(projects, scope.AllowedProjects)
		sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })

		if _, err := fmt.Fprintln(w, "  target_project_ids = ["); err != nil {
			return err
		}
		for _, tp := range projects {
			if name, ok := projectRefs[tp.ID]; ok && name != "" {
				if _, err := fmt.Fprintf(w, "    gitlab_project.%s.id,\n", name); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(w, "    %d, # unmanaged: %s\n", tp.ID, tp.PathWithNamespace); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "  ]"); err != nil {
			return err
		}
	}

	if len(scope.AllowedGroups) > 0 {
		groups := make([]*gl.Group, len(scope.AllowedGroups))
		copy(groups, scope.AllowedGroups)
		sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })

		if _, err := fmt.Fprintln(w, "  target_group_ids = ["); err != nil {
			return err
		}
		for _, tg := range groups {
			if name, ok := groupRefs[tg.ID]; ok && name != "" {
				if _, err := fmt.Fprintf(w, "    gitlab_group.%s.id,\n", name); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(w, "    %d, # unmanaged: %s\n", tg.ID, tg.FullPath); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "  ]"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, "}"); err != nil {
		return err
	}
	return nil
}
