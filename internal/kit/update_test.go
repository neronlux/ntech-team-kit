package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsStale_NoFile(t *testing.T) {
	if !IsStale("/tmp/ntech-team-kit-nonexistent-stamp-test") {
		t.Error("IsStale should return true when file does not exist")
	}
}

func TestIsStale_RecentFile(t *testing.T) {
	dir := t.TempDir()
	stamp := filepath.Join(dir, "stamp")
	if err := os.WriteFile(stamp, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	if IsStale(stamp) {
		t.Error("IsStale should return false for a freshly created file")
	}
}

func TestTouchStamp(t *testing.T) {
	dir := t.TempDir()
	stamp := filepath.Join(dir, "subdir", "stamp")

	if err := TouchStamp(stamp); err != nil {
		t.Fatalf("TouchStamp failed: %v", err)
	}
	if _, err := os.Stat(stamp); err != nil {
		t.Errorf("stamp file not created: %v", err)
	}
}

func TestDefaultUpdateStampPath(t *testing.T) {
	p := DefaultUpdateStampPath()
	if !strings.Contains(p, "ntech-team-kit") {
		t.Errorf("stamp path should contain ntech-team-kit, got %s", p)
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "1.0.0", false},
		{"v1.0.0", "1.0.1", true},
		{"1.0.0", "v1.0.1", true},
		{"v1.0.0", "v1.0.0", false},
		{"1.0.0", "", false},
		{"dev", "1.0.1", true},
		{"", "1.0.1", true},
	}
	for _, tt := range tests {
		got := IsNewerVersion(tt.current, tt.latest)
		if got != tt.want {
			t.Errorf("IsNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}
