package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"ineedspace/pkg/model"
)

func TestScanDevToolsDoesNotPanic(t *testing.T) {
	components := ScanDevTools()
	for _, c := range components {
		if c.ID == "" {
			t.Errorf("Component missing ID: %+v", c)
		}
		if c.Category != model.CategoryDevTools && c.Category != model.CategorySDKs {
			t.Errorf("Unexpected category %s for component %s", c.Category, c.ID)
		}
	}
}

func TestDevToolsScanners(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("PIPX_HOME", filepath.Join(tmpHome, ".local", "pipx"))

	// Create dummy files for various dev toolchains
	testDirs := []struct {
		relPath string
		sizeMB  int
	}{
		{"Library/Caches/pip/http", 12},
		{".local/pipx/.cache/venv", 6},
		{".pnpm-store/v3", 22},
		{".yarn/cache", 22},
		{".bun/install/cache", 12},
		{".cache/pypoetry/cache", 12},
		{".cache/pipenv/cache", 12},
		{".conda/pkgs/cache", 55},
		{".rustup/downloads", 12},
		{".m2/repository/com", 55},
		{".bundle/cache", 12},
		{".cache/vcpkg/archives", 55},
		{"Library/Caches/CocoaPods", 22},
		{"Library/Caches/org.swift.swiftpm", 12},
	}

	for _, td := range testDirs {
		p := filepath.Join(tmpHome, td.relPath)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", p, err)
		}
		data := make([]byte, td.sizeMB*1024*1024)
		if err := os.WriteFile(filepath.Join(p, "dummy.dat"), data, 0644); err != nil {
			t.Fatalf("Failed to write dummy file in %s: %v", p, err)
		}
	}

	components := ScanDevTools()
	detected := make(map[string]model.Component)
	for _, c := range components {
		detected[c.ID] = c
	}

	expectedIDs := []string{
		"dev-pip-cache",
		"dev-pipx-cache",
		"dev-pnpm-store",
		"dev-yarn-cache",
		"dev-bun-cache",
		"dev-poetry-cache",
		"dev-pipenv-cache",
		"dev-conda-pkgs",
		"dev-rustup-cache",
		"dev-maven-cache",
		"dev-ruby-cache",
		"dev-cpp-cache",
		"dev-cocoapods-cache",
		"dev-swiftpm-cache",
	}

	for _, id := range expectedIDs {
		c, ok := detected[id]
		if !ok {
			t.Errorf("Expected to detect %s, but it was not found", id)
			continue
		}
		if c.Risk != model.RiskSafe {
			t.Errorf("Component %s risk expected Safe, got %v", id, c.Risk)
		}
		if c.Cleaner == nil {
			t.Errorf("Component %s missing Cleaner function", id)
			continue
		}
		// Test dry-run cleaner
		freed, err := c.Cleaner(true)
		if err != nil {
			t.Errorf("Component %s dry-run cleaner error: %v", id, err)
		}
		if freed != c.SizeBytes {
			t.Errorf("Component %s dry-run freed mismatch: got %d, expected %d", id, freed, c.SizeBytes)
		}
	}
}
