package terraform

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
	"github.com/xMoelletschi/terraform-gitlab-drift/internal/skip"
	gl "gitlab.com/gitlab-org/api/client-go"
)

func normalizeToTerraformName(s string) string {
	return normalizeName(s)
}

func WriteAll(resources *gitlab.Resources, dir string, mainGroup string, skipSet skip.Set) error {
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

	// Write group_membership.tf with variable + user data source
	if !skipSet.Has("memberships") {
		if err := writeFile(filepath.Join(dir, "group_membership.tf"), func(w io.Writer) error {
			if err := WriteGroupMembershipVariable(resources.Groups, resources.GroupMembers, w); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				return err
			}
			return WriteUserDataSource(w)
		}); err != nil {
			errs = append(errs, fmt.Errorf("group_membership.tf: %w", err))
		}
	}

	// Write project_membership.tf with variable + helpers only
	if !skipSet.Has("memberships") {
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
	}

	// Write group_labels.tf with variable
	if !skipSet.Has("labels") {
		if err := writeFile(filepath.Join(dir, "group_labels.tf"), func(w io.Writer) error {
			return WriteGroupLabelVariable(resources.Groups, resources.GroupLabels, w)
		}); err != nil {
			errs = append(errs, fmt.Errorf("group_labels.tf: %w", err))
		}
	}

	// Write project_labels.tf with variable
	if !skipSet.Has("labels") {
		if err := writeFile(filepath.Join(dir, "project_labels.tf"), func(w io.Writer) error {
			return WriteProjectLabelVariable(resources.Projects, resources.ProjectLabels, w)
		}); err != nil {
			errs = append(errs, fmt.Errorf("project_labels.tf: %w", err))
		}
	}

	// Write pipeline_schedules.tf with individual resource blocks
	if !skipSet.Has("schedules") {
		if err := writeFile(filepath.Join(dir, "pipeline_schedules.tf"), func(w io.Writer) error {
			first := true
			for _, p := range resources.Projects {
				if p == nil {
					continue
				}
				if scheds := resources.PipelineSchedules[p.ID]; len(scheds) > 0 {
					if !first {
						if _, err := w.Write([]byte("\n")); err != nil {
							return err
						}
					}
					if err := WritePipelineSchedules(p, scheds, w); err != nil {
						return err
					}
					first = false
				}
			}
			return nil
		}); err != nil {
			errs = append(errs, fmt.Errorf("pipeline_schedules.tf: %w", err))
		}
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
				if !skipSet.Has("memberships") {
					if err := WriteGroupMembershipResource(group, w); err != nil {
						return err
					}
				}
				if !skipSet.Has("labels") && len(resources.GroupLabels[group.ID]) > 0 {
					if err := WriteGroupLabelResource(group, w); err != nil {
						return err
					}
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
					if !skipSet.Has("memberships") {
						if err := WriteProjectShareGroupResource(p, w); err != nil {
							return err
						}
					}
					if !skipSet.Has("labels") && len(resources.ProjectLabels[p.ID]) > 0 {
						if err := WriteProjectLabelResource(p, w); err != nil {
							return err
						}
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
