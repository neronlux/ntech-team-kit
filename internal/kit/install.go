package kit

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// These lists define all kit assets. Add new skills/agents/commands/rules here.
var (
	skills = []string{
		"check-compiler-errors",
		"control-cli",
		"control-ui",
		"deslop",
		"fix-ci",
		"fix-merge-conflicts",
		"get-pr-comments",
		"loop-on-ci",
		"make-pr-easy-to-review",
		"new-branch-and-pr",
		"pr-review-canvas",
		"review-and-ship",
		"run-smoke-tests",
		"thermo-nuclear-code-quality-review",
		"verify-this",
		"weekly-review",
		"what-did-i-get-done",
		"workflow-from-chats",
	}

	agents = []string{
		"ci-watcher",
		"thermo-nuclear-code-quality-review",
	}

	commands = skills[:]

	rules = []string{
		"no-inline-imports",
		"typescript-exhaustive-switch",
	}

	pluginDep        = "@opencode-ai/plugin"
	pluginDepVersion = "^1.14.0"
)

var skillExtras = map[string][]string{
	"pr-review-canvas": {"renderer.js", "styles.css", "template.html"},
}

const (
	ComponentSkills   = "skills"
	ComponentAgents   = "agents"
	ComponentCommands = "commands"
	ComponentRules    = "rules"
	ComponentPlugin   = "plugin"
	ComponentConfig   = "config"
)

type ComponentSet map[string]bool

func FullComponentSet() ComponentSet {
	return ComponentSet{
		ComponentSkills:   true,
		ComponentAgents:   true,
		ComponentCommands: true,
		ComponentRules:    true,
		ComponentPlugin:   true,
		ComponentConfig:   true,
	}
}

func LiteComponentSet() ComponentSet {
	return ComponentSet{
		ComponentSkills:   true,
		ComponentCommands: true,
		ComponentRules:    true,
		ComponentConfig:   true,
	}
}

func ValidComponent(name string) bool {
	switch name {
	case ComponentSkills, ComponentAgents, ComponentCommands, ComponentRules, ComponentPlugin, ComponentConfig:
		return true
	default:
		return false
	}
}

// IncludesOrAll reports whether the component is in the set.
// An empty set means "all components included", which is the desired default
// for uninstall and update flows where no filter means "everything".
func (c ComponentSet) Includes(name string) bool {
	return len(c) == 0 || c[name]
}

func (c ComponentSet) Names() []string {
	ordered := []string{ComponentSkills, ComponentAgents, ComponentCommands, ComponentRules, ComponentPlugin, ComponentConfig}
	var names []string
	for _, name := range ordered {
		if c.Includes(name) {
			names = append(names, name)
		}
	}
	return names
}

type manifestEntry struct {
	Component string
	Path      string
}

type InstallOptions struct {
	KitRoot    string
	ConfigDir  string // usually ~/.config/opencode
	Mode       string // "copy" or "link"
	Components ComponentSet
	DryRun     bool
	Verbose    bool
}

