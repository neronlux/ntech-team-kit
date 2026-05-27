package kit

import (
	"os"
	"path/filepath"
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

func TestPerformCodexUninstall_RemovesOnlyOwnedSkillPaths(t *testing.T) {
	root := createFakeKitRoot(t)
	skillsDir := t.TempDir()
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

	if err := PerformCodexUninstall(skillsDir); err != nil {
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
	if err := PrintCodexStatus(skillsDir); err != nil {
		t.Fatalf("PrintCodexStatus with no manifest should not error: %v", err)
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
