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

func TestPromptComponentToggle_CodexTargetDefaultsToSkillsAndAgents(t *testing.T) {
	components, err := promptComponentToggleForTarget(bufio.NewReader(strings.NewReader("\n")), "install", "codex")
	if err != nil {
		t.Fatalf("promptComponentToggleForTarget failed: %v", err)
	}
	if !components.Includes(kit.ComponentSkills) {
		t.Fatalf("Codex target should include skills, got %v", components.Names())
	}
	if !components.Includes(kit.ComponentAgents) {
		t.Fatalf("Codex target should include generated Codex agents, got %v", components.Names())
	}
	if components.Includes(kit.ComponentCommands) {
		t.Fatalf("Codex target should not include OpenCode-only commands, got %v", components.Names())
	}
}

func TestPromptComponentToggle_CodexTargetCanSelectAgentsOnly(t *testing.T) {
	components, err := promptComponentToggleForTarget(bufio.NewReader(strings.NewReader("2\n\n")), "install", "codex")
	if err != nil {
		t.Fatalf("promptComponentToggleForTarget failed: %v", err)
	}
	if !components.Includes(kit.ComponentAgents) {
		t.Fatalf("Codex target should include agents, got %v", components.Names())
	}
	if components.Includes(kit.ComponentSkills) {
		t.Fatalf("Codex target should not include skills when selecting agents only, got %v", components.Names())
	}
}

func TestMaybeRefreshCodexSkillsView_RunsForCodexSkillInstall(t *testing.T) {
	previous := refreshCodexSkillsView
	defer func() {
		refreshCodexSkillsView = previous
	}()

	calls := 0
	refreshCodexSkillsView = func() error {
		calls++
		return nil
	}

	maybeRefreshCodexSkillsView("codex", kit.ComponentSet{kit.ComponentSkills: true})

	if calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", calls)
	}
}

func TestMaybeRefreshCodexSkillsView_SkipsWhenDisabled(t *testing.T) {
	t.Setenv("NTECH_TEAM_KIT_CODEX_SKIP_APP_REFRESH", "1")
	previous := refreshCodexSkillsView
	defer func() {
		refreshCodexSkillsView = previous
	}()

	calls := 0
	refreshCodexSkillsView = func() error {
		calls++
		return nil
	}

	maybeRefreshCodexSkillsView("codex", kit.ComponentSet{kit.ComponentSkills: true})

	if calls != 0 {
		t.Fatalf("refresh calls = %d, want 0", calls)
	}
}
