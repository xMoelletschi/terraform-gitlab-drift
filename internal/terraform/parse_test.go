package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseExistingResourcesFindsAllBlocks(t *testing.T) {
	dir := t.TempDir()

	content := `resource "gitlab_group" "my_group" {
  name = "My Group"
  path = "my-group"
}

resource "gitlab_project" "my_group_my_project" {
  name = "My Project"
  path = "my-project"
}
`
	if err := os.WriteFile(filepath.Join(dir, "my_group.tf"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseExistingResources(dir)
	if err != nil {
		t.Fatalf("ParseExistingResources error: %v", err)
	}

	want := map[string]bool{
		"gitlab_group.my_group":              true,
		"gitlab_project.my_group_my_project": true,
	}

	if len(got) != len(want) {
		t.Fatalf("got %d resources, want %d", len(got), len(want))
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing resource %q", k)
		}
	}
}

func TestParseExistingResourcesMultipleFiles(t *testing.T) {
	dir := t.TempDir()

	file1 := `resource "gitlab_group" "alpha" {
  name = "alpha"
  path = "alpha"
}
`
	file2 := `resource "gitlab_group_membership" "alpha" {
  for_each = var.gitlab_group_membership["mygroup/alpha"]
  group_id = gitlab_group.alpha.id
}
`
	if err := os.WriteFile(filepath.Join(dir, "alpha.tf"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "group_membership.tf"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseExistingResources(dir)
	if err != nil {
		t.Fatalf("ParseExistingResources error: %v", err)
	}

	if !got["gitlab_group.alpha"] {
		t.Error("missing gitlab_group.alpha")
	}
	if !got["gitlab_group_membership.alpha"] {
		t.Error("missing gitlab_group_membership.alpha")
	}
}

func TestParseExistingResourcesEmptyDir(t *testing.T) {
	dir := t.TempDir()

	got, err := ParseExistingResources(dir)
	if err != nil {
		t.Fatalf("ParseExistingResources error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}

func TestParseExistingResourcesIgnoresNonResourceBlocks(t *testing.T) {
	dir := t.TempDir()

	content := `variable "gitlab_group_membership" {
  description = "test"
  default     = {}
}

data "gitlab_user" "main" {
  for_each = local.users_by_groups
  username = each.key
}

resource "gitlab_group" "test" {
  name = "test"
  path = "test"
}
`
	if err := os.WriteFile(filepath.Join(dir, "test.tf"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseExistingResources(dir)
	if err != nil {
		t.Fatalf("ParseExistingResources error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(got))
	}
	if !got["gitlab_group.test"] {
		t.Error("missing gitlab_group.test")
	}
}
