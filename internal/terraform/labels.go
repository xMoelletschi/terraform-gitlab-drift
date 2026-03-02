package terraform

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func WriteGroupLabelVariable(groups []*gl.Group, groupLabels gitlab.GroupLabels, w io.Writer) error {
	var b strings.Builder
	b.WriteString("variable \"gitlab_group_label\" {\n")
	b.WriteString("  description = \"Labels for gitlab groups.\"\n")
	b.WriteString("  default = {\n")

	for _, g := range groups {
		if g == nil {
			continue
		}
		labels := groupLabels[g.ID]
		if len(labels) == 0 {
			fmt.Fprintf(&b, "    \"%s\" = {}\n", g.FullPath)
		} else {
			fmt.Fprintf(&b, "    \"%s\" = {\n", g.FullPath)
			for _, l := range labels {
				fmt.Fprintf(&b, "      \"%s\" = {\n", l.Name)
				fmt.Fprintf(&b, "        color       = \"%s\"\n", l.Color)
				fmt.Fprintf(&b, "        description = \"%s\"\n", l.Description)
				b.WriteString("      }\n")
			}
			b.WriteString("    }\n")
		}
	}

	b.WriteString("  }\n")
	b.WriteString("}\n")

	_, err := w.Write(hclwrite.Format([]byte(b.String())))
	return err
}

func WriteGroupLabelResource(group *gl.Group, w io.Writer) error {
	name := normalizeToTerraformName(group.Path)
	_, err := fmt.Fprintf(w, `resource "gitlab_group_label" "%s" {
  for_each    = var.gitlab_group_label["%s"]
  group       = gitlab_group.%s.id
  name        = each.key
  color       = each.value.color
  description = each.value.description
}
`, name, group.FullPath, name)
	return err
}

func WriteProjectLabelVariable(projects []*gl.Project, projectLabels gitlab.ProjectLabels, w io.Writer) error {
	var b strings.Builder
	b.WriteString("variable \"gitlab_project_label\" {\n")
	b.WriteString("  description = \"Labels for gitlab projects.\"\n")
	b.WriteString("  default = {\n")

	for _, p := range projects {
		if p == nil {
			continue
		}
		path := projectFullPath(p)
		labels := projectLabels[p.ID]
		if len(labels) == 0 {
			fmt.Fprintf(&b, "    \"%s\" = {}\n", path)
		} else {
			fmt.Fprintf(&b, "    \"%s\" = {\n", path)
			for _, l := range labels {
				fmt.Fprintf(&b, "      \"%s\" = {\n", l.Name)
				fmt.Fprintf(&b, "        color       = \"%s\"\n", l.Color)
				fmt.Fprintf(&b, "        description = \"%s\"\n", l.Description)
				b.WriteString("      }\n")
			}
			b.WriteString("    }\n")
		}
	}

	b.WriteString("  }\n")
	b.WriteString("}\n")

	_, err := w.Write(hclwrite.Format([]byte(b.String())))
	return err
}

func WriteProjectLabelResource(project *gl.Project, w io.Writer) error {
	name := projectResourceName(project)
	path := projectFullPath(project)
	_, err := fmt.Fprintf(w, `resource "gitlab_project_label" "%s" {
  for_each    = var.gitlab_project_label["%s"]
  project     = gitlab_project.%s.id
  name        = each.key
  color       = each.value.color
  description = each.value.description
}
`, name, path, name)
	return err
}