// PerformInstall is the robust, pure-Go implementation of installation.
// It replaces the fragile shell logic for the common `ntech-team-kit install` and `update` paths.
func PerformInstall(opts InstallOptions) error {
	if opts.KitRoot == "" {
		return fmt.Errorf("kit root is required")
	}
	if opts.ConfigDir == "" {
		return fmt.Errorf("config dir is required")
	}
	if opts.Mode == "" {
		opts.Mode = "copy"
	}
	if opts.Mode != "copy" && opts.Mode != "link" {
		return fmt.Errorf("invalid mode %q (expected copy or link)", opts.Mode)
	}
	if len(opts.Components) == 0 {
		opts.Components = FullComponentSet()
	}

	manifestPath := filepath.Join(opts.ConfigDir, ".ntech-team-kit-manifest")
	var manifestEntries []manifestEntry
	existingManifest, _ := readManifest(manifestPath, opts.ConfigDir)

	// Create top-level directories (pure Go, no shell brace expansion issues)
	dirs := []string{"skills", "agents", "commands", "rules", "plugins"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(opts.ConfigDir, d), 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", d, err)
		}
	}

	// Helper to install a single file
	installFile := func(component, src, dest string) error {
		if opts.DryRun {
			if opts.Verbose {
				fmt.Printf("[dry-run] %s %s -> %s\n", opts.Mode, src, dest)
			}
			return nil
		}

		// Always ensure parent exists (this was the root cause of the cp failures)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("failed to create parent for %s: %w", dest, err)
		}

		// Give a clear error if the source file is missing from the kit package
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("source file missing from kit root (%s): %s", opts.KitRoot, src)
		}

		if opts.Mode == "link" {
			// Remove existing link/target first
			_ = os.Remove(dest)
			if err := os.Symlink(src, dest); err != nil {
				return fmt.Errorf("failed to symlink %s -> %s: %w", src, dest, err)
			}
		} else {
			// Copy
			if err := copyFile(src, dest); err != nil {
				return fmt.Errorf("failed to copy %s -> %s: %w", src, dest, err)
			}
		}

		manifestEntries = append(manifestEntries, manifestEntry{Component: component, Path: dest})
		return nil
	}

	// Install skills (with any extra assets)
	if opts.Components.Includes(ComponentSkills) {
		for _, skill := range skills {
			destDir := filepath.Join(opts.ConfigDir, "skills", skill)
			if err := installFile(
				ComponentSkills,
				filepath.Join(opts.KitRoot, "skills", skill, "SKILL.md"),
				filepath.Join(destDir, "SKILL.md"),
			); err != nil {
				return err
			}

			for _, asset := range skillExtras[skill] {
				if err := installFile(
					ComponentSkills,
					filepath.Join(opts.KitRoot, "skills", skill, asset),
					filepath.Join(destDir, asset),
				); err != nil {
					return err
				}
			}
		}
	}

	// Install agents
	if opts.Components.Includes(ComponentAgents) {
		for _, agent := range agents {
			if err := installFile(
				ComponentAgents,
				filepath.Join(opts.KitRoot, "agents", agent+".md"),
				filepath.Join(opts.ConfigDir, "agents", agent+".md"),
			); err != nil {
				return err
			}
		}
	}

	// Install commands
	if opts.Components.Includes(ComponentCommands) {
		for _, cmd := range commands {
			if err := installFile(
				ComponentCommands,
				filepath.Join(opts.KitRoot, "commands", cmd+".md"),
				filepath.Join(opts.ConfigDir, "commands", cmd+".md"),
			); err != nil {
				return err
			}
		}
	}

	// Install rules
	if opts.Components.Includes(ComponentRules) {
		for _, rule := range rules {
			if err := installFile(
				ComponentRules,
				filepath.Join(opts.KitRoot, "rules", rule+".md"),
				filepath.Join(opts.ConfigDir, "rules", rule+".md"),
			); err != nil {
				return err
			}
		}
	}

	// Seed default config for first-time users so installed rules and the plugin
	// are actually loaded. Do not overwrite an existing OpenCode config.
	if opts.Components.Includes(ComponentConfig) && !hasOpenCodeConfig(opts.ConfigDir) {
		if err := installFile(
			ComponentConfig,
			filepath.Join(opts.KitRoot, "opencode.jsonc"),
			filepath.Join(opts.ConfigDir, "opencode.jsonc"),
		); err != nil {
			return err
		}
	}

	if opts.Components.Includes(ComponentPlugin) {
		// Plugin (ci-watcher)
		pluginSrc := filepath.Join(opts.KitRoot, "plugins", "ci-watcher.ts")
		pluginDest := filepath.Join(opts.ConfigDir, "plugins", "ci-watcher.ts")

		// Remove old package.json to avoid conflicts (same as shell)
		_ = os.Remove(filepath.Join(opts.ConfigDir, "plugins", "package.json"))

		if err := installFile(ComponentPlugin, pluginSrc, pluginDest); err != nil {
			return err
		}

		// Ensure @opencode-ai/plugin dependency
		if err := ensurePluginDependency(opts.ConfigDir, opts.DryRun); err != nil {
			return err
		}
	}

	// Write manifest atomically (collect, then write temp + rename)
	if !opts.DryRun && len(manifestEntries) > 0 {
		merged := mergeManifestEntries(existingManifest, manifestEntries, opts.Components)
		if err := writeManifest(manifestPath, merged); err != nil {
			return fmt.Errorf("failed to write manifest: %w", err)
		}
	}

	if !opts.DryRun {
		printBanner()
		fmt.Printf("  install complete (%s mode)\n", opts.Mode)
		fmt.Printf("  components: %s\n", strings.Join(opts.Components.Names(), ", "))
		if opts.Components.Includes(ComponentSkills) {
			fmt.Printf("  skills:     %d\n", len(skills))
		}
		if opts.Components.Includes(ComponentAgents) {
			fmt.Printf("  agents:     %d\n", len(agents))
		}
		if opts.Components.Includes(ComponentCommands) {
			fmt.Printf("  commands:   %d\n", len(commands))
		}
		if opts.Components.Includes(ComponentRules) {
			fmt.Printf("  rules:      %d\n", len(rules))
		}
		if opts.Components.Includes(ComponentPlugin) {
			fmt.Printf("  plugins:    1 (ci-watcher)\n")
		}
		fmt.Println("\n  To enable background CI watching, set:")
		fmt.Println("    export OPENCODE_NTECH_CI_WATCH=1")
	}

	return nil
}

func hasOpenCodeConfig(ocDir string) bool {
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		if _, err := os.Stat(filepath.Join(ocDir, name)); err == nil {
			return true
		}
	}
	return false
}

func mergeManifestEntries(existing, installed []manifestEntry, replaced ComponentSet) []manifestEntry {
	merged := make([]manifestEntry, 0, len(existing)+len(installed))
	for _, entry := range existing {
		if !replaced.Includes(entry.Component) {
			merged = append(merged, entry)
		}
	}
	merged = append(merged, installed...)
	return merged
}

