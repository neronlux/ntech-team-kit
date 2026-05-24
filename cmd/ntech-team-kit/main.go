package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

var (
	version     = "dev"
	defaultRoot string // set via ldflags at build time for Homebrew
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
			ConfigDir:  resolveConfigDir(),
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
		if err := kit.PerformUninstallSelected(resolveConfigDir(), components); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if err := kit.PrintStatus(resolveConfigDir()); err != nil {
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

func runInteractive(root string) error {
	reader := bufio.NewReader(os.Stdin)
	rootValid := kit.ValidateKitRoot(root) == nil

	for {
		fmt.Println()
		fmt.Printf("  ntech-team-kit %s\n", getVersion())
		fmt.Println("  ─────────────────────────────")
		if rootValid {
			fmt.Println("  Kit root: " + root)
		} else {
			fmt.Println("  Kit root: not found (install/update unavailable)")
		}
		configDir := resolveConfigDir()
		manifest := filepath.Join(configDir, ".ntech-team-kit-manifest")
		if entries, err := readManifestCount(manifest); err == nil {
			fmt.Printf("  Installed: %d files\n", entries)
		} else {
			fmt.Println("  Installed: nothing")
		}
		fmt.Println("  ─────────────────────────────")
		fmt.Println()
		fmt.Println("  What would you like to do?")
		fmt.Println()
		if rootValid {
			fmt.Println("    1) Install full pack (recommended)")
			fmt.Println("    2) Install lite pack")
			fmt.Println("    3) Install agents only")
			fmt.Println("    4) Install skills only")
			fmt.Println("    5) Custom install  (pick components)")
		} else {
			fmt.Println("    (install options unavailable — kit root not found)")
		}
		fmt.Println("    6) Custom uninstall (pick components)")
		fmt.Println("    7) Check status")
		fmt.Println("    8) Run doctor")
		if rootValid {
			fmt.Println("    9) Update (refresh all content)")
		} else {
			fmt.Println("    9) Update (unavailable — kit root not found)")
		}
		fmt.Println("    0) Quit")
		fmt.Println()

		choice := prompt(reader, "  Select [1]: ")
		if choice == "" {
			if rootValid {
				choice = "1"
			} else {
				choice = "0"
			}
		}

		var actionErr error
		switch choice {
		case "1":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			actionErr = interactiveInstall(root, kit.FullComponentSet())
		case "2":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			actionErr = interactiveInstall(root, kit.LiteComponentSet())
		case "3":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			actionErr = interactiveInstall(root, kit.ComponentSet{kit.ComponentAgents: true})
		case "4":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			actionErr = interactiveInstall(root, kit.ComponentSet{kit.ComponentSkills: true})
		case "5":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			components, err := promptComponentToggle(reader, "install")
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				continue
			}
			actionErr = interactiveInstall(root, components)
		case "6":
			components, err := promptComponentToggle(reader, "uninstall")
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				continue
			}
			if !confirmAction(reader, "uninstall these components") {
				fmt.Println("  Cancelled.")
				continue
			}
			actionErr = kit.PerformUninstallSelected(configDir, components)
		case "7":
			actionErr = kit.PrintStatus(configDir)
		case "8":
			actionErr = runDoctorInteractive(root)
		case "9":
			if !rootValid {
				fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
				continue
			}
			runUpdate(root)
		case "0", "q", "quit":
			fmt.Println("\n  Bye.")
			return nil
		default:
			fmt.Printf("\n  Unknown selection: %s  (enter 0-9)\n", choice)
			continue
		}

		if actionErr != nil {
			fmt.Fprintf(os.Stderr, "\n  Error: %v\n", actionErr)
		}

		fmt.Println()
		if !isTerminal(os.Stdin) {
			break
		}
		again := prompt(reader, "  Do something else? [y/N]: ")
		if again != "y" && again != "yes" {
			break
		}
	}

	fmt.Println("\n  Bye.")
	return nil
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func interactiveInstall(root string, components kit.ComponentSet) error {
	return kit.PerformInstall(kit.InstallOptions{
		KitRoot:    root,
		ConfigDir:  resolveConfigDir(),
		Mode:       "copy",
		Components: components,
	})
}

func promptComponentToggle(reader *bufio.Reader, action string) (kit.ComponentSet, error) {
	allComponents := []string{
		kit.ComponentSkills,
		kit.ComponentAgents,
		kit.ComponentCommands,
		kit.ComponentRules,
		kit.ComponentPlugin,
		kit.ComponentConfig,
	}
	descriptions := map[string]string{
		kit.ComponentSkills:   "On-demand workflows (review, CI, shipping)",
		kit.ComponentAgents:   "Specialized subagents (ci-watcher, code-quality)",
		kit.ComponentCommands: "/command shortcuts for every skill",
		kit.ComponentRules:    "Auto-loaded coding standards",
		kit.ComponentPlugin:   "Background CI watcher plugin",
		kit.ComponentConfig:   "opencode.jsonc defaults (first-time only)",
	}

	fmt.Printf("\n  Select components to %s.\n", action)
	fmt.Println("  Toggle with numbers, or type the component names comma-separated.")
	fmt.Println("  Press Enter when done, or type a pack name (full, lite).")
	fmt.Println()

	defaultOn := kit.FullComponentSet()
	if action == "uninstall" {
		defaultOn = nil
	}
	selected := make(kit.ComponentSet)
	for i, comp := range allComponents {
		marker := " "
		if defaultOn != nil && defaultOn.Includes(comp) {
			marker = "*"
		}
		fmt.Printf("    [%s] %d) %-10s %s\n", marker, i+1, comp, descriptions[comp])
	}
	fmt.Println()
	fmt.Println("  Examples: \"1 3 5\" or \"skills,commands\" or \"lite\" or Enter to confirm")

	for {
		value := prompt(reader, "  Components: ")
		if value == "" {
			if len(selected) == 0 {
				if defaultOn != nil {
					return defaultOn, nil
				}
				return kit.FullComponentSet(), nil
			}
			return selected, nil
		}
		if value == "full" {
			return kit.FullComponentSet(), nil
		}
		if value == "lite" {
			return kit.LiteComponentSet(), nil
		}
		if value == "agents" {
			return kit.ComponentSet{kit.ComponentAgents: true}, nil
		}
		if value == "skills" {
			return kit.ComponentSet{kit.ComponentSkills: true}, nil
		}

		numberSet := make(map[int]bool)
		allNumbers := true
		for _, part := range strings.Fields(value) {
			var n int
			if _, err := fmt.Sscanf(part, "%d", &n); err == nil && n >= 1 && n <= len(allComponents) {
				numberSet[n] = true
			} else {
				allNumbers = false
				break
			}
		}

		if allNumbers && len(numberSet) > 0 {
			selected = make(kit.ComponentSet)
			for n := range numberSet {
				selected[allComponents[n-1]] = true
			}
			fmt.Printf("  Selected: %s\n", strings.Join(selected.Names(), ", "))
			continue
		}

		parsed, err := parseComponentList(value)
		if err != nil {
			fmt.Printf("  %v  Try again.\n", err)
			continue
		}
		return parsed, nil
	}
}

func promptInstallComponents(reader *bufio.Reader) (kit.ComponentSet, error) {
	return promptComponentToggle(reader, "install")
}

func promptUninstallComponents(reader *bufio.Reader) (kit.ComponentSet, error) {
	return promptComponentToggle(reader, "uninstall")
}

func confirmAction(reader *bufio.Reader, action string) bool {
	confirm := prompt(reader, fmt.Sprintf("  Confirm %s? [y/N]: ", action))
	return confirm == "y" || confirm == "yes"
}

func prompt(reader *bufio.Reader, label string) string {
	fmt.Print(label)
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF && line != "" {
			return strings.TrimSpace(line)
		}
		return ""
	}
	return strings.TrimSpace(line)
}

