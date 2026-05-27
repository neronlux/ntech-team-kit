package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexSkillsDir_DefaultsToUserAgentsSkills(t *testing.T) {
	home := t.TempDir()
	got := codexSkillsDirForHome(home)
	want := filepath.Join(home, ".agents", "skills")
	if got != want {
		t.Fatalf("codexSkillsDirForHome() = %q, want %q", got, want)
	}
}

func TestCodexAgentsDir_DefaultsToUserCodexAgents(t *testing.T) {
	home := t.TempDir()
	got := codexAgentsDirForHome(home)
	want := filepath.Join(home, ".codex", "agents")
	if got != want {
		t.Fatalf("codexAgentsDirForHome() = %q, want %q", got, want)
	}
}

func TestPerformCodexInstall_InstallsSkillsAndExtras(t *testing.T) {
	root := createFakeKitRoot(t)
	skillsDir := t.TempDir()

	if err := PerformCodexInstall(CodexInstallOptions{
		KitRoot:   root,
		SkillsDir: skillsDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("PerformCodexInstall failed: %v", err)
	}

	for _, skill := range skills {
		if _, err := os.Stat(filepath.Join(skillsDir, skill, "SKILL.md")); err != nil {
			t.Fatalf("skill %s not installed for Codex: %v", skill, err)
		}
	}
	for _, asset := range skillExtras["pr-review-canvas"] {
		if _, err := os.Stat(filepath.Join(skillsDir, "pr-review-canvas", asset)); err != nil {
			t.Fatalf("pr-review-canvas extra asset %s not installed for Codex: %v", asset, err)
		}
	}
	if _, err := os.Stat(filepath.Join(skillsDir, "review-and-ship", "agents", "openai.yaml")); err != nil {
		t.Fatalf("Codex UI metadata not installed: %v", err)
	}
}

func TestPerformCodexAgentInstall_GeneratesTomlAgents(t *testing.T) {
	root := createFakeKitRoot(t)
	agentsDir := t.TempDir()
	source := filepath.Join(root, "agents", "ci-watcher.md")
	if err := os.WriteFile(source, []byte("---\ndescription: Watch CI checks.\nmode: subagent\npermission:\n  edit: deny\n---\n\n# CI watcher\n\nUse gh checks.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PerformCodexAgentInstall(CodexAgentInstallOptions{
		KitRoot:   root,
		SkillsDir: t.TempDir(),
		AgentsDir: agentsDir,
	}); err != nil {
		t.Fatalf("PerformCodexAgentInstall failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(agentsDir, "ci-watcher.toml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		`name = "ci-watcher"`,
		`description = "Watch CI checks."`,
		`sandbox_mode = "read-only"`,
		`developer_instructions = """`,
		"Use gh checks.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated Codex agent missing %q:\n%s", want, text)
		}
	}
}

