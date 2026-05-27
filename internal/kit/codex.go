package kit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CodexInstallOptions struct {
	KitRoot   string
	SkillsDir string
	Mode      string // "copy" or "link"
	DryRun    bool
	Verbose   bool
}

func CodexSkillsDir() string {
	if dir := os.Getenv("NTECH_TEAM_KIT_CODEX_SKILLS_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return codexSkillsDirForHome(home)
}

func codexSkillsDirForHome(home string) string {
	return filepath.Join(home, ".agents", "skills")
}

func PerformCodexInstall(opts CodexInstallOptions) error {
	if opts.KitRoot == "" {
		return fmt.Errorf("kit root is required")
	}
	if opts.SkillsDir == "" {
		return fmt.Errorf("Codex skills dir is required")
	}
	if opts.Mode == "" {
		opts.Mode = "copy"
	}
	if opts.Mode != "copy" && opts.Mode != "link" {
		return fmt.Errorf("invalid mode %q (expected copy or link)", opts.Mode)
	}

	manifestPath := codexManifestPath(opts.SkillsDir)
	existingManifest, _ := readManifest(manifestPath, filepath.Dir(opts.SkillsDir))
	existingManifest = filterOwnedCodexManifestEntries(opts.SkillsDir, existingManifest)

	var manifestEntries []manifestEntry
	installFile := func(src, dest string) error {
		if opts.DryRun {
			if opts.Verbose {
				fmt.Printf("[dry-run] %s %s -> %s\n", opts.Mode, src, dest)
			}
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("failed to create parent for %s: %w", dest, err)
		}
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("source file missing from kit root (%s): %s", opts.KitRoot, src)
		}
		if opts.Mode == "link" {
			_ = os.Remove(dest)
			if err := os.Symlink(src, dest); err != nil {
				return fmt.Errorf("failed to symlink %s -> %s: %w", src, dest, err)
			}
		} else if err := copyFile(src, dest); err != nil {
			return fmt.Errorf("failed to copy %s -> %s: %w", src, dest, err)
		}
		manifestEntries = append(manifestEntries, manifestEntry{Component: ComponentSkills, Path: dest})
		return nil
	}

	for _, skill := range skills {
		destDir := filepath.Join(opts.SkillsDir, skill)
		skillSrc := filepath.Join(opts.KitRoot, "skills", skill, "SKILL.md")
		if err := installFile(
			skillSrc,
			filepath.Join(destDir, "SKILL.md"),
		); err != nil {
			return err
		}
		for _, asset := range skillExtras[skill] {
			if err := installFile(
				filepath.Join(opts.KitRoot, "skills", skill, asset),
				filepath.Join(destDir, asset),
			); err != nil {
				return err
			}
		}
		if err := installGeneratedCodexMetadata(skill, skillSrc, filepath.Join(destDir, "agents", "openai.yaml"), opts.DryRun, opts.Verbose); err != nil {
			return err
		}
		manifestEntries = append(manifestEntries, manifestEntry{Component: ComponentSkills, Path: filepath.Join(destDir, "agents", "openai.yaml")})
	}

	if !opts.DryRun && len(manifestEntries) > 0 {
		merged := mergeManifestEntries(existingManifest, manifestEntries, ComponentSet{ComponentSkills: true})
		if err := writeManifest(manifestPath, merged); err != nil {
			return fmt.Errorf("failed to write Codex manifest: %w", err)
		}
	}

	if !opts.DryRun {
		printBanner()
		fmt.Printf("  Codex skills install complete (%s mode)\n", opts.Mode)
		fmt.Printf("  skills:     %d\n", len(skills))
		fmt.Printf("  location:   %s\n", opts.SkillsDir)
	}
	return nil
}

func PerformCodexUninstall(skillsDir string) error {
	manifest := codexManifestPath(skillsDir)
	entries, err := readManifest(manifest, filepath.Dir(skillsDir))
	if err != nil {
		log.Printf("[ntech-team-kit] no Codex manifest found at %s — nothing to uninstall", manifest)
		return nil
	}

	removed := 0
	ownedPaths := codexOwnedManifestPaths(skillsDir)
	for _, entry := range entries {
		if !isOwnedManifestEntry(ownedPaths, entry) {
			log.Printf("[ntech-team-kit] skipping unsafe Codex manifest entry: %s", entry.Path)
			continue
		}
		if _, err := os.Stat(entry.Path); err == nil {
			if err := os.Remove(entry.Path); err == nil {
				removed++
			}
		}
	}

	cleanupCodexSkillDirs(skillsDir)
	_ = os.Remove(manifest)
	log.Printf("[ntech-team-kit] uninstalled %d Codex skill files", removed)
	return nil
}