func readManifestCount(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count, nil
}

func runDoctorInteractive(root string) error {
	results := kit.RunDoctor(root)
	allGood := true
	for _, r := range results {
		status := "✅"
		if !r.Passed {
			status = "❌"
			allGood = false
		}
		fmt.Printf("  %s %-20s %s\n", status, r.Name, r.Message)
	}
	if !allGood {
		fmt.Println("\n  Some checks failed. See above.")
		return fmt.Errorf("some checks failed")
	}
	fmt.Println("\n  All checks passed.")

	cur := getVersion()
	stamp := kit.DefaultUpdateStampPath()
	if kit.IsStale(stamp) {
		if latest, avail, err := kit.CheckForUpdate(cur); err == nil && avail {
			fmt.Printf("\n  💡 Update available: v%s → v%s  (option 9 or: ntech-team-kit update)\n", cur, latest)
		}
		_ = kit.TouchStamp(stamp)
	}
	return nil
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
		ConfigDir:  resolveConfigDir(),
		Mode:       "copy",
		Components: kit.FullComponentSet(),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Content refresh failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Update complete. Content is now fresh from this kit tree.")
	_ = kit.TouchStamp(kit.DefaultUpdateStampPath())
}

func runDoctor(root string) {
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

	cur := getVersion()
	stamp := kit.DefaultUpdateStampPath()
	if kit.IsStale(stamp) {
		if latest, avail, err := kit.CheckForUpdate(cur); err == nil && avail {
			fmt.Printf("\n💡 Update available: v%s → v%s  (run: ntech-team-kit update)\n", cur, latest)
		}
		_ = kit.TouchStamp(stamp)
	}
}

