package kit

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var agentNamePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestInstallListsMatchRepositoryContent(t *testing.T) {
	root := repoRootForTest(t)

	assertSameSetForTest(t, "skills slice vs skills directory", setFromSliceForTest(skills), dirNamesForTest(t, filepath.Join(root, "skills")))
	assertSameSetForTest(t, "commands slice vs commands directory", setFromSliceForTest(commands), markdownNamesForTest(t, filepath.Join(root, "commands")))
	assertSameSetForTest(t, "commands slice vs skills slice", setFromSliceForTest(commands), setFromSliceForTest(skills))
	assertSameSetForTest(t, "agents slice vs agents directory", setFromSliceForTest(agents), markdownNamesForTest(t, filepath.Join(root, "agents")))
}

func TestSkillFilesMeetOpenCodeAndCodexSpec(t *testing.T) {
	root := repoRootForTest(t)
	for _, skill := range skills {
		path := filepath.Join(root, "skills", skill, "SKILL.md")
		frontmatter, _ := readMarkdownFrontmatterForTest(t, path)

		if got := frontmatter["name"]; got != skill {
			t.Fatalf("%s name = %q, want %q", path, got, skill)
		}
		if !agentNamePattern.MatchString(skill) || len(skill) > 64 {
			t.Fatalf("%s does not match OpenCode/Codex skill naming rules", skill)
		}
		description := frontmatter["description"]
		if description == "" || len(description) > 1024 {
			t.Fatalf("%s description length = %d, want 1..1024", skill, len(description))
		}
		if got := frontmatter["compatibility"]; got != "opencode" {
			t.Fatalf("%s source compatibility = %q, want opencode", skill, got)
		}
	}
}

func TestOpenCodeAgentFilesMeetSpec(t *testing.T) {
	root := repoRootForTest(t)
	for _, agent := range agents {
		path := filepath.Join(root, "agents", agent+".md")
		frontmatter, body := readMarkdownFrontmatterForTest(t, path)

		if !agentNamePattern.MatchString(agent) || len(agent) > 64 {
			t.Fatalf("%s does not match agent naming rules", agent)
		}
		if frontmatter["description"] == "" {
			t.Fatalf("%s missing description", path)
		}
		if got := frontmatter["mode"]; got != "subagent" {
			t.Fatalf("%s mode = %q, want subagent", path, got)
		}
		if strings.Contains(body, "tools:") {
			t.Fatalf("%s uses deprecated tools config in body", path)
		}
		if !strings.Contains(string(mustReadForTest(t, path)), "edit: deny") {
			t.Fatalf("%s should explicitly deny edit permission", path)
		}
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func dirNamesForTest(t *testing.T, dir string) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names[entry.Name()] = true
		}
	}
	return names
}

func markdownNamesForTest(t *testing.T, dir string) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".") {
			names[strings.TrimSuffix(name, ".md")] = true
		}
	}
	return names
}

func setFromSliceForTest(values []string) map[string]bool {
	names := map[string]bool{}
	for _, value := range values {
		names[value] = true
	}
	return names
}

func assertSameSetForTest(t *testing.T, label string, want map[string]bool, got map[string]bool) {
	t.Helper()
	var missing []string
	var extra []string
	for name := range want {
		if !got[name] {
			missing = append(missing, name)
		}
	}
	for name := range got {
		if !want[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 || len(extra) > 0 {
		t.Fatalf("%s mismatch: missing %v, extra %v", label, missing, extra)
	}
}

func readMarkdownFrontmatterForTest(t *testing.T, path string) (map[string]string, string) {
	t.Helper()
	data := string(mustReadForTest(t, path))
	if !strings.HasPrefix(data, "---\n") {
		t.Fatalf("%s missing YAML frontmatter", path)
	}
	rest := strings.TrimPrefix(data, "---\n")
	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		t.Fatalf("%s missing closing YAML frontmatter", path)
	}

	frontmatter := map[string]string{}
	for _, line := range strings.Split(parts[0], "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if ok {
			frontmatter[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return frontmatter, parts[1]
}

func mustReadForTest(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
