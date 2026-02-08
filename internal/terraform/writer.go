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

	// Write all groups into a single file.
	if err := writeFile(filepath.Join(dir, "gitlab_groups.tf"), func(w io.Writer) error {
		return WriteGroups(resources.Groups, w, groupRefs)
	}); err != nil {
		errs = append(errs, fmt.Errorf("gitlab_groups.tf: %w", err))
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

	// Write one file per namespace.
	for ns, projects := range byNamespace {
		// Strip main group prefix from namespace for filename
		trimmedNs := strings.TrimPrefix(ns, mainGroup+"/")
		if trimmedNs == mainGroup {
			// If ns equals mainGroup exactly, keep it as-is
			trimmedNs = ns
		}

		filename := normalizeToTerraformName(trimmedNs) + ".tf"
		if ns == "" {
			filename = "gitlab_projects.tf"
		}
		if err := writeFile(filepath.Join(dir, filename), func(w io.Writer) error {
			return WriteProjects(projects, w, groupRefs)
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