func TestPerformCodexAgentInstall_CreatesManifestParentForAgentsOnly(t *testing.T) {
	root := createFakeKitRoot(t)
	base := t.TempDir()
	skillsDir := filepath.Join(base, ".agents", "skills")
	agentsDir := filepath.Join(base, ".codex", "agents")

	if err := PerformCodexAgentInstall(CodexAgentInstallOptions{
		KitRoot:   root,
		SkillsDir: skillsDir,
		AgentsDir: agentsDir,
	}); err != nil {
		t.Fatalf("PerformCodexAgentInstall failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(agentsDir, "ci-watcher.toml")); err != nil {
		t.Fatalf("Codex agent should be installed: %v", err)
	}
	if _, err := os.Stat(codexManifestPath(skillsDir)); err != nil {
		t.Fatalf("Codex manifest should be written for agents-only install: %v", err)
	}
}

func TestPerformCodexInstall_RewritesCompatibilityForCodexCopy(t *testing.T) {
	root := createFakeKitRoot(t)
	skillsDir := t.TempDir()
	src := filepath.Join(root, "skills", "review-and-ship", "SKILL.md")
	if err := os.WriteFile(src, []byte("---\nname: review-and-ship\ndescription: Review work.\ncompatibility: opencode\n---\n\n# Review\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PerformCodexInstall(CodexInstallOptions{
		KitRoot:   root,
		SkillsDir: skillsDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("PerformCodexInstall failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skillsDir, "review-and-ship", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "compatibility: opencode") {
		t.Fatalf("Codex copy should not remain OpenCode-only:\n%s", string(data))
	}
	if !strings.Contains(string(data), "compatibility: codex") {
		t.Fatalf("Codex copy should be marked Codex-compatible:\n%s", string(data))
	}
}

func TestPerformCodexUninstall_RemovesOnlyOwnedSkillPaths(t *testing.T) {
	root := createFakeKitRoot(t)
	skillsDir := t.TempDir()
	agentsDir := t.TempDir()
	outsideFile := filepath.Join(t.TempDir(), "keep.txt")
	if err := os.WriteFile(outsideFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PerformCodexInstall(CodexInstallOptions{
		KitRoot:   root,
		SkillsDir: skillsDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("PerformCodexInstall failed: %v", err)
	}

	manifest := codexManifestPath(skillsDir)
	f, err := os.OpenFile(manifest, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(ComponentSkills + "\t" + outsideFile + "\n"); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := PerformCodexUninstallSelected(skillsDir, agentsDir, ComponentSet{ComponentSkills: true}); err != nil {
		t.Fatalf("PerformCodexUninstall failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillsDir, skills[0], "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("Codex skill should be removed, got err=%v", err)
	}
	if data, err := os.ReadFile(outsideFile); err != nil {
		t.Fatalf("outside file should not be removed: %v", err)
	} else if string(data) != "keep" {
		t.Fatalf("outside file modified: %q", string(data))
	}
	if _, err := os.Stat(manifest); !os.IsNotExist(err) {
		t.Fatalf("Codex manifest should be removed, got err=%v", err)
	}
}

func TestPrintCodexStatus_NoManifest(t *testing.T) {
	skillsDir := t.TempDir()
	agentsDir := t.TempDir()
	if err := PrintCodexStatusWithDirs(skillsDir, agentsDir); err != nil {
		t.Fatalf("PrintCodexStatus with no manifest should not error: %v", err)
	}
}

func TestRefreshCodexSkillsView_UsesMacDeepLink(t *testing.T) {
	var ranName string
	var ranArgs []string
	err := refreshCodexSkillsView("darwin", func(name string) (string, error) {
		if name == "open" {
			return "/usr/bin/open", nil
		}
		return "", os.ErrNotExist
	}, func(name string, args ...string) error {
		ranName = name
		ranArgs = append([]string{}, args...)
		return nil
	})
	if err != nil {
		t.Fatalf("refreshCodexSkillsView failed: %v", err)
	}
	if ranName != "open" || len(ranArgs) != 1 || ranArgs[0] != "codex://skills" {
		t.Fatalf("ran %q %v, want open codex://skills", ranName, ranArgs)
	}
}

func TestRefreshCodexSkillsView_UsesLinuxDeepLink(t *testing.T) {
	var ranName string
	var ranArgs []string
	err := refreshCodexSkillsView("linux", func(name string) (string, error) {
		if name == "xdg-open" {
			return "/usr/bin/xdg-open", nil
		}
		return "", os.ErrNotExist
	}, func(name string, args ...string) error {
		ranName = name
		ranArgs = append([]string{}, args...)
		return nil
	})
	if err != nil {
		t.Fatalf("refreshCodexSkillsView failed: %v", err)
	}
	if ranName != "xdg-open" || len(ranArgs) != 1 || ranArgs[0] != "codex://skills" {
		t.Fatalf("ran %q %v, want xdg-open codex://skills", ranName, ranArgs)
	}
}

func TestRefreshCodexSkillsView_ReportsUnsupportedPlatform(t *testing.T) {
	err := refreshCodexSkillsView("windows", func(string) (string, error) {
		return "", os.ErrNotExist
	}, func(string, ...string) error {
		t.Fatal("runner should not be called")
		return nil
	})
	if err == nil {
		t.Fatal("expected unsupported platform error")
	}
}

func TestPerformCodexInstall_DoesNotFollowExistingMetadataSymlink(t *testing.T) {
	root := createFakeKitRoot(t)
	skillsDir := t.TempDir()
	outsideFile := filepath.Join(t.TempDir(), "outside.yaml")
	if err := os.WriteFile(outsideFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	metadata := filepath.Join(skillsDir, "review-and-ship", "agents", "openai.yaml")
	if err := os.MkdirAll(filepath.Dir(metadata), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, metadata); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	if err := PerformCodexInstall(CodexInstallOptions{
		KitRoot:   root,
		SkillsDir: skillsDir,
		Mode:      "copy",
	}); err != nil {
		t.Fatalf("PerformCodexInstall failed: %v", err)
	}

	if data, err := os.ReadFile(outsideFile); err != nil {
		t.Fatalf("outside file should remain readable: %v", err)
	} else if string(data) != "keep" {
		t.Fatalf("metadata write followed symlink and modified outside file: %q", string(data))
	}

	if info, err := os.Lstat(metadata); err != nil {
		t.Fatalf("metadata should exist after install: %v", err)
	} else if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("metadata should replace existing symlink with a regular file")
	}
}
