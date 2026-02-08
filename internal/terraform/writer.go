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

	// Write one file per namespace with group at the top, followed by projects
	for ns := range allNamespaces {
		trimmedNs := strings.TrimPrefix(ns, mainGroup+"/")
		if trimmedNs == mainGroup {
			// If ns equals mainGroup exactly, keep it as-is
			trimmedNs = ns
		}

		filename := normalizeToTerraformName(trimmedNs) + ".tf"
		if err := writeFile(filepath.Join(dir, filename), func(w io.Writer) error {
			hasGroup := false
			if group, ok := groupsByPath[ns]; ok {
				if err := WriteGroups([]*gl.Group{group}, w, groupRefs); err != nil {
					return err
				}
				hasGroup = true
			}

			if projects := byNamespace[ns]; len(projects) > 0 {
				if hasGroup {
					if _, err := w.Write([]byte("\n")); err != nil {
						return err
					}
				}
				if err := WriteProjects(projects, w, groupRefs); err != nil {
					return err
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
	defer f.Close()
	return writeFn(f)
}
