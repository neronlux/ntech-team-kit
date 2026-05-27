package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

var (
	version     = "dev"
	kitRootFlag string
)

var refreshCodexSkillsView = kit.RefreshCodexSkillsView

func main() {
	flag.StringVar(&kitRootFlag, "root", "", "Path to ntech-team-kit directory (overrides auto-detection)")
	flag.Parse()

	root := resolveKitRoot()

	if len(flag.Args()) == 0 {
		if err := runInteractive(root); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	command := flag.Arg(0)
	args := flag.Args()[1:]

	if command == "install" || command == "update" {
		if err := kit.ValidateKitRoot(root); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			fmt.Fprintln(os.Stderr, "Use --root <path> or NTECH_TEAM_KIT_ROOT=/path/to/kit to override.")
			os.Exit(1)
		}
	}

	switch command {
	case "install":
		target, args, err := splitTargetOption(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		mode := "copy"
		for _, a := range args {
			if a == "--link" {
				mode = "link"
			}
		}
		components := kit.FullComponentSet()
		if hasArg(args, "--select") {
			target, err = resolveTargetForComponentPrompt(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			components, err = promptInstallComponentsForTarget(bufio.NewReader(os.Stdin), target)
		} else {
			components, err = parseInstallComponents(args)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := performInstallTarget(root, mode, components, target); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "uninstall":
		target, args, err := splitTargetOption(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		var components kit.ComponentSet
		if hasArg(args, "--select") {
			target, err = resolveTargetForComponentPrompt(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			components, err = promptUninstallComponentsForTarget(bufio.NewReader(os.Stdin), target)
		} else {
			components, err = parseUninstallComponents(args)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := performUninstallTarget(components, target); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		target, _, err := splitTargetOption(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := performStatusTarget(target); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "update":
		target, _, err := splitTargetOption(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		runUpdate(root, target)

	case "doctor":
		runDoctor(root)

	case "version":
		fmt.Println("ntech-team-kit", getVersion())

	case "path":
		fmt.Println(root)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func resolveKitRoot() string {
	if kitRootFlag != "" {
		abs, _ := filepath.Abs(kitRootFlag)
		return abs
	}
	return kit.GetKitRoot()
}

func getVersion() string {
	if version != "" && version != "dev" {
		return version
	}

	if root := kit.GetKitRoot(); root != "" {
		if data, err := os.ReadFile(filepath.Join(root, "VERSION")); err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v
			}
		}
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

func printDoctorResults(root string, indent string) error {
	results := kit.RunDoctor(root)
	allGood := true
	for _, r := range results {
		status := "✅"
		if !r.Passed {
			if r.Optional {
				status = "⚠️"
			} else {
				status = "❌"
				allGood = false
			}
		}
		fmt.Printf("%s%s %-20s %s\n", indent, status, r.Name, r.Message)
	}
	if !allGood {
		fmt.Printf("\n%sSome checks failed. See above.\n", indent)
		return fmt.Errorf("some checks failed")
	}
	fmt.Printf("\n%sAll checks passed.\n", indent)

	cur := getVersion()
	stamp := kit.DefaultUpdateStampPath()
	if kit.IsStale(stamp) {
		if latest, avail, err := kit.CheckForUpdate(cur); err == nil && avail {
			fmt.Printf("\n%s💡 Update available: v%s → v%s  (run: ntech-team-kit update)\n", indent, cur, latest)
		}
		_ = kit.TouchStamp(stamp)
	}
	return nil
}

func runDoctor(root string) {
	if err := printDoctorResults(root, ""); err != nil {
		os.Exit(1)
	}
}

func resolveTargetForComponentPrompt(target string) (string, error) {
	if target != "auto" {
		return target, nil
	}
	resolved, err := kit.ResolveInstallTarget(target)
	if err != nil {
		return "", err
	}
	fmt.Printf("Auto-detected target: %s\n", resolved)
	return resolved, nil
}

func performInstallTarget(root string, mode string, components kit.ComponentSet, target string) error {
	target, err := kit.ResolveInstallTarget(target)
	if err != nil {
		return err
	}
	if kit.TargetIncludesOpenCode(target) {
		if err := kit.PerformInstall(kit.InstallOptions{
			KitRoot:    root,
			ConfigDir:  kit.ConfigDir(),
			Mode:       mode,
			Components: components,
		}); err != nil {
			return err
		}
	}
	if kit.TargetIncludesCodex(target) {
		if !components.Includes(kit.ComponentSkills) && !components.Includes(kit.ComponentAgents) && target == "codex" {
			return fmt.Errorf("Codex target supports skills and agents; include at least one")
		}
		if components.Includes(kit.ComponentSkills) {
			if err := kit.PerformCodexInstall(kit.CodexInstallOptions{
				KitRoot:   root,
				SkillsDir: kit.CodexSkillsDir(),
				AgentsDir: kit.CodexAgentsDir(),
				Mode:      mode,
			}); err != nil {
				return err
			}
			maybeRefreshCodexSkillsView(target, components)
		}
		if components.Includes(kit.ComponentAgents) {
			if err := kit.PerformCodexAgentInstall(kit.CodexAgentInstallOptions{
				KitRoot:   root,
				SkillsDir: kit.CodexSkillsDir(),
				AgentsDir: kit.CodexAgentsDir(),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func maybeRefreshCodexSkillsView(target string, components kit.ComponentSet) {
	if !kit.TargetIncludesCodex(target) || !components.Includes(kit.ComponentSkills) {
		return
	}
	if os.Getenv("NTECH_TEAM_KIT_CODEX_SKIP_APP_REFRESH") == "1" {
		return
	}
	if err := refreshCodexSkillsView(); err != nil {
		fmt.Fprintf(os.Stderr, "Codex app refresh skipped: %v\n", err)
		return
	}
	fmt.Println("Codex app: opened Skills view")
}

func performUninstallTarget(components kit.ComponentSet, target string) error {
	target, err := kit.ResolveInstallTarget(target)
	if err != nil {
		return err
	}
	if kit.TargetIncludesOpenCode(target) {
		if err := kit.PerformUninstallSelected(kit.ConfigDir(), components); err != nil {
			return err
		}
	}
	if kit.TargetIncludesCodex(target) {
		if !components.Includes(kit.ComponentSkills) && !components.Includes(kit.ComponentAgents) && target == "codex" {
			return fmt.Errorf("Codex target supports skills and agents; include at least one")
		}
		if components.Includes(kit.ComponentSkills) || components.Includes(kit.ComponentAgents) {
			return kit.PerformCodexUninstallSelected(kit.CodexSkillsDir(), kit.CodexAgentsDir(), components)
		}
	}
	return nil
}

func performStatusTarget(target string) error {
	target, err := kit.NormalizeInstallTarget(target)
	if err != nil {
		return err
	}
	if target == "auto" {
		resolved, err := kit.ResolveInstallTarget(target)
		if err == nil {
			fmt.Printf("Auto-detected target: %s\n", resolved)
			target = resolved
		} else {
			fmt.Printf("Auto-detection failed: %v\n", err)
			fmt.Println("Showing both status reports instead.")
			target = "both"
		}
	}
	if kit.TargetIncludesOpenCode(target) {
		fmt.Println("OpenCode status:")
		if err := kit.PrintStatus(kit.ConfigDir()); err != nil {
			return err
		}
	}
	if kit.TargetIncludesCodex(target) {
		fmt.Println("Codex status:")
		if err := kit.PrintCodexStatusWithDirs(kit.CodexSkillsDir(), kit.CodexAgentsDir()); err != nil {
			return err
		}
	}
	return nil
}

func runUpdate(root string, target string) {
	cur := getVersion()
	fmt.Printf("Current CLI: %s\n", cur)
	if latest, avail, err := kit.CheckForUpdate(cur); err == nil {
		if avail {
			fmt.Printf("New version available: v%s\n", latest)
			fmt.Println("→ Recommended: brew upgrade ntech-team-kit   (or git pull + rebuild)")
		} else {
			fmt.Println("CLI binary is up to date.")
		}
	} else {
		fmt.Printf("Could not reach GitHub for version check: %v\n", err)
	}

	fmt.Println("\n→ Refreshing kit components...")
	if err := performInstallTarget(root, "copy", kit.FullComponentSet(), target); err != nil {
		fmt.Fprintf(os.Stderr, "Content refresh failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Update complete. Content is now fresh from this kit tree.")
	_ = kit.TouchStamp(kit.DefaultUpdateStampPath())
}

func printUsage() {
	fmt.Print(`ntech-team-kit - OpenCode and Codex skills, agents, and rules installer

Usage:
  ntech-team-kit                     Start interactive setup
  ntech-team-kit <command> [options]

Commands:
  install     Install selected kit components
  uninstall   Remove installed files, optionally by component
  status      Show current installation status
  update      Check for newer CLI + refresh all skills/agents/commands/rules
  doctor      Run health checks (OpenCode, Codex, gh, auth, etc.)
  version     Print version
  path        Print the resolved kit root directory

Global options:
  --root <path>   Override kit location (useful for development)

Environment variables:
  NTECH_TEAM_KIT_ROOT   Same as --root
  OPENCODE_CONFIG_DIR   Override ~/.config/opencode location
  NTECH_TEAM_KIT_CODEX_SKILLS_DIR   Override ~/.agents/skills for Codex skills
  NTECH_TEAM_KIT_CODEX_AGENTS_DIR   Override ~/.codex/agents for Codex agents
  NTECH_TEAM_KIT_CODEX_SKIP_APP_REFRESH=1   Skip opening the Codex Skills view

Examples:
  ntech-team-kit install
  ntech-team-kit install --target codex
  ntech-team-kit install --target both
  ntech-team-kit install --pack lite
  ntech-team-kit install --select
  ntech-team-kit install --only skills,commands
  ntech-team-kit install --without plugin,agents
  ntech-team-kit install --link
  ntech-team-kit status --target codex
  ntech-team-kit uninstall --select
  ntech-team-kit uninstall --only agents
  ntech-team-kit doctor
  NTECH_TEAM_KIT_ROOT=/path/to/kit ntech-team-kit status

Install options:
  --target opencode|codex|both|auto   Target for install/status/update/uninstall (default: opencode)
  --pack full|lite|agents|skills   Install a named component pack (default: full)
  --select                         Choose components interactively
  --only <components>              Install only a comma-separated component list
  --without <components>           Exclude components from the selected pack
  --link                           Symlink instead of copy

Components:
  skills, agents, commands, rules, plugin, config
`)
}
