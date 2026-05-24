package kit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

	rel, err := fetchLatestRelease()
	if err != nil {
		return "", false, err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	if IsNewerVersion(currentVersion, latest) {
		return latest, true, nil
	}
	return latest, false, nil
}

// IsNewerVersion returns true if latest differs from current after normalizing "v" prefixes.
func IsNewerVersion(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	return latest != "" && latest != current
}

// GetLatestVersion fetches just the tag (for doctor/version banners).
func GetLatestVersion() (string, error) {
	rel, err := fetchLatestRelease()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: defaultTimeout}
	req, err := http.NewRequest("GET", releasesAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ntech-team-kit")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
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
