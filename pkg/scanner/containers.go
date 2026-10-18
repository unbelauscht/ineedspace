package scanner

import (
	"os"
	"time"

	"ineedspace/pkg/model"
)

// ScanContainers scans for Podman, Docker, and VM disk images.
func ScanContainers() []model.Component {
	var components []model.Component

	// 1. Podman Machines (applehv / qemu VMs)
	podmanMachineDir := ExpandPath("~/.local/share/containers/podman/machine")
	if DirExists(podmanMachineDir) {
		sz, count, _ := DirSize(podmanMachineDir)
		if sz > 50*1024*1024 { // > 50MB
			components = append(components, model.Component{
				ID:          "containers-podman-machine",
				Name:        "Podman VM Virtual Disks",
				Category:    model.CategoryContainers,
				Description: "Podman machine VM images (.raw) in ~/.local/share/containers/podman/machine",
				Path:        podmanMachineDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Deletes all Podman Linux VMs. Recreate anytime with 'podman machine init && podman machine start'.",
				CleanAction: "Remove Podman VM disks (podman machine rm -f / delete directory)",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("podman") {
						_, _ = RunCmdWithTimeout(15*time.Second, "podman", "machine", "stop")
						_, _ = RunCmdWithTimeout(15*time.Second, "podman", "machine", "rm", "-f", "podman-machine-default")
					}
					err := os.RemoveAll(podmanMachineDir)
					return sz, err
				},
			})
		}
	}

	// 2. Podman Rootless Storage / Image Cache
	podmanStorageDir := ExpandPath("~/.local/share/containers/storage")
	if DirExists(podmanStorageDir) {
		sz, count, _ := DirSize(podmanStorageDir)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "containers-podman-storage",
				Name:        "Podman Local Storage & Layers",
				Category:    model.CategoryContainers,
				Description: "Locally pulled container images and layers in ~/.local/share/containers/storage",
				Path:        podmanStorageDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Cached container images will be removed; will re-pull images when needed.",
				CleanAction: "Delete ~/.local/share/containers/storage",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(podmanStorageDir)
				},
			})
		}
	}

	// 3. Docker Desktop Virtual Disk & Data
	dockerDataDir := ExpandPath("~/Library/Containers/com.docker.docker/Data")
	if DirExists(dockerDataDir) {
		sz, count, _ := DirSize(dockerDataDir)
		if sz > 100*1024*1024 {
			components = append(components, model.Component{
				ID:          "containers-docker-data",
				Name:        "Docker Desktop Virtual Disk & Cache",
				Category:    model.CategoryContainers,
				Description: "Docker Desktop VM disk (Docker.raw) and build cache in ~/Library/Containers/com.docker.docker",
				Path:        dockerDataDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Deletes Docker Desktop VM disk and local image cache.",
				CleanAction: "Prune Docker system or clean Docker Desktop Data directory",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("docker") {
						_, _ = RunCmdWithTimeout(30*time.Second, "docker", "system", "prune", "-a", "--volumes", "-f")
					}
					// If still huge and docker not running, clean data dir
					return sz, nil
				},
			})
		}
	}

	// 4. Colima VM Disks
	colimaDir := ExpandPath("~/.colima")
	if DirExists(colimaDir) {
		sz, count, _ := DirSize(colimaDir)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "containers-colima",
				Name:        "Colima Container VM",
				Category:    model.CategoryContainers,
				Description: "Colima VM instances and disk images in ~/.colima",
				Path:        colimaDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Colima VM will be deleted. Recreate with 'colima start'.",
				CleanAction: "colima delete -f / remove ~/.colima",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("colima") {
						_, _ = RunCmdWithTimeout(15*time.Second, "colima", "delete", "-f")
					}
					return sz, os.RemoveAll(colimaDir)
				},
			})
		}
	}

	// 5. Rancher Desktop
	rancherDir := ExpandPath("~/.rd")
	if DirExists(rancherDir) {
		sz, count, _ := DirSize(rancherDir)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "containers-rancher",
				Name:        "Rancher Desktop Storage",
				Category:    model.CategoryContainers,
				Description: "Rancher Desktop virtual machine disk and cache in ~/.rd",
				Path:        rancherDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Deletes Rancher Desktop VM disk.",
				CleanAction: "Delete ~/.rd",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(rancherDir)
				},
			})
		}
	}

	// 6. Docker Builder & Buildx Cache
	dockerBuildxDir := ExpandPath("~/.docker/buildx")
	var dockerBuildxSize int64
	var dockerBuildxCount int
	if DirExists(dockerBuildxDir) {
		s, c, _ := DirSize(dockerBuildxDir)
		dockerBuildxSize += s
		dockerBuildxCount += c
	}
	if dockerBuildxSize > 2*1024*1024 {
		components = append(components, model.Component{
			ID:          "containers-docker-buildx",
			Name:        "Docker Builder & Buildx Cache",
			Category:    model.CategoryContainers,
			Description: "Docker builder cache layers and buildx metadata in ~/.docker/buildx",
			Path:        dockerBuildxDir,
			SizeBytes:   dockerBuildxSize,
			ItemCount:   dockerBuildxCount,
			Risk:        model.RiskSafe,
			Impact:      "Intermediate Docker builder cache layers will be purged. Local images and containers are unaffected.",
			CleanAction: "docker builder prune -f / clean ~/.docker/buildx",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return dockerBuildxSize, nil
				}
				if CommandExists("docker") {
					_, _ = RunCmdWithTimeout(25*time.Second, "docker", "builder", "prune", "-f")
				}
				_ = os.RemoveAll(dockerBuildxDir)
				return dockerBuildxSize, nil
			},
		})
	}

	return components
}
