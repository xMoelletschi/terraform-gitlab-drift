package terraform

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func normalizeToTerraformName(path string) string {
	normalized := strings.ToLower(path)
	normalized = strings.ReplaceAll(normalized, "/", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	return normalized
}

func WriteAll(resources *gitlab.Resources, dir string, mainGroup string) error {
	var errs []error

	groupRefs := buildGroupRefMap(resources.Groups)

	groupsByPath := make(map[string]*gl.Group)
	for _, g := range resources.Groups {
		if g != nil && g.FullPath != "" {
			groupsByPath[g.FullPath] = g
		}
	}

	// Group projects by namespace full path.
	byNamespace := make(map[string][]*gl.Project)
	for _, p := range resources.Projects {
		ns := ""
		if p.Namespace != nil {
			ns = p.Namespace.FullPath
		}
		byNamespace[ns] = append(byNamespace[ns], p)
	}

	// Collect all unique namespaces
	allNamespaces := make(map[string]bool)
	for ns := range byNamespace {
		if ns != "" {
			allNamespaces[ns] = true
		}
	}
	for _, g := range resources.Groups {
		if g != nil && g.FullPath != "" {
			allNamespaces[g.FullPath] = true
		}
	}

	// Write group_membership.tf with variable only
	if err := writeFile(filepath.Join(dir, "group_membership.tf"), func(w io.Writer) error {
		return WriteGroupMembershipVariable(resources.Groups, resources.GroupMembers, w)
	}); err != nil {
		errs = append(errs, fmt.Errorf("group_membership.tf: %w", err))
	}

	// Write project_membership.tf with variable + helpers only
	if err := writeFile(filepath.Join(dir, "project_membership.tf"), func(w io.Writer) error {
		if err := WriteProjectMembershipVariable(resources.Projects, w); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
		return WriteProjectMembershipHelpers(w)
	}); err != nil {
		errs = append(errs, fmt.Errorf("project_membership.tf: %w", err))
	}

	// Write one file per namespace: group → group membership resource → projects → project share group resources
	for ns := range allNamespaces {
		trimmedNs := strings.TrimPrefix(ns, mainGroup+"/")
		if trimmedNs == mainGroup {
			trimmedNs = ns
		}

		filename := normalizeToTerraformName(trimmedNs) + ".tf"
		if err := writeFile(filepath.Join(dir, filename), func(w io.Writer) error {
			written := false

			if group, ok := groupsByPath[ns]; ok {
				if err := WriteGroups([]*gl.Group{group}, w, groupRefs); err != nil {
					return err
				}
				if err := WriteGroupMembershipResource(group, w); err != nil {
					return err
				}
				written = true
			}

			if projects := byNamespace[ns]; len(projects) > 0 {
				for i, p := range projects {
					if written || i > 0 {
						if _, err := w.Write([]byte("\n")); err != nil {
							return err
						}
					}
					if err := WriteProjects([]*gl.Project{p}, w, groupRefs); err != nil {
						return err
					}
					if err := WriteProjectShareGroupResource(p, w); err != nil {
						return err
					}
				}
			}

			return nil
		}); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filename, err))
		}
	}

	return errors.Join(errs...)
}

func writeFile(path string, writeFn func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close() //nolint:errcheck
	return writeFn(f)
}