func hasArg(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func parseInstallComponents(args []string) (kit.ComponentSet, error) {
	components := kit.FullComponentSet()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--link":
			continue
		case arg == "--select":
			continue
		case arg == "--pack":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--pack requires a value")
			}
			next, err := componentSetForPack(args[i+1])
			if err != nil {
				return nil, err
			}
			components = next
			i++
		case strings.HasPrefix(arg, "--pack="):
			next, err := componentSetForPack(strings.TrimPrefix(arg, "--pack="))
			if err != nil {
				return nil, err
			}
			components = next
		case arg == "--only":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--only requires a comma-separated component list")
			}
			next, err := parseComponentList(args[i+1])
			if err != nil {
				return nil, err
			}
			components = next
			i++
		case strings.HasPrefix(arg, "--only="):
			next, err := parseComponentList(strings.TrimPrefix(arg, "--only="))
			if err != nil {
				return nil, err
			}
			components = next
		case arg == "--without":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--without requires a comma-separated component list")
			}
			if err := removeComponents(components, args[i+1]); err != nil {
				return nil, err
			}
			i++
		case strings.HasPrefix(arg, "--without="):
			if err := removeComponents(components, strings.TrimPrefix(arg, "--without=")); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown install option: %s", arg)
		}
	}
	return components, nil
}

func parseUninstallComponents(args []string) (kit.ComponentSet, error) {
	if len(args) == 0 {
		return nil, nil
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--select":
			continue
		case arg == "--only":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--only requires a comma-separated component list")
			}
			components, err := parseComponentList(args[i+1])
			if err != nil {
				return nil, err
			}
			if i+2 != len(args) {
				return nil, fmt.Errorf("unexpected uninstall option: %s", args[i+2])
			}
			return components, nil
		case strings.HasPrefix(arg, "--only="):
			if i+1 != len(args) {
				return nil, fmt.Errorf("unexpected uninstall option: %s", args[i+1])
			}
			return parseComponentList(strings.TrimPrefix(arg, "--only="))
		default:
			return nil, fmt.Errorf("unknown uninstall option: %s", arg)
		}
	}
	return nil, nil
}

func componentSetForPack(name string) (kit.ComponentSet, error) {
	switch name {
	case "full":
		return kit.FullComponentSet(), nil
	case "lite":
		return kit.LiteComponentSet(), nil
	case "agents":
		return kit.ComponentSet{kit.ComponentAgents: true}, nil
	case "skills":
		return kit.ComponentSet{kit.ComponentSkills: true}, nil
	default:
		return nil, fmt.Errorf("unknown pack %q (expected full, lite, agents, or skills)", name)
	}
}

func parseComponentList(value string) (kit.ComponentSet, error) {
	components := kit.ComponentSet{}
	for _, raw := range strings.Split(value, ",") {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if !kit.ValidComponent(name) {
			return nil, fmt.Errorf("unknown component %q", name)
		}
		components[name] = true
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("component list cannot be empty")
	}
	return components, nil
}

func removeComponents(components kit.ComponentSet, value string) error {
	remove, err := parseComponentList(value)
	if err != nil {
		return err
	}
	for name := range remove {
		delete(components, name)
	}
	if len(components) == 0 {
		return fmt.Errorf("selection cannot exclude every component")
	}
	return nil
}

func resolveConfigDir() string {
	if dir := os.Getenv("OPENCODE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
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

	root := kit.GetKitRoot()
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
