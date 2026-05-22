package kit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetKitRoot_EnvOverride(t *testing.T) {
	fake := "/tmp/ntech-test-kit-root"
	os.Setenv("NTECH_TEAM_KIT_ROOT", fake)
	defer os.Unsetenv("NTECH_TEAM_KIT_ROOT")

	got := GetKitRoot()
	if got != fake {
		t.Errorf("GetKitRoot = %q, want %q", got, fake)
	}
}

func TestHasKitLayout(t *testing.T) {
	dir := t.TempDir()

	if hasKitLayout(dir) {
		t.Error("empty dir should not pass hasKitLayout")
	}

	os.MkdirAll(filepath.Join(dir, "skills"), 0o755)
	os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.0"), 0o644)

	if !hasKitLayout(dir) {
		t.Error("dir with skills/ + VERSION should pass hasKitLayout")
	}
}

func TestValidateKitRoot_Empty(t *testing.T) {
	err := ValidateKitRoot("")
	if err == nil {
		t.Error("expected error for empty root")
	}
}

func TestValidateKitRoot_Nonexistent(t *testing.T) {
	err := ValidateKitRoot("/tmp/ntech-nonexistent-path-xyz")
	if err == nil {
		t.Error("expected error for nonexistent root")
	}
}

func TestValidateKitRoot_Valid(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "skills", "check-compiler-errors"), 0o755)
	os.WriteFile(filepath.Join(dir, "skills", "check-compiler-errors", "SKILL.md"), []byte("# test"), 0o644)
	os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.0"), 0o644)

	if err := ValidateKitRoot(dir); err != nil {
		t.Errorf("ValidateKitRoot on valid dir: %v", err)
	}
}

func TestSearchUpForKit(t *testing.T) {
	dir := t.TempDir()
	child := filepath.Join(dir, "sub1", "sub2")
	os.MkdirAll(child, 0o755)
	os.MkdirAll(filepath.Join(dir, "skills"), 0o755)
	os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.0"), 0o644)

	got := searchUpForKit(child)
	if got != dir {
		t.Errorf("searchUpForKit = %q, want %q", got, dir)
	}
}
