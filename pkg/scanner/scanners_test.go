package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"ineedspace/pkg/model"
)

func TestScanSystemAndAIAndContainers(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create simulated simulator cache
	simPath := filepath.Join(tmpHome, "Library", "Developer", "CoreSimulator", "Caches")
	_ = os.MkdirAll(simPath, 0755)
	_ = os.WriteFile(filepath.Join(simPath, "cache.bin"), make([]byte, 12*1024*1024), 0644)

	// Create simulated torch cache
	torchPath := filepath.Join(tmpHome, ".cache", "torch", "hub")
	_ = os.MkdirAll(torchPath, 0755)
	_ = os.WriteFile(filepath.Join(torchPath, "model.pt"), make([]byte, 12*1024*1024), 0644)

	// Create simulated docker buildx cache
	dockerPath := filepath.Join(tmpHome, ".docker", "buildx", "refs")
	_ = os.MkdirAll(dockerPath, 0755)
	_ = os.WriteFile(filepath.Join(dockerPath, "ref.dat"), make([]byte, 4*1024*1024), 0644)

	sysComps := ScanSystem()
	var foundSim bool
	for _, c := range sysComps {
		if c.ID == "sys-xcode-coresimulator" {
			foundSim = true
			if c.Risk != model.RiskSafe {
				t.Errorf("Expected RiskSafe for coresimulator, got %v", c.Risk)
			}
		}
	}
	if !foundSim {
		t.Errorf("Expected to find sys-xcode-coresimulator")
	}

	aiComps := ScanAI()
	var foundTorch bool
	for _, c := range aiComps {
		if c.ID == "ai-torch" {
			foundTorch = true
			if c.Risk != model.RiskSafe {
				t.Errorf("Expected RiskSafe for torch, got %v", c.Risk)
			}
		}
	}
	if !foundTorch {
		t.Errorf("Expected to find ai-torch")
	}

	containerComps := ScanContainers()
	var foundDocker bool
	for _, c := range containerComps {
		if c.ID == "containers-docker-buildx" {
			foundDocker = true
			if c.Risk != model.RiskSafe {
				t.Errorf("Expected RiskSafe for docker-buildx, got %v", c.Risk)
			}
		}
	}
	if !foundDocker {
		t.Errorf("Expected to find containers-docker-buildx")
	}
}

func TestScanTrashAndClean(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	trashDir := filepath.Join(tmpHome, ".Trash")
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		t.Fatalf("Failed to create mock .Trash: %v", err)
	}

	testFile := filepath.Join(trashDir, "trash_item.bin")
	if err := os.WriteFile(testFile, make([]byte, 5*1024*1024), 0644); err != nil {
		t.Fatalf("Failed to write mock trash item: %v", err)
	}

	comps := ScanSystem()
	var trashComp *model.Component
	for _, c := range comps {
		if c.ID == "sys-trash" {
			trashComp = &c
			break
		}
	}

	if trashComp == nil {
		t.Fatalf("Expected sys-trash to be detected in simulated home")
	}

	if trashComp.SizeBytes != 5*1024*1024 {
		t.Errorf("Expected 5MB size, got %d", trashComp.SizeBytes)
	}

	// Test dry run cleaner
	freedDry, err := trashComp.Cleaner(true)
	if err != nil {
		t.Errorf("Unexpected error during dry run: %v", err)
	}
	if freedDry != 5*1024*1024 {
		t.Errorf("Expected dry run freed to match size, got %d", freedDry)
	}

	// File should still exist after dry run
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("Expected file to exist after dry run")
	}

	// Test real cleaner
	freedReal, err := trashComp.Cleaner(false)
	if err != nil {
		t.Errorf("Unexpected error during real clean: %v", err)
	}
	if freedReal != 5*1024*1024 {
		t.Errorf("Expected real clean freed to match size, got %d", freedReal)
	}

	// File should be deleted after real clean
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("Expected file to be deleted after real clean, stat err: %v", err)
	}
}
