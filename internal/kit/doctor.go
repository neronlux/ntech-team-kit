package kit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type CheckResult struct {
	Name     string
	Passed   bool
	Message  string
	Optional bool
}

func RunDoctor(kitRoot string) []CheckResult {
	return RunDoctorWithDirs(kitRoot, ConfigDir(), CodexSkillsDir(), CodexAgentsDir())
}

func RunDoctorWithDir(kitRoot string, ocDir string) []CheckResult {
	return RunDoctorWithDirs(kitRoot, ocDir, CodexSkillsDir(), CodexAgentsDir())
}

func RunDoctorWithDirs(kitRoot string, ocDir string, codexSkillsDir string, codexAgentsDir string) []CheckResult {
	var results []CheckResult

	openCode := checkOpenCode()
	codexCLI := checkCodexCLI()
	codexGUI := checkCodexGUI()

	results = append(results, openCode)
	results = append(results, codexCLI)
	results = append(results, codexGUI)
	results = append(results, checkSupportedRuntime(openCode, codexCLI, codexGUI))
	results = append(results, checkGhCLI())
	results = append(results, checkGhAuth())
	results = append(results, checkKitRoot(kitRoot))
	results = append(results, checkInstallManifests(ocDir, codexSkillsDir, codexAgentsDir))
	results = append(results, checkKitContents(kitRoot))

	return results
}

func checkOpenCode() CheckResult {
	return checkOpenCodeWith(exec.LookPath)
}

func checkOpenCodeWith(lookPath func(string) (string, error)) CheckResult {
	paths := []string{"opencode", "opencode-ai"}
	for _, p := range paths {
		if found, err := lookPath(p); err == nil {
			return CheckResult{Name: "OpenCode CLI", Passed: true, Message: "found at " + found, Optional: true}
		}
	}
	return CheckResult{Name: "OpenCode CLI", Passed: false, Message: "not found in PATH. Install from https://opencode.ai", Optional: true}
}

func checkCodexCLI() CheckResult {
	home, _ := os.UserHomeDir()
	return checkCodexCLIWith(exec.LookPath, pathExists, os.Getenv, home, runtime.GOOS)
}

func checkCodexCLIWith(lookPath func(string) (string, error), exists func(string) bool, getenv func(string) string, home string, goos string) CheckResult {
	if envPath := getenv("CODEX_CLI_PATH"); envPath != "" && exists(envPath) {
		return CheckResult{Name: "Codex CLI", Passed: true, Message: "found at " + envPath, Optional: true}
	}
	if found, err := lookPath("codex"); err == nil {
		return CheckResult{Name: "Codex CLI", Passed: true, Message: "found at " + found, Optional: true}
	}
	for _, path := range codexBundledCLIPaths(home, goos) {
		if exists(path) {
			return CheckResult{Name: "Codex CLI", Passed: true, Message: "found bundled at " + path, Optional: true}
		}
	}
	return CheckResult{Name: "Codex CLI", Passed: false, Message: "not found. Install with npm install -g @openai/codex or install the Codex app.", Optional: true}
}

func checkCodexGUI() CheckResult {
	home, _ := os.UserHomeDir()
	return checkCodexGUIWith(pathExists, globPaths, home, runtime.GOOS)
}

func checkCodexGUIWith(exists func(string) bool, glob func(string) ([]string, error), home string, goos string) CheckResult {
	for _, path := range codexGUIPaths(home, goos) {
		if exists(path) {
			return CheckResult{Name: "Codex GUI", Passed: true, Message: "found at " + path, Optional: true}
		}
	}
	for _, pattern := range codexGUIGlobs(home, goos) {
		matches, err := glob(pattern)
		if err == nil && len(matches) > 0 {
			return CheckResult{Name: "Codex GUI", Passed: true, Message: "found at " + matches[0], Optional: true}
		}
	}
	if goos != "darwin" && goos != "linux" {
		return CheckResult{Name: "Codex GUI", Passed: false, Message: "auto-detection supports macOS and Linux", Optional: true}
	}
	return CheckResult{Name: "Codex GUI", Passed: false, Message: "not found in common " + goos + " locations", Optional: true}
}

func checkSupportedRuntime(results ...CheckResult) CheckResult {
	var found []string
	for _, result := range results {
		if result.Passed {
			found = append(found, result.Name)
		}
	}
	if len(found) > 0 {
		return CheckResult{Name: "Agent Runtime", Passed: true, Message: "found " + strings.Join(found, ", ")}
	}
	return CheckResult{Name: "Agent Runtime", Passed: false, Message: "OpenCode or Codex is required"}
}

