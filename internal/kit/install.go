package kit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// These lists must be kept in sync with install.sh
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

	commands = []string{
		"review-and-ship",
		"loop-on-ci",
		"verify-this",
		"run-smoke-tests",
		"fix-ci",
		"new-branch-and-pr",
		"make-pr-easy-to-review",
		"fix-merge-conflicts",
	}

	rules = []string{
		"no-inline-imports",
		"typescript-exhaustive-switch",
	}

	pluginDep = "@opencode-ai/plugin"
	pluginDepVersion = "^1.14.0"
)

type InstallOptions struct {
	KitRoot     string
	ConfigDir   string // usually ~/.config/opencode
	Mode        string // "copy" or "link"
	DryRun      bool
	Verbose     bool
}

// PerformInstall is the robust, pure-Go implementation of installation.
// It replaces the fragile shell logic for the common `ntech-team-kit install` and `update` paths.
func PerformInstall(opts InstallOptions) error {
	if opts.KitRoot == "" {
		return fmt.Errorf("kit root is required")
	}
	if opts.ConfigDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		opts.ConfigDir = filepath.Join(home, ".config", "opencode")
	}
	if opts.Mode == "" {
		opts.Mode = "copy"
	}

	manifestPath := filepath.Join(opts.ConfigDir, ".ntech-team-kit-manifest")

	// Create top-level directories (pure Go, no shell brace expansion issues)
	dirs := []string{"skills", "agents", "commands", "rules", "plugins"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(opts.ConfigDir, d), 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", d, err)
		}
	}

	// Truncate manifest
	if !opts.DryRun {
		if err := os.WriteFile(manifestPath, []byte{}, 0o644); err != nil {
			return fmt.Errorf("failed to initialize manifest: %w", err)
		}
	}

	// Helper to install a single file
	installFile := func(src, dest string) error {
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

		// Append to manifest
		f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := fmt.Fprintf(f, "%s\n", dest); err != nil {
			return err
		}
		return nil
	}

	// Install skills
	for _, skill := range skills {
		destDir := filepath.Join(opts.ConfigDir, "skills", skill)
		if err := installFile(
			filepath.Join(opts.KitRoot, "skills", skill, "SKILL.md"),
			filepath.Join(destDir, "SKILL.md"),
		); err != nil {
			return err
		}

		if skill == "pr-review-canvas" {
			for _, asset := range []string{"renderer.js", "styles.css", "template.html"} {
				if err := installFile(
					filepath.Join(opts.KitRoot, "skills", skill, asset),
					filepath.Join(destDir, asset),
				); err != nil {
					return err
				}
			}
		}
	}

	// Install agents
	for _, agent := range agents {
		if err := installFile(
			filepath.Join(opts.KitRoot, "agents", agent+".md"),
			filepath.Join(opts.ConfigDir, "agents", agent+".md"),
		); err != nil {
			return err
		}
	}

	// Install commands
	for _, cmd := range commands {
		if err := installFile(
			filepath.Join(opts.KitRoot, "commands", cmd+".md"),
			filepath.Join(opts.ConfigDir, "commands", cmd+".md"),
		); err != nil {
			return err
		}
	}

	// Install rules
	for _, rule := range rules {
		if err := installFile(
			filepath.Join(opts.KitRoot, "rules", rule+".md"),
			filepath.Join(opts.ConfigDir, "rules", rule+".md"),
		); err != nil {
			return err
		}
	}

	// Plugin (ci-watcher)
	pluginSrc := filepath.Join(opts.KitRoot, "plugins", "ci-watcher.ts")
	pluginDest := filepath.Join(opts.ConfigDir, "plugins", "ci-watcher.ts")

	// Remove old package.json to avoid conflicts (same as shell)
	_ = os.Remove(filepath.Join(opts.ConfigDir, "plugins", "package.json"))

	if err := installFile(pluginSrc, pluginDest); err != nil {
		return err
	}

	// Ensure @opencode-ai/plugin dependency
	if err := ensurePluginDependency(opts.ConfigDir, opts.DryRun); err != nil {
		return err
	}

	if !opts.DryRun {
		fmt.Printf("ntech-team-kit: install complete (%s mode)\n", opts.Mode)
		fmt.Printf("  skills:   %d\n", len(skills))
		fmt.Printf("  agents:   %d\n", len(agents))
		fmt.Printf("  commands: %d\n", len(commands))
		fmt.Printf("  rules:    %d\n", len(rules))
		fmt.Printf("  plugins:  1 (ci-watcher)\n")
		fmt.Println("\nTo enable background CI watching, set:")
		fmt.Println("  export OPENCODE_NTECH_CI_WATCH=1")
	}

	return nil
}

