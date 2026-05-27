package kit

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type CodexInstallOptions struct {
	KitRoot   string
	SkillsDir string
	AgentsDir string
	Mode      string // "copy" or "link"
	DryRun    bool
	Verbose   bool
}

type CodexAgentInstallOptions struct {
	KitRoot   string
	SkillsDir string
	AgentsDir string
	DryRun    bool
	Verbose   bool
}

type codexTextReplacement struct {
	from string
	to   string
}

var codexSkillTextReplacements = map[string][]codexTextReplacement{
	"workflow-from-chats": {
		{
			from: "recent OpenCode sessions",
			to:   "recent Codex or OpenCode sessions",
		},
		{
			from: "relevant subagent transcripts. Use subagent content",
			to:   "relevant subagent/custom-agent transcripts. Use child-agent content",
		},
		{
			from: "relevant subagents",
			to:   "relevant subagents/custom agents",
		},
		{
			from: "subagent consensus",
			to:   "child-agent consensus",
		},
	},
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

func CodexAgentsDir() string {
	if dir := os.Getenv("NTECH_TEAM_KIT_CODEX_AGENTS_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return codexAgentsDirForHome(home)
}

func codexAgentsDirForHome(home string) string {
	return filepath.Join(home, ".codex", "agents")
}

func PerformCodexInstall(opts CodexInstallOptions) error {
	if opts.KitRoot == "" {
		return fmt.Errorf("kit root is required")
	}
	if opts.SkillsDir == "" {
		return fmt.Errorf("Codex skills dir is required")
	}
	if opts.AgentsDir == "" {
		opts.AgentsDir = CodexAgentsDir()
	}
	if opts.Mode == "" {
		opts.Mode = "copy"
	}
	if opts.Mode != "copy" && opts.Mode != "link" {
		return fmt.Errorf("invalid mode %q (expected copy or link)", opts.Mode)
	}

	manifestPath := codexManifestPath(opts.SkillsDir)
	existingManifest, _ := readManifest(manifestPath, filepath.Dir(opts.SkillsDir))
	existingManifest = filterOwnedCodexManifestEntries(opts.SkillsDir, opts.AgentsDir, existingManifest)

	var manifestEntries []manifestEntry
	installSkillFile := func(src, dest string) error {
		if opts.DryRun {
			if opts.Verbose {
				fmt.Printf("[dry-run] write Codex skill -> %s\n", dest)
			}
			return nil
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("source file missing from kit root (%s): %s", opts.KitRoot, src)
		}
		skill := filepath.Base(filepath.Dir(src))
		if err := writeFileAtomic(dest, codexSkillMarkdown(skill, data), 0o644); err != nil {
			return fmt.Errorf("failed to write Codex skill %s: %w", dest, err)
		}
		manifestEntries = append(manifestEntries, manifestEntry{Component: ComponentSkills, Path: dest})
		return nil
	}
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
		if err := installSkillFile(
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

func PerformCodexAgentInstall(opts CodexAgentInstallOptions) error {
	if opts.KitRoot == "" {
		return fmt.Errorf("kit root is required")
	}
	if opts.SkillsDir == "" {
		opts.SkillsDir = CodexSkillsDir()
	}
	if opts.AgentsDir == "" {
		opts.AgentsDir = CodexAgentsDir()
	}

	manifestPath := codexManifestPath(opts.SkillsDir)
	existingManifest, _ := readManifest(manifestPath, filepath.Dir(opts.SkillsDir))
	existingManifest = filterOwnedCodexManifestEntries(opts.SkillsDir, opts.AgentsDir, existingManifest)

	var manifestEntries []manifestEntry
	for _, agent := range agents {
		src := filepath.Join(opts.KitRoot, "agents", agent+".md")
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("source file missing from kit root (%s): %s", opts.KitRoot, src)
		}
		dest := filepath.Join(opts.AgentsDir, agent+".toml")
		if opts.DryRun {
			if opts.Verbose {
				fmt.Printf("[dry-run] write Codex agent -> %s\n", dest)
			}
		} else if err := writeFileAtomic(dest, codexAgentTOML(agent, data), 0o644); err != nil {
			return fmt.Errorf("failed to write Codex agent %s: %w", dest, err)
		}
		manifestEntries = append(manifestEntries, manifestEntry{Component: ComponentAgents, Path: dest})
	}

	if !opts.DryRun && len(manifestEntries) > 0 {
		merged := mergeManifestEntries(existingManifest, manifestEntries, ComponentSet{ComponentAgents: true})
		if err := writeManifest(manifestPath, merged); err != nil {
			return fmt.Errorf("failed to write Codex manifest: %w", err)
		}
	}

	if !opts.DryRun {
		printBanner()
		fmt.Printf("  Codex agents install complete\n")
		fmt.Printf("  agents:     %d\n", len(agents))
		fmt.Printf("  location:   %s\n", opts.AgentsDir)
	}
	return nil
}

func PerformCodexUninstall(skillsDir string) error {
	return PerformCodexUninstallSelected(skillsDir, CodexAgentsDir(), nil)
}

func PerformCodexUninstallSelected(skillsDir string, agentsDir string, components ComponentSet) error {
	manifest := codexManifestPath(skillsDir)
	entries, err := readManifest(manifest, filepath.Dir(skillsDir))
	if err != nil {
		log.Printf("[ntech-team-kit] no Codex manifest found at %s — nothing to uninstall", manifest)
		return nil
	}

	removed := 0
	var remaining []manifestEntry
	ownedPaths := codexOwnedManifestPaths(skillsDir, agentsDir)
	for _, entry := range entries {
		if !isOwnedManifestEntry(ownedPaths, entry) {
			log.Printf("[ntech-team-kit] skipping unsafe Codex manifest entry: %s", entry.Path)
			continue
		}
		if !components.Includes(entry.Component) {
			remaining = append(remaining, entry)
			continue
		}
		if _, err := os.Stat(entry.Path); err == nil {
			if err := os.Remove(entry.Path); err == nil {
				removed++
			}
		}
	}

	if components.Includes(ComponentSkills) {
		cleanupCodexSkillDirs(skillsDir)
	}
	if components.Includes(ComponentAgents) {
		cleanupCodexAgentDirs(agentsDir)
	}

	if len(remaining) == 0 {
		_ = os.Remove(manifest)
	} else if err := writeManifest(manifest, remaining); err != nil {
		return fmt.Errorf("failed to update Codex manifest: %w", err)
	}

	log.Printf("[ntech-team-kit] uninstalled %d Codex files", removed)
	return nil
}

func PrintCodexStatus(skillsDir string) error {
	return PrintCodexStatusWithDirs(skillsDir, CodexAgentsDir())
}

func PrintCodexStatusWithDirs(skillsDir string, agentsDir string) error {
	manifest := codexManifestPath(skillsDir)

	entries, err := readManifest(manifest, filepath.Dir(skillsDir))
	if err != nil {
		log.Printf("[ntech-team-kit] Codex not installed (no manifest at %s)", manifest)
		return nil
	}
	entries = filterOwnedCodexManifestEntries(skillsDir, agentsDir, entries)

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

func RefreshCodexSkillsView() error {
	return refreshCodexSkillsView(runtime.GOOS, exec.LookPath, runCodexRefreshCommand)
}

func refreshCodexSkillsView(goos string, lookPath func(string) (string, error), run func(string, ...string) error) error {
	switch goos {
	case "darwin":
		if _, err := lookPath("open"); err != nil {
			return fmt.Errorf("open command not found")
		}
		return run("open", "codex://skills")
	case "linux":
		if _, err := lookPath("xdg-open"); err != nil {
			return fmt.Errorf("xdg-open command not found")
		}
		return run("xdg-open", "codex://skills")
	default:
		return fmt.Errorf("Codex app refresh is not supported on %s", goos)
	}
}

func runCodexRefreshCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}

func codexManifestPath(skillsDir string) string {
	return filepath.Join(filepath.Dir(skillsDir), ".ntech-team-kit-codex-manifest")
}

func codexOwnedManifestPaths(skillsDir string, agentsDir string) map[string]string {
	paths := map[string]string{}
	add := func(component string, path string) {
		abs, err := absClean(path)
		if err == nil {
			paths[abs] = component
		}
	}
	for _, skill := range skills {
		destDir := filepath.Join(skillsDir, skill)
		add(ComponentSkills, filepath.Join(destDir, "SKILL.md"))
		for _, asset := range skillExtras[skill] {
			add(ComponentSkills, filepath.Join(destDir, asset))
		}
		add(ComponentSkills, filepath.Join(destDir, "agents", "openai.yaml"))
	}
	for _, agent := range agents {
		add(ComponentAgents, filepath.Join(agentsDir, agent+".toml"))
	}
	return paths
}

func filterOwnedCodexManifestEntries(skillsDir string, agentsDir string, entries []manifestEntry) []manifestEntry {
	ownedPaths := codexOwnedManifestPaths(skillsDir, agentsDir)
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

func cleanupCodexAgentDirs(agentsDir string) {
	removeIfEmpty(agentsDir)
}

func installGeneratedCodexMetadata(skill string, skillSrc string, dest string, dryRun bool, verbose bool) error {
	description := codexSkillDescription(skillSrc)
	if description == "" {
		description = "ntech-team-kit workflow for " + skill + "."
	}
	description = applyCodexSkillTextReplacements(skill, description)
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

func codexSkillMarkdown(skill string, data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	frontmatterClosed := false
	replaced := false

	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && line == "---" {
			frontmatterClosed = true
			if !replaced {
				lines = append(lines[:i], append([]string{"compatibility: codex"}, lines[i:]...)...)
			}
			break
		}
		if inFrontmatter && strings.HasPrefix(line, "compatibility:") {
			lines[i] = "compatibility: codex"
			replaced = true
		}
	}

	if !inFrontmatter || !frontmatterClosed {
		return data
	}
	return []byte(applyCodexSkillTextReplacements(skill, strings.Join(lines, "\n")))
}

func applyCodexSkillTextReplacements(skill string, text string) string {
	for _, replacement := range codexSkillTextReplacements[skill] {
		text = strings.ReplaceAll(text, replacement.from, replacement.to)
	}
	return text
}

func codexAgentTOML(name string, data []byte) []byte {
	frontmatter, body := splitMarkdownFrontmatter(data)
	description := frontmatter["description"]
	if description == "" {
		description = "ntech-team-kit " + name + " agent."
	}
	instructions := "This Codex custom agent was generated from ntech-team-kit agents/" + name + ".md.\n" +
		"Use Codex custom-agent behavior; ignore OpenCode-only markdown frontmatter, @-mentions, and Task tool wording in the source instructions.\n\n" +
		strings.TrimSpace(body) + "\n"

	toml := "name = " + strconv.Quote(name) + "\n" +
		"description = " + strconv.Quote(description) + "\n" +
		"sandbox_mode = \"read-only\"\n" +
		"developer_instructions = \"\"\"\n" +
		escapeTOMLMultilineBasicString(instructions) +
		"\"\"\"\n"
	return []byte(toml)
}

func splitMarkdownFrontmatter(data []byte) (map[string]string, string) {
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return map[string]string{}, text
	}
	parts := strings.SplitN(strings.TrimPrefix(text, "---\n"), "\n---\n", 2)
	if len(parts) != 2 {
		return map[string]string{}, text
	}

	frontmatter := map[string]string{}
	for _, line := range strings.Split(parts[0], "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		frontmatter[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return frontmatter, parts[1]
}

func escapeTOMLMultilineBasicString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"""`, `\"\"\"`)
	return value
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
