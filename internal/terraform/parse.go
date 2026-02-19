package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// ParseExistingResources reads all .tf files in dir and returns a set of
// "type.name" strings for every resource block found.
func ParseExistingResources(dir string) (map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("listing tf files: %w", err)
	}

	resources := make(map[string]bool)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		f, diags := hclwrite.ParseConfig(data, path, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return nil, fmt.Errorf("parsing %s: %s", path, diags.Error())
		}
		for _, block := range f.Body().Blocks() {
			if block.Type() == "resource" {
				labels := block.Labels()
				if len(labels) == 2 {
					resources[labels[0]+"."+labels[1]] = true
				}
			}
		}
	}

	return resources, nil
}
