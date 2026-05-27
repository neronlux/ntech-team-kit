package kit

import (
	"os"
	"path/filepath"
	"strings"
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

func TestCheckOpenCodeReportsDetectedPath(t *testing.T) {
	r := checkOpenCodeWith(func(name string) (string, error) {
		if name == "opencode" {
			return "/usr/local/bin/opencode", nil
		}
		return "", os.ErrNotExist
	})

	if !r.Passed {
		t.Fatalf("OpenCode check should pass: %s", r.Message)
	}
	if !strings.Contains(r.Message, "/usr/local/bin/opencode") {
		t.Fatalf("OpenCode check should report path, got %q", r.Message)
	}
}

func TestCheckCodexCLIDetectsPathBinary(t *testing.T) {
	r := checkCodexCLIWith(func(name string) (string, error) {
		if name == "codex" {
			return "/usr/local/bin/codex", nil
		}
		return "", os.ErrNotExist
	}, func(string) bool {
		return false
	}, func(string) string {
		return ""
	}, "/home/user", "linux")

	if !r.Passed {
		t.Fatalf("Codex CLI check should pass: %s", r.Message)
	}
	if !strings.Contains(r.Message, "/usr/local/bin/codex") {
		t.Fatalf("Codex CLI check should report path, got %q", r.Message)
	}
}

func TestCheckCodexGUIDetectsMacApp(t *testing.T) {
	home := t.TempDir()
	app := filepath.Join(home, "Applications", "Codex.app")
	if err := os.MkdirAll(app, 0o755); err != nil {
		t.Fatal(err)
	}

	r := checkCodexGUIWith(pathExists, globPaths, home, "darwin")
	if !r.Passed {
		t.Fatalf("Codex GUI check should detect mac app: %s", r.Message)
	}
	if !strings.Contains(r.Message, app) {
		t.Fatalf("Codex GUI check should report app path, got %q", r.Message)
	}
}

func TestCheckCodexGUIDetectsLinuxDesktopEntry(t *testing.T) {
	home := t.TempDir()
	desktopEntry := filepath.Join(home, ".local", "share", "applications", "codex.desktop")
	if err := os.MkdirAll(filepath.Dir(desktopEntry), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(desktopEntry, []byte("[Desktop Entry]\nName=Codex\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := checkCodexGUIWith(pathExists, globPaths, home, "linux")
	if !r.Passed {
		t.Fatalf("Codex GUI check should detect linux desktop entry: %s", r.Message)
	}
	if !strings.Contains(r.Message, desktopEntry) {
		t.Fatalf("Codex GUI check should report desktop entry, got %q", r.Message)
	}
}
