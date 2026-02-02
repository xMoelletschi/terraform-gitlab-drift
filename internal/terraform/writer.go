package terraform

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xMoelletschi/terraform-gitlab-drift/internal/gitlab"
)

func normalizeToTerraformName(path string) string {
	normalized := strings.ToLower(path)
	normalized = strings.ReplaceAll(normalized, "/", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	return normalized
}

func WriteAll(resources *gitlab.Resources, dir string) error {
	writers := []struct {
		filename string
		writeFn  func(io.Writer) error
	}{
		{"gitlab_groups.tf", func(w io.Writer) error { return WriteGroups(resources.Groups, w) }},
		{"gitlab_projects.tf", func(w io.Writer) error { return WriteProjects(resources.Projects, w) }},
		{"gitlab_users.tf", func(w io.Writer) error { return WriteUsers(resources.Users, w) }},
	}

	var errs []error
	for _, wr := range writers {
		if err := writeFile(filepath.Join(dir, wr.filename), wr.writeFn); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", wr.filename, err))
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
