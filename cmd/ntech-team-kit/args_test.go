package main

import "testing"

func TestSplitTargetOption_LongForm(t *testing.T) {
	target, remaining, err := splitTargetOption([]string{"--target", "codex", "--only", "skills"})
	if err != nil {
		t.Fatalf("splitTargetOption failed: %v", err)
	}
	if target != "codex" {
		t.Fatalf("target = %q, want codex", target)
	}
	if len(remaining) != 2 || remaining[0] != "--only" || remaining[1] != "skills" {
		t.Fatalf("remaining args = %v", remaining)
	}
}

func TestSplitTargetOption_EqualsForm(t *testing.T) {
	target, remaining, err := splitTargetOption([]string{"--target=both", "--pack", "lite"})
	if err != nil {
		t.Fatalf("splitTargetOption failed: %v", err)
	}
	if target != "both" {
		t.Fatalf("target = %q, want both", target)
	}
	if len(remaining) != 2 || remaining[0] != "--pack" || remaining[1] != "lite" {
		t.Fatalf("remaining args = %v", remaining)
	}
}

func TestSplitTargetOption_DefaultsToOpenCode(t *testing.T) {
	target, remaining, err := splitTargetOption([]string{"--pack", "lite"})
	if err != nil {
		t.Fatalf("splitTargetOption failed: %v", err)
	}
	if target != "opencode" {
		t.Fatalf("target = %q, want opencode", target)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining args = %v", remaining)
	}
}
