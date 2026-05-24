package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDoctor_ValidKitRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "skills", "check-compiler-errors"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "VERSION"), []byte("1.0"), 0o644)
	os.WriteFile(filepath.Join(root, "skills", "check-compiler-errors", "SKILL.md"), []byte("# test"), 0o644)

	results := RunDoctor(root)

	foundKitRoot := false
	foundKitContents := false
	for _, r := range results {
		if r.Name == "Kit Root" {
			foundKitRoot = true
			if !r.Passed {
				t.Errorf("Kit Root check should pass with valid root: %s", r.Message)
			}
		}
		if r.Name == "Kit Contents" {
			foundKitContents = true
		}
	}
	if !foundKitRoot {
		t.Error("missing Kit Root check")
	}
	if !foundKitContents {
		t.Error("missing Kit Contents check")
	}
}

func TestCheckKitRoot_Empty(t *testing.T) {
	r := checkKitRoot("")
	if r.Passed {
		t.Error("empty root should not pass")
	}
}

func TestCheckKitRoot_Valid(t *testing.T) {
	dir := t.TempDir()
	r := checkKitRoot(dir)
	if !r.Passed {
		t.Error("existing dir should pass")
	}
}