func copyFile(src, dst string) error {
	// Defensive: always ensure the parent directory exists right before writing.
	// This eliminates rare edge cases on macOS where a previous MkdirAll
	// appeared to succeed but the directory was not actually usable when
	// os.Create ran (permissions, sync tools, broken symlinks, etc.).
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", dst, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// ensurePluginDependency replicates the logic from install.sh
func ensurePluginDependency(ocDir string, dryRun bool) error {
	pkgPath := filepath.Join(ocDir, "package.json")

	if dryRun {
		return nil
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		// File doesn't exist — create a minimal one
		pkg := map[string]any{
			"dependencies": map[string]string{
				pluginDep: pluginDepVersion,
			},
		}
		b, _ := json.MarshalIndent(pkg, "", "  ")
		return os.WriteFile(pkgPath, b, 0o644)
	}

	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("invalid package.json: %w", err)
	}

	deps, ok := pkg["dependencies"].(map[string]any)
	if !ok {
		deps = map[string]any{}
		pkg["dependencies"] = deps
	}

	if _, exists := deps[pluginDep]; exists {
		return nil // already present
	}

	deps[pluginDep] = pluginDepVersion

	b, _ := json.MarshalIndent(pkg, "", "  ")
	// backup
	_ = os.WriteFile(pkgPath+".ntech-team-kit.bak", data, 0o644)
	return os.WriteFile(pkgPath, b, 0o644)
}

// logPrefix prints consistent with the old install.sh
func logPrefix(format string, a ...any) {
	fmt.Printf("[ntech-team-kit] "+format+"\n", a...)
}

// PrintStatus replicates the original do_status behavior using the manifest.
func PrintStatus(ocDir string) error {
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")

	data, err := os.ReadFile(manifest)
	if err != nil {
		logPrefix("not installed (no manifest at %s)", manifest)
		return nil
	}

	lines := strings.Split(string(data), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	logPrefix("%d files tracked in manifest", count)

	broken := 0
	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			logPrefix("  MISSING: %s", path)
			broken++
		}
	}

	if broken == 0 {
		logPrefix("all files present")
	} else {
		logPrefix("%d files missing — consider reinstalling", broken)
	}
	return nil
}

// PerformUninstall removes all files listed in the manifest and cleans up
// empty per-skill directories, matching the original shell behavior.
func PerformUninstall(ocDir string) error {
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")

	data, err := os.ReadFile(manifest)
	if err != nil {
		logPrefix("no manifest found at %s — nothing to uninstall", manifest)
		return nil
	}

	lines := strings.Split(string(data), "\n")
	removed := 0
	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err == nil {
				removed++
			}
		}
	}

	// Remove empty skill subdirectories (same as original do_uninstall)
	skillBase := filepath.Join(ocDir, "skills")
	for _, skill := range skills {
		dir := filepath.Join(skillBase, skill)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(dir)
			if len(entries) == 0 {
				_ = os.Remove(dir)
			}
		}
	}

	_ = os.Remove(manifest)

	logPrefix("uninstalled %d files", removed)
	return nil
}
