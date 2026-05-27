package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/neronlux/ntech-team-kit/internal/kit"
)

func TestPromptComponentToggle_UninstallEmptySelectionErrors(t *testing.T) {
	components, err := promptComponentToggle(bufio.NewReader(strings.NewReader("\n")), "uninstall")
	if err == nil {
		t.Fatalf("expected empty uninstall selection to fail, got components=%v", components)
	}
}

func TestPromptComponentToggle_UninstallFullSelectionIsExplicit(t *testing.T) {
	components, err := promptComponentToggle(bufio.NewReader(strings.NewReader("full\n")), "uninstall")
	if err != nil {
		t.Fatalf("explicit full uninstall selection failed: %v", err)
	}
	if !components.Includes(kit.ComponentSkills) || !components.Includes(kit.ComponentPlugin) {
		t.Fatalf("explicit full uninstall should include all components, got %v", components.Names())
	}
}

func TestPromptTarget_SelectsCodexByNumber(t *testing.T) {
	target, err := promptTarget(bufio.NewReader(strings.NewReader("2\n")), "install")
	if err != nil {
		t.Fatalf("promptTarget failed: %v", err)
	}
	if target != "codex" {
		t.Fatalf("target = %q, want codex", target)
	}
}

func TestPromptTarget_DefaultsToOpenCode(t *testing.T) {
	target, err := promptTarget(bufio.NewReader(strings.NewReader("\n")), "install")
	if err != nil {
		t.Fatalf("promptTarget failed: %v", err)
	}
	if target != "opencode" {
		t.Fatalf("target = %q, want opencode", target)
	}
}

func TestPromptComponentToggle_CodexTargetUsesSkills(t *testing.T) {
	components, err := promptComponentToggleForTarget(bufio.NewReader(strings.NewReader("")), "install", "codex")
	if err != nil {
		t.Fatalf("promptComponentToggleForTarget failed: %v", err)
	}
	if !components.Includes(kit.ComponentSkills) {
		t.Fatalf("Codex target should include skills, got %v", components.Names())
	}
	if components.Includes(kit.ComponentAgents) {
		t.Fatalf("Codex target should not include OpenCode-only agents, got %v", components.Names())
	}
}