func codexBundledCLIPaths(home string, goos string) []string {
	if goos != "darwin" {
		return nil
	}
	return []string{
		filepath.Join(home, "Applications", "Codex.app", "Contents", "Resources", "codex"),
		"/Applications/Codex.app/Contents/Resources/codex",
		filepath.Join(home, "Applications", "CodexBar.app", "Contents", "Resources", "codex"),
		"/Applications/CodexBar.app/Contents/Resources/codex",
	}
}

func codexGUIPaths(home string, goos string) []string {
	switch goos {
	case "darwin":
		return []string{
			filepath.Join(home, "Applications", "Codex.app"),
			"/Applications/Codex.app",
			filepath.Join(home, "Applications", "CodexBar.app"),
			"/Applications/CodexBar.app",
		}
	case "linux":
		return []string{
			filepath.Join(home, ".local", "share", "applications", "codex.desktop"),
			"/usr/local/share/applications/codex.desktop",
			"/usr/share/applications/codex.desktop",
			"/var/lib/flatpak/exports/share/applications/com.openai.codex.desktop",
			filepath.Join(home, ".local", "bin", "codex-desktop"),
			"/opt/Codex/codex",
			"/opt/codex/codex",
		}
	default:
		return nil
	}
}

func codexGUIGlobs(home string, goos string) []string {
	if goos != "linux" {
		return nil
	}
	return []string{
		filepath.Join(home, ".local", "share", "applications", "*codex*.desktop"),
		"/usr/local/share/applications/*codex*.desktop",
		"/usr/share/applications/*codex*.desktop",
		filepath.Join(home, "Applications", "*Codex*.AppImage"),
		filepath.Join(home, "Applications", "*codex*.AppImage"),
		filepath.Join(home, "Downloads", "*Codex*.AppImage"),
		filepath.Join(home, "Downloads", "*codex*.AppImage"),
		"/opt/*Codex*/*.AppImage",
		"/opt/*codex*/*.AppImage",
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func globPaths(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func checkGhCLI() CheckResult {
	return checkGhCLIWith(exec.LookPath)
}

func checkGhCLIWith(lookPath func(string) (string, error)) CheckResult {
	if _, err := lookPath("gh"); err == nil {
		return CheckResult{Name: "GitHub CLI (gh)", Passed: true, Message: "found in PATH", Optional: true}
	}
	return CheckResult{Name: "GitHub CLI (gh)", Passed: false, Message: "gh not found. Required for GitHub-dependent skills.", Optional: true}
}

func checkGhAuth() CheckResult {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()
	return checkGhAuthResult(output, err)
}

func checkGhAuthResult(output []byte, err error) CheckResult {
	if err != nil {
		return CheckResult{Name: "gh auth", Passed: false, Message: "not authenticated. Run 'gh auth login' before GitHub-dependent skills.", Optional: true}
	}
	if strings.Contains(string(output), "Logged in") {
		return CheckResult{Name: "gh auth", Passed: true, Message: "authenticated", Optional: true}
	}
	return CheckResult{Name: "gh auth", Passed: false, Message: "authentication issue. Run 'gh auth status' for details.", Optional: true}
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

func checkInstallManifests(ocDir string, codexSkillsDir string, codexAgentsDir string) CheckResult {
	var installed []string
	if count, ok := openCodeManifestCount(ocDir); ok {
		installed = append(installed, fmt.Sprintf("OpenCode: %d files", count))
	}
	if count, ok := codexManifestCount(codexSkillsDir, codexAgentsDir); ok {
		installed = append(installed, fmt.Sprintf("Codex: %d files", count))
	}
	if len(installed) > 0 {
		return CheckResult{Name: "Install Manifest", Passed: true, Message: strings.Join(installed, "; ")}
	}
	return CheckResult{Name: "Install Manifest", Passed: false, Message: "no OpenCode or Codex install manifest found (run 'ntech-team-kit install')"}
}

func openCodeManifestCount(ocDir string) (int, bool) {
	manifest := filepath.Join(ocDir, ".ntech-team-kit-manifest")
	entries, err := readManifest(manifest, ocDir)
	if err != nil {
		return 0, false
	}
	return len(filterOwnedManifestEntries(ocDir, entries)), true
}

func codexManifestCount(skillsDir string, agentsDir string) (int, bool) {
	manifest := codexManifestPath(skillsDir)
	entries, err := readManifest(manifest, filepath.Dir(skillsDir))
	if err != nil {
		return 0, false
	}
	return len(filterOwnedCodexManifestEntries(skillsDir, agentsDir, entries)), true
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
	return CheckResult{Name: "Kit Contents", Passed: true, Message: "skills/ + VERSION present and valid"}
}
