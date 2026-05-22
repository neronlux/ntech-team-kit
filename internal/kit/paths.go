package kit

import (
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
