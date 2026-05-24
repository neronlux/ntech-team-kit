package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

func runInteractive(root string) error {
	reader := bufio.NewReader(os.Stdin)
	rootValid := kit.ValidateKitRoot(root) == nil

	requireRoot := func() bool {
		if !rootValid {
			fmt.Println("\n  Kit root not found. Use --root <path> to specify.")
			return false
		}
		return true
	}

	for {
		fmt.Println()
		fmt.Printf("  ntech-team-kit %s\n", getVersion())
		fmt.Println("  ─────────────────────────────")
		if rootValid {
			fmt.Println("  Kit root: " + root)
		} else {
			fmt.Println("  Kit root: not found (install/update unavailable)")
		}
		configDir := kit.ConfigDir()
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
			if !requireRoot() {
				continue
			}
			actionErr = performInstall(root, kit.FullComponentSet())
		case "2":
			if !requireRoot() {
				continue
			}
			actionErr = performInstall(root, kit.LiteComponentSet())
		case "3":
			if !requireRoot() {
				continue
			}
			actionErr = performInstall(root, kit.ComponentSet{kit.ComponentAgents: true})
		case "4":
			if !requireRoot() {
				continue
			}
			actionErr = performInstall(root, kit.ComponentSet{kit.ComponentSkills: true})
		case "5":
			if !requireRoot() {
				continue
			}
			components, err := promptComponentToggle(reader, "install")
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				continue
			}
			actionErr = performInstall(root, components)
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
			actionErr = printDoctorResults(root, "  ")
		case "9":
			if !requireRoot() {
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

func performInstall(root string, components kit.ComponentSet) error {
	return kit.PerformInstall(kit.InstallOptions{
		KitRoot:    root,
		ConfigDir:  kit.ConfigDir(),
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

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
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
