package terraform

import (
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	gl "gitlab.com/gitlab-org/api/client-go"
)

// WriteUsers writes GitLab users as Terraform HCL resources.
func WriteUsers(users []*gl.User, w io.Writer) error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for _, u := range users {
		// TODO: Implement similar to WriteGroups
		// - Resource type: "gitlab_user"
		// - Resource name: normalizeToTerraformName(u.Username)
		// - Required: name, username, email
		// - Optional: is_admin, can_create_group, ...
		// - Check gl.User struct: go doc gitlab.com/gitlab-org/api/client-go.User
		// - Terraform docs: https://registry.terraform.io/providers/gitlabhq/gitlab/latest/docs/resources/user
		_ = u
		_ = rootBody
	}

	_, err := w.Write(f.Bytes())
	return err
}
