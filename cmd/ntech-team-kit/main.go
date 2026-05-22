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

	switch command {
	case "install", "uninstall", "status":
		// Delegate to the battle-tested install.sh
		if err := kit.RunInstaller(root, append([]string{command}, args...)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

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

func printUsage() {
	fmt.Print(`ntech-team-kit - OpenCode skills, agents, and rules installer

Usage:
  ntech-team-kit <command> [options]

Commands:
  install     Install skills, agents, commands, rules, and plugin
  uninstall   Remove all installed files
  status      Show current installation status
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
