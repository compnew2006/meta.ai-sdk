package metaai

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestModuleIsSelfContained(t *testing.T) {
	requiredFiles := []string{
		".env.example",
		".gitattributes",
		".github/workflows/ci.yml",
		".gitignore",
		"LICENSE",
		"Makefile",
	}
	for _, path := range requiredFiles {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("standalone project file %q is unavailable: %v", path, err)
		}
	}
}

func TestModuleHasNoLegacyRuntimeCoupling(t *testing.T) {
	legacyRuntime := regexp.MustCompile(`(?i)` + "py" + `thon|\.py\b|pyproject|requirements\.txt`)
	textExtensions := map[string]bool{
		".go": true, ".html": true, ".md": true, ".mod": true,
		".toml": true, ".yaml": true, ".yml": true,
	}

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			// Skip tooling metadata, VCS, node_modules and built frontends.
			name := entry.Name()
			if path == ".codebase-memory" || path == ".git" || path == ".serena" || name == "node_modules" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if !textExtensions[filepath.Ext(path)] {
			return nil
		}
		if path == "isolation_test.go" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if legacyRuntime.Match(content) {
			t.Errorf("legacy runtime reference found in %s", path)
		}
		if strings.Contains(string(content), "../") {
			t.Errorf("parent-project reference found in %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestModuleDoesNotContainBuiltExecutables(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != "" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			t.Errorf("built executable %q must not be stored in the module", entry.Name())
		}
	}
}
