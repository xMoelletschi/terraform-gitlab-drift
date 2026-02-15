package terraform

import (
	"fmt"
	"io"
	"strings"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func WriteProjectMembershipVariable(projects []*gl.Project, w io.Writer) error {
	var b strings.Builder
	b.WriteString("variable \"gitlab_project_membership\" {\n")
	b.WriteString("  description = \"Share projects with groups.\"\n")
	b.WriteString("  default = {\n")

	for _, p := range projects {
		if p == nil {
			continue
		}
		path := projectFullPath(p)
		if len(p.SharedWithGroups) == 0 {
			fmt.Fprintf(&b, "    \"%s\" = {}\n", path)
		} else {
			fmt.Fprintf(&b, "    \"%s\" = {\n", path)
			for _, sg := range p.SharedWithGroups {
				fmt.Fprintf(&b, "      \"%s\" = \"%s\"\n", sg.GroupFullPath, accessLevelIntToString(sg.GroupAccessLevel))
			}
			b.WriteString("    }\n")
		}
	}

	b.WriteString("  }\n")
	b.WriteString("}\n")

	_, err := w.Write([]byte(b.String()))
	return err
}

func WriteProjectMembershipHelpers(w io.Writer) error {
	_, err := fmt.Fprint(w, `locals {
  groups_by_projects = toset(distinct(flatten([
    for key, project in var.gitlab_project_membership : [
      for group, permission in project : [
        group
      ]
    ]
  ])))
}

data "gitlab_group" "by_projects" {
  for_each  = local.groups_by_projects
  full_path = each.key
}
`)
	return err
}

func WriteProjectShareGroupResource(project *gl.Project, w io.Writer) error {
	name := normalizeToTerraformName(project.Path)
	if project.Namespace != nil && project.Namespace.FullPath != "" {
		parts := strings.Split(project.Namespace.FullPath, "/")
		parentGroup := parts[len(parts)-1]
		name = normalizeToTerraformName(parentGroup + "_" + project.Path)
	}
	path := projectFullPath(project)
	_, err := fmt.Fprintf(w, `resource "gitlab_project_share_group" "%s" {
  for_each     = var.gitlab_project_membership["%s"]
  project      = gitlab_project.%s.id
  group_id     = data.gitlab_group.by_projects[each.key].id
  group_access = each.value
}
`, name, path, name)
	return err
}

func projectFullPath(p *gl.Project) string {
	if p.PathWithNamespace != "" {
		return p.PathWithNamespace
	}
	if p.Namespace != nil && p.Namespace.FullPath != "" {
		return p.Namespace.FullPath + "/" + p.Path
	}
	return p.Path
}
