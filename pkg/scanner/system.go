package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ineedspace/pkg/model"
)

// ScanSystem scans for system caches, package manager caches, and large app caches.
func ScanSystem() []model.Component {
	var components []model.Component

	// 1. Homebrew Download Cache
	brewCache := ExpandPath("~/Library/Caches/Homebrew")
	if DirExists(brewCache) {
		sz, count, _ := DirSize(brewCache)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-brew-cache",
				Name:        "Homebrew Download Cache",
				Category:    model.CategoryPackageMgr,
				Description: "Cached formula bottles and git clones in ~/Library/Caches/Homebrew",
				Path:        brewCache,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Cached bottle archives will be deleted. Will re-download on next 'brew install'.",
				CleanAction: "brew cleanup -s --prune=all / delete cache",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("brew") {
						_, _ = RunCmdWithTimeout(20*time.Second, "brew", "cleanup", "-s", "--prune=all")
					}
					return sz, os.RemoveAll(brewCache)
				},
			})
		}
	}

	// 2. Xcode DerivedData
	derivedData := ExpandPath("~/Library/Developer/Xcode/DerivedData")
	if DirExists(derivedData) {
		sz, count, _ := DirSize(derivedData)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-xcode-deriveddata",
				Name:        "Xcode DerivedData",
				Category:    model.CategoryDevTools,
				Description: "Xcode intermediate build products, indexes and logs",
				Path:        derivedData,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Deletes intermediate build artifacts. Xcode will re-index on next project open.",
				CleanAction: "Delete ~/Library/Developer/Xcode/DerivedData",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(derivedData)
				},
			})
		}
	}

	// 3. Xcode iOS Device Support
	deviceSupport := ExpandPath("~/Library/Developer/Xcode/iOS DeviceSupport")
	if DirExists(deviceSupport) {
		sz, count, _ := DirSize(deviceSupport)
		if sz > 100*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-xcode-devicesupport",
				Name:        "Xcode iOS Device Support Symbols",
				Category:    model.CategoryDevTools,
				Description: "Debug symbols for previously connected physical iOS devices",
				Path:        deviceSupport,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Debug symbols for old devices will be deleted. Re-created automatically when debugging.",
				CleanAction: "Delete ~/Library/Developer/Xcode/iOS DeviceSupport",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(deviceSupport)
				},
			})
		}
	}

	// 4. Xcode CoreSimulator Caches & Unavailable Simulators
	simCaches := ExpandPath("~/Library/Developer/CoreSimulator/Caches")
	var simTotal int64
	var simCount int
	if DirExists(simCaches) {
		s, c, _ := DirSize(simCaches)
		simTotal += s
		simCount += c
	}
	if simTotal > 10*1024*1024 {
		components = append(components, model.Component{
			ID:          "sys-xcode-coresimulator",
			Name:        "Xcode Simulator Caches & Unavailable Devices",
			Category:    model.CategoryDevTools,
			Description: "CoreSimulator caches in ~/Library/Developer/CoreSimulator/Caches and defunct simulator runtimes",
			Path:        simCaches,
			SizeBytes:   simTotal,
			ItemCount:   simCount,
			Risk:        model.RiskSafe,
			Impact:      "Deletes simulator caches and removes unusable old simulator instances. Current simulators are preserved.",
			CleanAction: "xcrun simctl delete unavailable / delete simulator caches",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return simTotal, nil
				}
				if CommandExists("xcrun") {
					_, _ = RunCmdWithTimeout(20*time.Second, "xcrun", "simctl", "delete", "unavailable")
				}
				_ = os.RemoveAll(simCaches)
				return simTotal, nil
			},
		})
	}

	// 5. Trash
	trashPath := ExpandPath("~/.Trash")
	if DirExists(trashPath) {
		sz, count, _ := DirSize(trashPath)
		if sz > 10*1024*1024 {
			components = append(components, model.Component{
				ID:          "sys-trash",
				Name:        "macOS Trash Bin",
				Category:    model.CategorySystem,
				Description: "Files currently in your Trash (~/.Trash)",
				Path:        trashPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskCaution,
				Impact:      "Permanently empties your Trash bin.",
				CleanAction: "Empty Trash",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					// Empty contents inside ~/.Trash
					entries, err := os.ReadDir(trashPath)
					if err != nil {
						return 0, err
					}
					for _, e := range entries {
						_ = os.RemoveAll(filepath.Join(trashPath, e.Name()))
					}
					return sz, nil
				},
			})
		}
	}

	// 6. Heavy App Caches in ~/Library/Caches (> 300MB)
	userCaches := ExpandPath("~/Library/Caches")
	if DirExists(userCaches) {
		entries, err := os.ReadDir(userCaches)
		if err == nil {
			// Skip caches handled in dedicated components
			skipCaches := map[string]bool{
				"Homebrew":          true,
				"go-build":          true,
				"pip":               true,
				"pipx":              true,
				"pypoetry":          true,
				"pipenv":            true,
				"Yarn":              true,
				"CocoaPods":         true,
				"org.swift.swiftpm": true,
			}
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()
					if skipCaches[name] {
						continue
					}
					target := filepath.Join(userCaches, name)
					sz, count, _ := DirSize(target)
					if sz > 300*1024*1024 { // Flag any individual app cache larger than 300MB
						components = append(components, model.Component{
							ID:          "cache-" + name,
							Name:        fmt.Sprintf("App Cache: %s", name),
							Category:    model.CategorySystem,
							Description: fmt.Sprintf("Cached files for %s in ~/Library/Caches", name),
							Path:        target,
							SizeBytes:   sz,
							ItemCount:   count,
							Risk:        model.RiskSafe,
							Impact:      fmt.Sprintf("Temporary cache for %s will be wiped. App will re-cache as needed.", name),
							CleanAction: fmt.Sprintf("Delete ~/Library/Caches/%s", name),
							Cleaner: func(dryRun bool) (int64, error) {
								if dryRun {
									return sz, nil
								}
								return sz, os.RemoveAll(target)
							},
						})
					}
				}
			}
		}
	}

	return components
}
