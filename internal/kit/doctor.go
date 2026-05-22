package kit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

func RunDoctor(kitRoot string) []CheckResult {
	var results []CheckResult

	results = append(results, checkOpenCode())
	results = append(results, checkGhCLI())
	results = append(results, checkGhAuth())
	results = append(results, checkKitRoot(kitRoot))
	results = append(results, checkManifest())
	results = append(results, checkKitContents(kitRoot))

	return results
}

func checkOpenCode() CheckResult {
	// Check common locations for opencode binary
	paths := []string{"opencode", "opencode-ai"}
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			return CheckResult{Name: "OpenCode", Passed: true, Message: "found in PATH"}
		}
	}
	return CheckResult{Name: "OpenCode", Passed: false, Message: "opencode not found in PATH. Install from https://opencode.ai"}
}

func checkGhCLI() CheckResult {
	if _, err := exec.LookPath("gh"); err == nil {
		return CheckResult{Name: "GitHub CLI (gh)", Passed: true, Message: "found in PATH"}
	}
	return CheckResult{Name: "GitHub CLI (gh)", Passed: false, Message: "gh not found. Install from https://cli.github.com"}
}

func checkGhAuth() CheckResult {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{Name: "gh auth", Passed: false, Message: "gh is not authenticated. Run: gh auth login"}
	}
	if strings.Contains(string(output), "Logged in") {
		return CheckResult{Name: "gh auth", Passed: true, Message: "authenticated"}
	}
	return CheckResult{Name: "gh auth", Passed: false, Message: "gh authentication issue"}
}

func checkKitRoot(root string) CheckResult {
	if root == "" {
		return CheckResult{Name: "Kit Root", Passed: false, Message: "could not auto-detect kit root. Set NTECH_TEAM_KIT_ROOT or use --root"}
	}
	if _, err := os.Stat(root); err != nil {
		return CheckResult{Name: "Kit Root", Passed: false, Message: fmt.Sprintf("root does not exist: %s", root)}
	}
	return CheckResult{Name: "Kit Root", Passed: true, Message: root}
}

func checkManifest() CheckResult {
	ocDir := os.Getenv("OPENCODE_CONFIG_DIR")
	if ocDir == "" {
		ocDir = filepath.Join(os.Getenv("HOME"), ".config", "opencode")
	}
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")

	if _, err := os.Stat(manifest); err == nil {
		return CheckResult{Name: "Install Manifest", Passed: true, Message: "kit is installed (manifest found)"}
	}
	return CheckResult{Name: "Install Manifest", Passed: false, Message: "no install manifest found (run 'ntech-team-kit install')"}
}

// checkKitContents uses the strict ValidateKitRoot to surface packaging / layout problems.
func checkKitContents(root string) CheckResult {
	if err := ValidateKitRoot(root); err != nil {
		// Truncate long error for the one-line doctor output
		msg := err.Error()
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		return CheckResult{Name: "Kit Contents", Passed: false, Message: msg}
	}
	return CheckResult{Name: "Kit Contents", Passed: true, Message: "skills/ + VERSION present and valid" }
}
