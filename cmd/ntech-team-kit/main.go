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
		mode := "copy"
		for _, a := range args {
			if a == "--link" {
				mode = "link"
			}
		}
		var err error
		components := kit.FullComponentSet()
		if hasArg(args, "--select") {
			components, err = promptInstallComponents(bufio.NewReader(os.Stdin))
		} else {
			components, err = parseInstallComponents(args)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := kit.PerformInstall(kit.InstallOptions{
			KitRoot:    root,
			ConfigDir:  kit.ConfigDir(),
			Mode:       mode,
			Components: components,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "uninstall":
		var components kit.ComponentSet
		var err error
		if hasArg(args, "--select") {
			components, err = promptUninstallComponents(bufio.NewReader(os.Stdin))
		} else {
			components, err = parseUninstallComponents(args)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := kit.PerformUninstallSelected(kit.ConfigDir(), components); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if err := kit.PrintStatus(kit.ConfigDir()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "update":
		runUpdate(root)

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
			status = "❌"
			allGood = false
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

func runUpdate(root string) {
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

	fmt.Println("\n→ Refreshing all kit components into OpenCode config...")
	if err := kit.PerformInstall(kit.InstallOptions{
		KitRoot:    root,
		ConfigDir:  kit.ConfigDir(),
		Mode:       "copy",
		Components: kit.FullComponentSet(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Content refresh failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Update complete. Content is now fresh from this kit tree.")
	_ = kit.TouchStamp(kit.DefaultUpdateStampPath())
}

func printUsage() {
	fmt.Print(`ntech-team-kit - OpenCode skills, agents, and rules installer

Usage:
  ntech-team-kit                     Start interactive setup
  ntech-team-kit <command> [options]

Commands:
  install     Install selected kit components
  uninstall   Remove installed files, optionally by component
  status      Show current installation status
  update      Check for newer CLI + refresh all skills/agents/commands/rules
  doctor      Run health checks (OpenCode, gh, auth, etc.)
  version     Print version
  path        Print the resolved kit root directory

Global options:
  --root <path>   Override kit location (useful for development)

Environment variables:
  NTECH_TEAM_KIT_ROOT   Same as --root
  OPENCODE_CONFIG_DIR   Override ~/.config/opencode location

Examples:
  ntech-team-kit install
  ntech-team-kit install --pack lite
  ntech-team-kit install --select
  ntech-team-kit install --only skills,commands
  ntech-team-kit install --without plugin,agents
  ntech-team-kit install --link
  ntech-team-kit uninstall --select
  ntech-team-kit uninstall --only agents
  ntech-team-kit doctor
  NTECH_TEAM_KIT_ROOT=/path/to/kit ntech-team-kit status

Install options:
  --pack full|lite|agents|skills   Install a named component pack (default: full)
  --select                         Choose components interactively
  --only <components>              Install only a comma-separated component list
  --without <components>           Exclude components from the selected pack
  --link                           Symlink instead of copy

Components:
  skills, agents, commands, rules, plugin, config
`)
}
