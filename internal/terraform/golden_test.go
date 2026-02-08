package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

const updateGoldenEnv = "UPDATE_GOLDEN"

func compareGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)

	if os.Getenv(updateGoldenEnv) != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("writing golden file: %v", err)
		}
		t.Skipf("updated golden file %s", path)
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}

	if got != string(want) {
		t.Fatalf("golden file mismatch for %s\nGot:\n%s\nWant:\n%s", path, got, string(want))
	}
}
