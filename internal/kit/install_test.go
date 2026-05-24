package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createFakeKitRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{"skills", "agents", "commands", "rules", "plugins"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("0.0.0-test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "opencode.jsonc"), []byte("{\n  \"instructions\": [\"rules/*.md\"],\n  \"plugin\": [\"plugins/ci-watcher.ts\"]\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, skill := range skills {
		dir := filepath.Join(root, "skills", skill)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+skill), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, asset := range skillExtras[skill] {
			if err := os.WriteFile(filepath.Join(dir, asset), []byte("/* "+asset+" */"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, agent := range agents {
		if err := os.WriteFile(filepath.Join(root, "agents", agent+".md"), []byte("# "+agent), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for _, cmd := range commands {
		if err := os.WriteFile(filepath.Join(root, "commands", cmd+".md"), []byte("# "+cmd), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for _, rule := range rules {
		if err := os.WriteFile(filepath.Join(root, "rules", rule+".md"), []byte("# "+rule), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(root, "plugins", "ci-watcher.ts"), []byte("// plugin"), 0o644); err != nil {
		t.Fatal(err)
	}

	return root
}

func TestPerformInstall_Copy(t *testing.T) {
	root := createFakeKitRoot(t)
	configDir := t.TempDir()

	err := PerformInstall(InstallOptions{
		KitRoot:   root,
		ConfigDir: configDir,
		Mode:      "copy",
	})
	if err != nil {
		t.Fatalf("PerformInstall failed: %v", err)
	}

	for _, skill := range skills {
		p := filepath.Join(configDir, "skills", skill, "SKILL.md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("skill %s not installed: %v", skill, err)
		}
		for _, asset := range skillExtras[skill] {
			ap := filepath.Join(configDir, "skills", skill, asset)
			if _, err := os.Stat(ap); err != nil {
				t.Errorf("skill %s extra asset %s not installed: %v", skill, asset, err)
			}
		}
	}

	for _, agent := range agents {
		p := filepath.Join(configDir, "agents", agent+".md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("agent %s not installed: %v", agent, err)
		}
	}

	for _, cmd := range commands {
		p := filepath.Join(configDir, "commands", cmd+".md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("command %s not installed: %v", cmd, err)
		}
	}

	for _, rule := range rules {
		p := filepath.Join(configDir, "rules", rule+".md")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("rule %s not installed: %v", rule, err)
		}
	}

	if _, err := os.Stat(filepath.Join(configDir, "opencode.jsonc")); err != nil {
		t.Errorf("opencode.jsonc not installed for first-time config: %v", err)
	}

	manifest := filepath.Join(configDir, ".ntech-team-kit-manifest")
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatalf("manifest not found: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	expectedCount := len(skills) + len(skillExtras["pr-review-canvas"]) + len(agents) + len(commands) + len(rules) + 2
	if len(lines) != expectedCount {
		t.Errorf("manifest has %d lines, expected %d", len(lines), expectedCount)
	}
}

func TestPerformInstall_DoesNotOverwriteExistingConfig(t *testing.T) {
	root := createFakeKitRoot(t)
	configDir := t.TempDir()
	existing := filepath.Join(configDir, "opencode.jsonc")
	if err := os.WriteFile(existing, []byte("{\"instructions\": [\"custom.md\"]}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PerformInstall(InstallOptions{
		KitRoot:   root,
		ConfigDir: configDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("PerformInstall failed: %v", err)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{\"instructions\": [\"custom.md\"]}\n" {
		t.Fatalf("existing opencode.jsonc was overwritten: %s", string(data))
	}
}

func TestPerformInstall_Link(t *testing.T) {
	root := createFakeKitRoot(t)
	configDir := t.TempDir()

	err := PerformInstall(InstallOptions{
		KitRoot:   root,
		ConfigDir: configDir,
		Mode:      "link",
	})
	if err != nil {
		t.Fatalf("PerformInstall link mode failed: %v", err)
	}

	skillPath := filepath.Join(configDir, "skills", skills[0], "SKILL.md")
	info, err := os.Lstat(skillPath)
	if err != nil {
		t.Fatalf("stat installed skill: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink in link mode")
	}
}

func TestPerformInstall_DryRun(t *testing.T) {
	root := createFakeKitRoot(t)
	configDir := t.TempDir()

	err := PerformInstall(InstallOptions{
		KitRoot:   root,
		ConfigDir: configDir,
		Mode:      "copy",
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("PerformInstall dry run failed: %v", err)
	}

	manifest := filepath.Join(configDir, ".ntech-team-kit-manifest")
	if _, err := os.Stat(manifest); !os.IsNotExist(err) {
		t.Error("dry run should not create manifest file")
	}

	skillPath := filepath.Join(configDir, "skills", skills[0], "SKILL.md")
	if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
		t.Error("dry run should not copy skill files")
	}
}

func TestPerformInstall_MissingKitRoot(t *testing.T) {
	err := PerformInstall(InstallOptions{KitRoot: "", ConfigDir: "/tmp"})
	if err == nil {
		t.Error("expected error for empty kit root")
	}
}

func TestPerformInstall_MissingConfigDir(t *testing.T) {
	err := PerformInstall(InstallOptions{KitRoot: "/tmp", ConfigDir: ""})
	if err == nil {
		t.Error("expected error for empty config dir")
	}
}

func TestPerformUninstall(t *testing.T) {
	root := createFakeKitRoot(t)
	configDir := t.TempDir()

	if err := PerformInstall(InstallOptions{
		KitRoot:   root,
		ConfigDir: configDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	if err := PerformUninstall(configDir); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	manifest := filepath.Join(configDir, ".ntech-team-kit-manifest")
	if _, err := os.Stat(manifest); !os.IsNotExist(err) {
		t.Error("manifest should be removed after uninstall")
	}
}

func TestPerformUninstall_NoManifest(t *testing.T) {
	configDir := t.TempDir()
	err := PerformUninstall(configDir)
	if err != nil {
		t.Errorf("uninstall with no manifest should not error: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	content := "hello world\nline 2\n"
	src := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dstDir, "output.txt")
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestCopyFile_OverwritesSymlink(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	content := "new content\n"
	src := filepath.Join(srcDir, "file.txt")
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	brokenLink := filepath.Join(srcDir, "nonexistent")
	dst := filepath.Join(dstDir, "file.txt")
	if err := os.Symlink(brokenLink, dst); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile over broken symlink failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestPrintStatus_NoManifest(t *testing.T) {
	configDir := t.TempDir()
	err := PrintStatus(configDir)
	if err != nil {
		t.Errorf("PrintStatus with no manifest should not error: %v", err)
	}
}
