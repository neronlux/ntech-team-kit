package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

var (
	version      = "dev"
	defaultRoot  string // set via ldflags at build time for Homebrew
	kitRootFlag  string
)

func main() {
	// Global flags
	flag.StringVar(&kitRootFlag, "root", "", "Path to ntech-team-kit directory (overrides auto-detection)")

	flag.Parse()

	if len(flag.Args()) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	args := flag.Args()[1:]

	// Resolve final kit root
	root := resolveKitRoot()

	// Pre-flight validation for any command that needs a real kit tree.
	// This gives a clear, actionable error instead of letting install.sh
	// fail with cryptic "cp: ... No such file or directory".
	if command == "install" || command == "uninstall" || command == "status" || command == "update" {
		if err := kit.ValidateKitRoot(root); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
			fmt.Fprintln(os.Stderr, "Use --root <path> or NTECH_TEAM_KIT_ROOT=/path/to/kit to override.")
			os.Exit(1)
		}
	}

	switch command {
	case "install":
		// Use the robust pure-Go installer (no shell fragility)
		mode := "copy"
		for _, a := range args {
			if a == "--link" {
				mode = "link"
			}
		}
		ocDir := os.Getenv("OPENCODE_CONFIG_DIR")
		if ocDir == "" {
			home, _ := os.UserHomeDir()
			ocDir = filepath.Join(home, ".config", "opencode")
		}
		if err := kit.PerformInstall(kit.InstallOptions{
			KitRoot:   root,
			ConfigDir: ocDir,
			Mode:      mode,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "uninstall", "status":
		// These still use the shell script (they operate on the manifest)
		if err := kit.RunInstaller(root, append([]string{command}, args...)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "update":
		cur := getVersion()

		// 1. CLI version check (GitHub)
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

		// 2. Always refresh the kit content using the robust pure-Go installer.
		//    This is the critical path that must never fail with cryptic errors.
		fmt.Println("\n→ Refreshing skills, agents, commands and rules into OpenCode config...")
		ocDir := os.Getenv("OPENCODE_CONFIG_DIR")
		if ocDir == "" {
			home, _ := os.UserHomeDir()
			ocDir = filepath.Join(home, ".config", "opencode")
		}
		if err := kit.PerformInstall(kit.InstallOptions{
			KitRoot:   root,
			ConfigDir: ocDir,
			Mode:      "copy",
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Content refresh failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✓ Update complete. Content is now fresh from this kit tree.")
		_ = kit.TouchStamp(kit.DefaultUpdateStampPath())

	case "doctor":
		results := kit.RunDoctor(root)
		allGood := true
		for _, r := range results {
			status := "✅"
			if !r.Passed {
				status = "❌"
				allGood = false
			}
			fmt.Printf("%s %-20s %s\n", status, r.Name, r.Message)
		}
		if !allGood {
			fmt.Println("\nSome checks failed. See above.")
			os.Exit(1)
		}
		fmt.Println("\nAll checks passed.")

		// Non-blocking update hint (once per day)
		cur := getVersion()
		stamp := kit.DefaultUpdateStampPath()
		if kit.IsStale(stamp) {
			if latest, avail, err := kit.CheckForUpdate(cur); err == nil && avail {
				fmt.Printf("\n💡 Update available: v%s → v%s  (run: ntech-team-kit update)\n", cur, latest)
			}
			_ = kit.TouchStamp(stamp)
		}

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

	if env := os.Getenv("NTECH_TEAM_KIT_ROOT"); env != "" {
		return env
	}

	if defaultRoot != "" {
		return defaultRoot
	}

	// Auto-detect
	root := kit.GetKitRoot()
	if root == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not auto-detect ntech-team-kit location.")
		fmt.Fprintln(os.Stderr, "Set NTECH_TEAM_KIT_ROOT environment variable or use --root /path/to/kit")
		os.Exit(1)
	}
	return root
}

func getVersion() string {
	// 1. ldflags override (used in releases and Homebrew builds)
	if version != "" && version != "dev" {
		return version
	}

	// 2. Try to read VERSION file from kit root
	if root := kit.GetKitRoot(); root != "" {
		if data, err := os.ReadFile(filepath.Join(root, "VERSION")); err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v
			}
		}
	}

	// 3. Fallback to build info or "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

func hasYes(args []string) bool {
	for _, a := range args {
		if a == "--yes" || a == "-y" {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Print(`ntech-team-kit - OpenCode skills, agents, and rules installer

Usage:
  ntech-team-kit <command> [options]

Commands:
  install     Install skills, agents, commands, rules, and plugin
  uninstall   Remove all installed files
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
  ntech-team-kit install --copy
  ntech-team-kit doctor
  NTECH_TEAM_KIT_ROOT=/path/to/kit ntech-team-kit status
`)
}
