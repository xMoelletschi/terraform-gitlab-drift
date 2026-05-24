package terraform

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go/v2"
)

func WriteJobTokenScopes(p *gl.Project, scope *gitlab.JobTokenScope, projectRefs projectRefMap, groupRefs groupRefMap, w io.Writer) error {
	if scope == nil {
		return nil
	}
	projName := projectResourceName(p)

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "resource \"gitlab_project_job_token_scopes\" \"%s\" {\n", projName)
	fmt.Fprintf(&buf, "  project = gitlab_project.%s.id\n", projName)
	fmt.Fprintf(&buf, "  enabled = %t\n", scope.InboundEnabled)

	if len(scope.AllowedProjects) == 0 {
		fmt.Fprintln(&buf, "  target_project_ids = []")
	} else {
		projects := make([]*gl.Project, len(scope.AllowedProjects))
		copy(projects, scope.AllowedProjects)
		sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })

		fmt.Fprintln(&buf, "  target_project_ids = [")
		for _, tp := range projects {
			if name, ok := projectRefs[tp.ID]; ok && name != "" {
				fmt.Fprintf(&buf, "    gitlab_project.%s.id,\n", name)
				continue
			}
			fmt.Fprintf(&buf, "    %d, # unmanaged: %s\n", tp.ID, tp.PathWithNamespace)
		}
		fmt.Fprintln(&buf, "  ]")
	}

	if len(scope.AllowedGroups) > 0 {
		groups := make([]*gl.Group, len(scope.AllowedGroups))
		copy(groups, scope.AllowedGroups)
		sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })

		fmt.Fprintln(&buf, "  target_group_ids = [")
		for _, tg := range groups {
			if name, ok := groupRefs[tg.ID]; ok && name != "" {
				fmt.Fprintf(&buf, "    gitlab_group.%s.id,\n", name)
				continue
			}
			fmt.Fprintf(&buf, "    %d, # unmanaged: %s\n", tg.ID, tg.FullPath)
		}
		fmt.Fprintln(&buf, "  ]")
	}

	fmt.Fprintln(&buf, "}")

	// Run through hclwrite.Format so output matches `terraform fmt` exactly
	// (aligns the `=` columns) — otherwise every fmt run produces drift.
	_, err := w.Write(hclwrite.Format(buf.Bytes()))
	return err
}