func copyFile(src, dst string) error {
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", dst, err)
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("source is a directory, expected file: %s", src)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Write to a temp file in the destination directory, then rename it into
	// place. This avoids following an existing destination symlink. That matters
	// after Homebrew upgrades because old --link installs can leave symlinks into
	// removed Cellar versions; os.Create(dst) would follow the broken link and
	// fail with "no such file or directory".
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(dst)+".tmp-*")
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

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(srcInfo.Mode().Perm()); err != nil {
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

	if err := os.Rename(tmpName, dst); err != nil {
		return err
	}
	cleanup = false
	return nil
}

type packageJSON struct {
	Dependencies map[string]string `json:"dependencies"`
}

func ensurePluginDependency(ocDir string, dryRun bool) error {
	pkgPath := filepath.Join(ocDir, "package.json")

	if dryRun {
		return nil
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		pkg := packageJSON{
			Dependencies: map[string]string{
				pluginDep: pluginDepVersion,
			},
		}
		b, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal package.json: %w", err)
		}
		return os.WriteFile(pkgPath, b, 0o644)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("invalid package.json: %w", err)
	}

	if pkg.Dependencies == nil {
		pkg.Dependencies = map[string]string{}
	}

	if _, exists := pkg.Dependencies[pluginDep]; exists {
		return nil
	}

	pkg.Dependencies[pluginDep] = pluginDepVersion

	b, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}
	_ = os.WriteFile(pkgPath+".ntech-team-kit.bak", data, 0o644)
	return os.WriteFile(pkgPath, b, 0o644)
}

// writeManifest writes the manifest atomically using temp + rename.
func writeManifest(path string, entries []manifestEntry) error {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, entry.Component+"\t"+entry.Path)
	}
	data := strings.Join(lines, "\n") + "\n"

	tmp, err := os.CreateTemp(filepath.Dir(path), ".ntech-team-kit-manifest.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.WriteString(data); err != nil {
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

	return os.Rename(tmpName, path)
}

func printBanner() {
	fmt.Print(`
     _____         _     
 _ _|_   _|__  ___| |__  
| '_ \| |/ _ \/ __| '_ \ 
| | | | |  __/ (__| | | |
|_| |_|_|\___|\___|_| |_|
  ntech-team-kit
`)
}

// PrintStatus replicates the original do_status behavior using the manifest.
func PrintStatus(ocDir string) error {
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")

	entries, err := readManifest(manifest, ocDir)
	if err != nil {
		log.Printf("[ntech-team-kit] not installed (no manifest at %s)", manifest)
		return nil
	}

	log.Printf("[ntech-team-kit] %d files tracked in manifest", len(entries))

	broken := 0
	for _, entry := range entries {
		if _, err := os.Stat(entry.Path); err != nil {
			log.Printf("[ntech-team-kit]   MISSING: %s", entry.Path)
			broken++
		}
	}

	if broken == 0 {
		log.Printf("[ntech-team-kit] all files present")
	} else {
		log.Printf("[ntech-team-kit] %d files missing — consider reinstalling", broken)
	}
	return nil
}

// PerformUninstall removes all files listed in the manifest and cleans up empty directories.
func PerformUninstall(ocDir string) error {
	return PerformUninstallSelected(ocDir, nil)
}

func PerformUninstallSelected(ocDir string, components ComponentSet) error {
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")

	entries, err := readManifest(manifest, ocDir)
	if err != nil {
		log.Printf("[ntech-team-kit] no manifest found at %s — nothing to uninstall", manifest)
		return nil
	}

	removed := 0
	var remaining []manifestEntry
	for _, entry := range entries {
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

	cleanupEmptyDirs(ocDir)

	if len(remaining) == 0 {
		_ = os.Remove(manifest)
	} else if err := writeManifest(manifest, remaining); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	log.Printf("[ntech-team-kit] uninstalled %d files", removed)
	return nil
}

func readManifest(path, ocDir string) ([]manifestEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []manifestEntry
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		entry := parseManifestLine(line, ocDir)
		entries = append(entries, entry)
	}
	return entries, nil
}

func parseManifestLine(line, ocDir string) manifestEntry {
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) == 2 && ValidComponent(parts[0]) {
		return manifestEntry{Component: parts[0], Path: parts[1]}
	}
	return manifestEntry{Component: inferComponent(line, ocDir), Path: line}
}

func inferComponent(path, ocDir string) string {
	rel, err := filepath.Rel(ocDir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = path
	}
	first := strings.Split(filepath.ToSlash(rel), "/")[0]
	switch first {
	case "skills":
		return ComponentSkills
	case "agents":
		return ComponentAgents
	case "commands":
		return ComponentCommands
	case "rules":
		return ComponentRules
	case "plugins":
		return ComponentPlugin
	case "opencode.json", "opencode.jsonc":
		return ComponentConfig
	default:
		return ""
	}
}

func cleanupEmptyDirs(ocDir string) {
	for _, skill := range skills {
		removeIfEmpty(filepath.Join(ocDir, "skills", skill))
	}
	for _, dir := range []string{"skills", "agents", "commands", "rules", "plugins"} {
		removeIfEmpty(filepath.Join(ocDir, dir))
	}
}

func removeIfEmpty(dir string) {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		entries, _ := os.ReadDir(dir)
		if len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
}
