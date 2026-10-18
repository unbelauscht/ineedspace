package scanner

import (
	"os"
	"time"

	"ineedspace/pkg/model"
)

// ScanDevTools scans for developer build caches and SDK components.
func ScanDevTools() []model.Component {
	var components []model.Component

	// 1. Python uv cache (~/.cache/uv)
	uvCache := ExpandPath("~/.cache/uv")
	if DirExists(uvCache) {
		sz, count, _ := DirSize(uvCache)
		if sz > 10*1024*1024 {
			components = append(components, model.Component{
				ID:          "dev-uv-cache",
				Name:        "Python uv Cache",
				Category:    model.CategoryDevTools,
				Description: "Fast Python package manager cache in ~/.cache/uv",
				Path:        uvCache,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Cached Python wheels will be cleared; packages will be re-downloaded when installing.",
				CleanAction: "uv cache clean / delete ~/.cache/uv",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("uv") {
						_, _ = RunCmdWithTimeout(15*time.Second, "uv", "cache", "clean")
					}
					return sz, os.RemoveAll(uvCache)
				},
			})
		}
	}

	// 2. Go Build & Module Cache
	goBuildCache := ExpandPath("~/Library/Caches/go-build")
	goModCache := ExpandPath("~/go/pkg/mod")
	var goTotalSize int64
	var goTotalFiles int
	if DirExists(goBuildCache) {
		s, c, _ := DirSize(goBuildCache)
		goTotalSize += s
		goTotalFiles += c
	}
	if DirExists(goModCache) {
		s, c, _ := DirSize(goModCache)
		goTotalSize += s
		goTotalFiles += c
	}
	if goTotalSize > 50*1024*1024 {
		components = append(components, model.Component{
			ID:          "dev-go-cache",
			Name:        "Go Build & Module Cache",
			Category:    model.CategoryDevTools,
			Description: "Go compiler build cache and downloaded go modules (~/go/pkg/mod)",
			Path:        goBuildCache + ", " + goModCache,
			SizeBytes:   goTotalSize,
			ItemCount:   goTotalFiles,
			Risk:        model.RiskSafe,
			Impact:      "Go build cache and mod cache will be purged. Modules will be re-downloaded on 'go build'.",
			CleanAction: "go clean -cache -modcache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return goTotalSize, nil
				}
				if CommandExists("go") {
					_, _ = RunCmdWithTimeout(30*time.Second, "go", "clean", "-cache", "-modcache")
				}
				_ = os.RemoveAll(goBuildCache)
				_ = os.RemoveAll(goModCache)
				return goTotalSize, nil
			},
		})
	}

	// 3. Node / NPM Cache
	npmCache := ExpandPath("~/.npm")
	if DirExists(npmCache) {
		sz, count, _ := DirSize(npmCache)
		if sz > 20*1024*1024 {
			components = append(components, model.Component{
				ID:          "dev-npm-cache",
				Name:        "NPM Package Cache",
				Category:    model.CategoryDevTools,
				Description: "NPM tarballs and HTTP cache in ~/.npm",
				Path:        npmCache,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Cached npm packages will be purged. Fresh packages download on 'npm install'.",
				CleanAction: "npm cache clean --force",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					if CommandExists("npm") {
						_, _ = RunCmdWithTimeout(20*time.Second, "npm", "cache", "clean", "--force")
					}
					return sz, os.RemoveAll(npmCache)
				},
			})
		}
	}

	// 4. Gradle Build Cache
	gradleCache := ExpandPath("~/.gradle/caches")
	if DirExists(gradleCache) {
		sz, count, _ := DirSize(gradleCache)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "dev-gradle-cache",
				Name:        "Gradle Dependency & Build Cache",
				Category:    model.CategoryDevTools,
				Description: "Downloaded jars, aars and compiled transforms in ~/.gradle/caches",
				Path:        gradleCache,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Gradle dependencies will be deleted. Automatically re-downloaded on next build.",
				CleanAction: "Delete ~/.gradle/caches",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(gradleCache)
				},
			})
		}
	}

	// 5. Cargo / Rust Registry & Git Checkouts
	cargoCache := ExpandPath("~/.cargo/registry/cache")
	cargoGit := ExpandPath("~/.cargo/git/checkouts")
	var cargoTotal int64
	var cargoCount int
	if DirExists(cargoCache) {
		s, c, _ := DirSize(cargoCache)
		cargoTotal += s
		cargoCount += c
	}
	if DirExists(cargoGit) {
		s, c, _ := DirSize(cargoGit)
		cargoTotal += s
		cargoCount += c
	}
	if cargoTotal > 50*1024*1024 {
		components = append(components, model.Component{
			ID:          "dev-cargo-cache",
			Name:        "Cargo Crate Cache & Git Repos",
			Category:    model.CategoryDevTools,
			Description: "Downloaded crate archives (.crate) and cloned git checkouts in ~/.cargo",
			Path:        cargoCache,
			SizeBytes:   cargoTotal,
			ItemCount:   cargoCount,
			Risk:        model.RiskSafe,
			Impact:      "Downloaded .crate files will be removed. Will re-download when building.",
			CleanAction: "Remove ~/.cargo/registry/cache and git checkouts",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return cargoTotal, nil
				}
				_ = os.RemoveAll(cargoCache)
				_ = os.RemoveAll(cargoGit)
				return cargoTotal, nil
			},
		})
	}

	// 6. Kotlin/Native (Konan) Cache
	konanDir := ExpandPath("~/.konan")
	if DirExists(konanDir) {
		sz, count, _ := DirSize(konanDir)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "dev-konan-cache",
				Name:        "Kotlin/Native (Konan) Toolchains",
				Category:    model.CategoryDevTools,
				Description: "Precompiled LLVM toolchains and target dependencies in ~/.konan",
				Path:        konanDir,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Konan dependencies will be deleted. Will re-download on next Kotlin/Native build.",
				CleanAction: "Delete ~/.konan",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(konanDir)
				},
			})
		}
	}

	// 7. PlatformIO Packages & Platforms
	pioPackages := ExpandPath("~/.platformio/packages")
	pioPlatforms := ExpandPath("~/.platformio/platforms")
	var pioTotal int64
	var pioCount int
	if DirExists(pioPackages) {
		s, c, _ := DirSize(pioPackages)
		pioTotal += s
		pioCount += c
	}
	if DirExists(pioPlatforms) {
		s, c, _ := DirSize(pioPlatforms)
		pioTotal += s
		pioCount += c
	}
	if pioTotal > 50*1024*1024 {
		components = append(components, model.Component{
			ID:          "dev-platformio",
			Name:        "PlatformIO Toolchains & Frameworks",
			Category:    model.CategorySDKs,
			Description: "Embedded compilers (toolchain-gccarm, espressif, etc.) in ~/.platformio",
			Path:        pioPackages,
			SizeBytes:   pioTotal,
			ItemCount:   pioCount,
			Risk:        model.RiskNotice,
			Impact:      "Compilers will be deleted. PlatformIO will automatically re-install when opening an embedded project.",
			CleanAction: "Delete ~/.platformio/packages and platforms",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return pioTotal, nil
				}
				_ = os.RemoveAll(pioPackages)
				_ = os.RemoveAll(pioPlatforms)
				return pioTotal, nil
			},
		})
	}

	// 8. Arduino15 Hardware Cores & Tools
	arduinoPath := ExpandPath("~/Library/Arduino15/packages")
	if DirExists(arduinoPath) {
		sz, count, _ := DirSize(arduinoPath)
		if sz > 50*1024*1024 {
			components = append(components, model.Component{
				ID:          "dev-arduino15",
				Name:        "Arduino Hardware Packages & Cores",
				Category:    model.CategorySDKs,
				Description: "Downloaded board support packages (ESP32, AVR, STM32) in ~/Library/Arduino15",
				Path:        arduinoPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Board cores will be deleted. You can re-install needed boards in Arduino IDE.",
				CleanAction: "Delete ~/Library/Arduino15/packages",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(arduinoPath)
				},
			})
		}
	}

	// 9. Android Emulator AVDs & System Images
	androidSystemImages := ExpandPath("~/Library/Android/sdk/system-images")
	androidAVDs := ExpandPath("~/.android/avd")
	var androidTotal int64
	var androidCount int
	if DirExists(androidSystemImages) {
		s, c, _ := DirSize(androidSystemImages)
		androidTotal += s
		androidCount += c
	}
	if DirExists(androidAVDs) {
		s, c, _ := DirSize(androidAVDs)
		androidTotal += s
		androidCount += c
	}
	if androidTotal > 50*1024*1024 {
		components = append(components, model.Component{
			ID:          "dev-android-avd",
			Name:        "Android Emulators & System Images",
			Category:    model.CategorySDKs,
			Description: "Android Virtual Devices (AVD) and downloaded system images",
			Path:        androidSystemImages + ", " + androidAVDs,
			SizeBytes:   androidTotal,
			ItemCount:   androidCount,
			Risk:        model.RiskNotice,
			Impact:      "Virtual emulators will be deleted. Re-download images via Android Studio SDK Manager.",
			CleanAction: "Delete system-images and AVDs",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return androidTotal, nil
				}
				_ = os.RemoveAll(androidSystemImages)
				_ = os.RemoveAll(androidAVDs)
				return androidTotal, nil
			},
		})
	}

	return components
}
