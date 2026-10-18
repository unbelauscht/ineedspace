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
