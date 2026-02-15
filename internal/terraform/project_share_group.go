package terraform

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

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

	_, err := w.Write(hclwrite.Format([]byte(b.String())))
	return err
}

func WriteProjectMembershipHelpers(w io.Writer) error {
	src := []byte(`locals {
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
	_, err := w.Write(hclwrite.Format(src))
	return err
}

func WriteProjectShareGroupResource(project *gl.Project, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	name := normalizeToTerraformName(project.Path)
	if project.Namespace != nil && project.Namespace.FullPath != "" {
		parts := strings.Split(project.Namespace.FullPath, "/")
		parentGroup := parts[len(parts)-1]
		name = normalizeToTerraformName(parentGroup + "_" + project.Path)
	}
	path := projectFullPath(project)

	block := f.Body().AppendNewBlock("resource", []string{"gitlab_project_share_group", name})
	body := block.Body()

	body.SetAttributeTraversal("for_each", hcl.Traversal{
		hcl.TraverseRoot{Name: "var"},
		hcl.TraverseAttr{Name: "gitlab_project_membership"},
		hcl.TraverseIndex{Key: cty.StringVal(path)},
	})

	body.SetAttributeTraversal("project", hcl.Traversal{
		hcl.TraverseRoot{Name: "gitlab_project"},
		hcl.TraverseAttr{Name: name},
		hcl.TraverseAttr{Name: "id"},
	})

	body.SetAttributeRaw("group_id", tokensForIndexedTraversal(
		hcl.Traversal{
			hcl.TraverseRoot{Name: "data"},
			hcl.TraverseAttr{Name: "gitlab_group"},
			hcl.TraverseAttr{Name: "by_projects"},
		},
		hcl.Traversal{
			hcl.TraverseRoot{Name: "each"},
			hcl.TraverseAttr{Name: "key"},
		},
		"id",
	))

	body.SetAttributeTraversal("group_access", hcl.Traversal{
		hcl.TraverseRoot{Name: "each"},
		hcl.TraverseAttr{Name: "value"},
	})

	_, err := w.Write(f.Bytes())
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
