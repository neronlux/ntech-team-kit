package kit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	githubRepo     = "neronlux/ntech-team-kit"
	releasesAPI    = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	releasesPage   = "https://github.com/" + githubRepo + "/releases/latest"
	defaultTimeout = 5 * time.Second
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdate queries GitHub for the latest release and compares with current version.
// Returns (latestVersion, updateAvailable, err).
func CheckForUpdate(currentVersion string) (string, bool, error) {
	if currentVersion == "" || currentVersion == "dev" {
		return "", false, nil
	}

	client := &http.Client{Timeout: defaultTimeout}
	req, err := http.NewRequest("GET", releasesAPI, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ntech-team-kit/"+currentVersion)

	resp, err := client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", false, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	var rel githubRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", false, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest == "" || latest == current {
		return latest, false, nil
	}

	// Simple semver-ish compare: if latest != current, treat as newer (GitHub is source of truth)
	return latest, true, nil
}

// PerformUpdate attempts to update based on installation method.
// Returns a human-readable message about what happened or what the user should do.
func PerformUpdate(currentVersion string, yes bool) (string, error) {
	latest, available, err := CheckForUpdate(currentVersion)
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}
	if !available {
		return fmt.Sprintf("Already up to date (v%s).", currentVersion), nil
	}

	// Detect install method
	if isHomebrewInstall() {
		return runBrewUpgrade(yes, latest)
	}

	// Fallback: tell user to run the appropriate command
	return fmt.Sprintf("New version v%s available.\n\nUpdate with:\n  brew upgrade ntech-team-kit\n\nOr visit: %s", latest, releasesPage), nil
}

func isHomebrewInstall() bool {
	// Check if binary path contains Homebrew prefix
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(exe, "/opt/homebrew/") || strings.Contains(exe, "/usr/local/Homebrew/")
}

func runBrewUpgrade(yes bool, latest string) (string, error) {
	if !yes {
		return fmt.Sprintf("New version v%s available.\n\nRun: brew upgrade ntech-team-kit\nOr: brew upgrade && brew cleanup", latest), nil
	}

	cmd := exec.Command("brew", "upgrade", "ntech-team-kit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("brew upgrade failed: %w", err)
	}
	return "Updated via Homebrew.", nil
}

// GetLatestVersion fetches just the tag (for doctor/version banners).
func GetLatestVersion() (string, error) {
	client := &http.Client{Timeout: defaultTimeout}
	req, _ := http.NewRequest("GET", releasesAPI, nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ntech-team-kit")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

// IsStale checks a local timestamp file; returns true if >24h since last check.
func IsStale(stampPath string) bool {
	fi, err := os.Stat(stampPath)
	if err != nil {
		return true
	}
	return time.Since(fi.ModTime()) > 24*time.Hour
}

// TouchStamp updates the last-check timestamp.
func TouchStamp(stampPath string) error {
	dir := filepath.Dir(stampPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(stampPath)
	if err != nil {
		return err
	}
	return f.Close()
}

// DefaultUpdateStampPath returns a stable per-user location for the last-update-check timestamp.
func DefaultUpdateStampPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".cache", "ntech-team-kit", "last-update-check")
}
