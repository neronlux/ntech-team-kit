package main

import (
	"fmt"
	"strings"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

func hasArg(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func splitTargetOption(args []string) (string, []string, error) {
	target := "opencode"
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--target":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--target requires a value")
			}
			next, err := kit.NormalizeInstallTarget(args[i+1])
			if err != nil {
				return "", nil, err
			}
			target = next
			i++
		case strings.HasPrefix(arg, "--target="):
			next, err := kit.NormalizeInstallTarget(strings.TrimPrefix(arg, "--target="))
			if err != nil {
				return "", nil, err
			}
			target = next
		default:
			remaining = append(remaining, arg)
		}
	}
	return target, remaining, nil
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