func PrintCodexStatus(skillsDir string) error {
	manifest := codexManifestPath(skillsDir)

	entries, err := readManifest(manifest, filepath.Dir(skillsDir))
	if err != nil {
		log.Printf("[ntech-team-kit] Codex not installed (no manifest at %s)", manifest)
		return nil
	}
	entries = filterOwnedCodexManifestEntries(skillsDir, entries)

	log.Printf("[ntech-team-kit] Codex: %d files tracked in manifest", len(entries))

	broken := 0
	for _, entry := range entries {
		if _, err := os.Stat(entry.Path); err != nil {
			log.Printf("[ntech-team-kit]   MISSING: %s", entry.Path)
			broken++
		}
	}

	if broken == 0 {
		log.Printf("[ntech-team-kit] Codex: all files present")
	} else {
		log.Printf("[ntech-team-kit] Codex: %d files missing — consider reinstalling", broken)
	}
	return nil
}

func codexManifestPath(skillsDir string) string {
	return filepath.Join(filepath.Dir(skillsDir), ".ntech-team-kit-codex-manifest")
}

func codexOwnedManifestPaths(skillsDir string) map[string]string {
	paths := map[string]string{}
	add := func(path string) {
		abs, err := absClean(path)
		if err == nil {
			paths[abs] = ComponentSkills
		}
	}
	for _, skill := range skills {
		destDir := filepath.Join(skillsDir, skill)
		add(filepath.Join(destDir, "SKILL.md"))
		for _, asset := range skillExtras[skill] {
			add(filepath.Join(destDir, asset))
		}
		add(filepath.Join(destDir, "agents", "openai.yaml"))
	}
	return paths
}

func filterOwnedCodexManifestEntries(skillsDir string, entries []manifestEntry) []manifestEntry {
	ownedPaths := codexOwnedManifestPaths(skillsDir)
	filtered := make([]manifestEntry, 0, len(entries))
	for _, entry := range entries {
		if isOwnedManifestEntry(ownedPaths, entry) {
			filtered = append(filtered, entry)
			continue
		}
		log.Printf("[ntech-team-kit] skipping unsafe Codex manifest entry: %s", entry.Path)
	}
	return filtered
}

func cleanupCodexSkillDirs(skillsDir string) {
	for _, skill := range skills {
		removeIfEmpty(filepath.Join(skillsDir, skill, "agents"))
		removeIfEmpty(filepath.Join(skillsDir, skill))
	}
	removeIfEmpty(skillsDir)
	removeIfEmpty(filepath.Dir(skillsDir))
}

func installGeneratedCodexMetadata(skill string, skillSrc string, dest string, dryRun bool, verbose bool) error {
	description := codexSkillDescription(skillSrc)
	if description == "" {
		description = "ntech-team-kit workflow for " + skill + "."
	}
	data := "interface:\n" +
		"  display_name: " + strconv.Quote(displayNameForSkill(skill)) + "\n" +
		"  short_description: " + strconv.Quote(description) + "\n" +
		"  brand_color: \"#3B82F6\"\n"

	if dryRun {
		if verbose {
			fmt.Printf("[dry-run] write generated Codex metadata -> %s\n", dest)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create parent for %s: %w", dest, err)
	}
	if err := writeFileAtomic(dest, []byte(data), 0o644); err != nil {
		return fmt.Errorf("failed to write generated Codex metadata %s: %w", dest, err)
	}
	return nil
}

func writeFileAtomic(dest string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(dest)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func codexSkillDescription(skillSrc string) string {
	data, err := os.ReadFile(skillSrc)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "description:")), "\"'")
		}
		if line == "---" {
			continue
		}
	}
	return ""
}

func displayNameForSkill(skill string) string {
	parts := strings.Split(skill, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func TargetIncludesCodex(target string) bool {
	return target == "codex" || target == "both" || target == "auto"
}

func TargetIncludesOpenCode(target string) bool {
	return target == "" || target == "opencode" || target == "both" || target == "auto"
}

func NormalizeInstallTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "opencode", nil
	}
	switch target {
	case "opencode", "codex", "both", "auto":
		return target, nil
	default:
		return "", fmt.Errorf("unknown target %q (expected opencode, codex, both, or auto)", target)
	}
}

func ResolveInstallTarget(target string) (string, error) {
	target, err := NormalizeInstallTarget(target)
	if err != nil {
		return "", err
	}
	if target != "auto" {
		return target, nil
	}

	openCode := checkOpenCode().Passed
	codex := checkCodexCLI().Passed || checkCodexGUI().Passed
	switch {
	case openCode && codex:
		return "both", nil
	case codex:
		return "codex", nil
	case openCode:
		return "opencode", nil
	default:
		return "", fmt.Errorf("could not auto-detect OpenCode or Codex")
	}
}
