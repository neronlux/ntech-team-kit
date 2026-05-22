package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// defaultKitRoot is set at build time via -ldflags for Homebrew installs.
// Example: -X github.com/neronlux/ntech-team-kit/internal/kit.defaultKitRoot=/opt/homebrew/opt/ntech-team-kit/libexec
var defaultKitRoot string

// GetKitRoot returns the absolute path to the installed ntech-team-kit directory.
// Priority:
//  1. NTECH_TEAM_KIT_ROOT environment variable
//  2. Compiled default (set by Homebrew via ldflags)
//  3. Auto-detection by walking up from the current binary or working directory
func GetKitRoot() string {
	if env := os.Getenv("NTECH_TEAM_KIT_ROOT"); env != "" {
		return env
	}
	if defaultKitRoot != "" {
		return defaultKitRoot
	}
	return detectRepoRoot()
}

// detectRepoRoot tries to find the kit root by looking for install.sh + skills/ directory.
func detectRepoRoot() string {
	// Try from the executable location first (best for built binaries)
	if exe, err := os.Executable(); err == nil {
		if root := searchUpForKit(filepath.Dir(exe)); root != "" {
			return root
		}
	}

	// Fallback: search upward from current working directory (good for go run / dev)
	if cwd, err := os.Getwd(); err == nil {
		if root := searchUpForKit(cwd); root != "" {
			return root
		}
	}

	// Last resort: use the directory of this source file (only works when running from source tree)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		if root := searchUpForKit(filepath.Dir(filename)); root != "" {
			return root
		}
	}

	return ""
}

func searchUpForKit(start string) string {
	dir := start
	for {
		if hasKitLayout(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func hasKitLayout(dir string) bool {
	installScript := filepath.Join(dir, "install.sh")
	skillsDir := filepath.Join(dir, "skills")

	if _, err := os.Stat(installScript); err != nil {
		return false
	}
	if info, err := os.Stat(skillsDir); err != nil || !info.IsDir() {
		return false
	}
	return true
}

// ValidateKitRoot performs a strict pre-flight check on the resolved kit root.
// It returns a clear, actionable error if the directory does not look like a
// complete ntech-team-kit installation (missing install.sh, skills/, or key files).
// This prevents cryptic failures later in install.sh (e.g. "cp: ... No such file").
func ValidateKitRoot(root string) error {
	if root == "" {
		return fmt.Errorf("could not determine kit root (set NTECH_TEAM_KIT_ROOT or use --root)")
	}

	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("kit root does not exist: %s\n\nTry setting NTECH_TEAM_KIT_ROOT or --root to the correct path", root)
	}

	if !hasKitLayout(root) {
		return fmt.Errorf("kit root is missing required files (install.sh + skills/ directory):\n  %s\n\nThis usually means the installation is incomplete or corrupted.\n\nSuggestions:\n  • Homebrew users:   brew reinstall ntech-team-kit\n  • Source users:     cd /path/to/clone && git pull && ./bin/ntech-team-kit install\n  • Override:         NTECH_TEAM_KIT_ROOT=/path/to/kit ntech-team-kit install", root)
	}

	// Extra sanity: verify at least one real skill file exists (catches partial packaging)
	exampleSkill := filepath.Join(root, "skills", "check-compiler-errors", "SKILL.md")
	if _, err := os.Stat(exampleSkill); err != nil {
		return fmt.Errorf("kit root looks incomplete — expected skill file is missing:\n  %s\n\nThe packaged kit appears to be missing content. Please reinstall:\n  • Homebrew: brew reinstall ntech-team-kit\n  • Source:   git pull && ntech-team-kit install", exampleSkill)
	}

	return nil
}
