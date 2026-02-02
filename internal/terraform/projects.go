package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	gl "gitlab.com/gitlab-org/api/client-go"
)

// WriteProjects writes GitLab projects as Terraform HCL resources.
func WriteProjects(projects []*gl.Project, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for _, p := range projects {
		// TODO: Implement similar to WriteGroups
		// - Resource type: "gitlab_project"
		// - Resource name: normalizeToTerraformName(p.Path)
		// - Required: name, path
		// - Optional: namespace_id, description, visibility_level, ...
		// - Check gl.Project struct: go doc gitlab.com/gitlab-org/api/client-go.Project
		// - Terraform docs: https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/project
		_ = p
		_ = rootBody
	}

	_, err := w.Write(f.Bytes())
	return err
}
