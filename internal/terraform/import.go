package terraform

import (
	"fmt"
	"io"
	"strings"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go"
)

// ImportCommand represents a single terraform import command.
type ImportCommand struct {
	Address string
	ID      string
}

// GenerateImportCommands returns import commands for resources that exist in
// the API response but not in the existing terraform files.
func GenerateImportCommands(resources *gitlab.Resources, existingResources map[string]bool, mainGroup string, skipSet skip.Set) []ImportCommand {
	var cmds []ImportCommand

	for _, g := range resources.Groups {
		if g == nil {
			continue
		}
		name := normalizeToTerraformName(g.Path)
		key := "gitlab_group." + name
		if !existingResources[key] {
			cmds = append(cmds, ImportCommand{
				Address: key,
				ID:      fmt.Sprintf("%d", g.ID),
			})
		}
	}

	for _, p := range resources.Projects {
		if p == nil {
			continue
		}
		name := projectResourceName(p)
		key := "gitlab_project." + name
		if !existingResources[key] {
			cmds = append(cmds, ImportCommand{
				Address: key,
				ID:      fmt.Sprintf("%d", p.ID),
			})
		}
	}

	if !skipSet.Has("memberships") {
		for _, g := range resources.Groups {
			if g == nil {
				continue
			}
			name := normalizeToTerraformName(g.Path)
			key := "gitlab_group_membership." + name
			if !existingResources[key] {
				for _, m := range resources.GroupMembers[g.ID] {
					cmds = append(cmds, ImportCommand{
						Address: fmt.Sprintf("gitlab_group_membership.%s[\"%s\"]", name, m.Username),
						ID:      fmt.Sprintf("%d:%d", g.ID, m.ID),
					})
				}
			}
		}
	}

	if !skipSet.Has("memberships") {
		for _, p := range resources.Projects {
			if p == nil {
				continue
			}
			name := projectResourceName(p)
			key := "gitlab_project_share_group." + name
			if !existingResources[key] {
				for _, sg := range p.SharedWithGroups {
					cmds = append(cmds, ImportCommand{
						Address: fmt.Sprintf("gitlab_project_share_group.%s[\"%s\"]", name, sg.GroupFullPath),
						ID:      fmt.Sprintf("%d:%d", p.ID, sg.GroupID),
					})
				}
			}
		}
	}

	return cmds
}

// PrintImportCommands writes terraform import commands to w.
func PrintImportCommands(w io.Writer, cmds []ImportCommand) {
	for _, cmd := range cmds {
		fmt.Fprintf(w, "terraform import '%s' '%s'\n", cmd.Address, cmd.ID)
	}
}

// projectResourceName computes the terraform resource name for a project,
// matching the logic used in WriteProjects and WriteProjectShareGroupResource.
func projectResourceName(p *gl.Project) string {
	name := normalizeToTerraformName(p.Path)
	if p.Namespace != nil && p.Namespace.FullPath != "" {
		parts := strings.Split(p.Namespace.FullPath, "/")
		parentGroup := parts[len(parts)-1]
		name = normalizeToTerraformName(parentGroup + "_" + p.Path)
	}
	return name
}
