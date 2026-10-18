package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"ineedspace/pkg/model"
)

// ScanDevTools scans for developer build caches and SDK components.
func ScanDevTools() []model.Component {
	var components []model.Component

	// -------------------------------------------------------------
	// Python Ecosystem: uv, pip, pipx, poetry, pipenv, conda
	// -------------------------------------------------------------

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

	// 2. Python pip cache (~/Library/Caches/pip, ~/.cache/pip)
	pipPaths, pipSize, pipCount := AggregatePaths([]string{
		ExpandPath("~/Library/Caches/pip"),
		ExpandPath("~/.cache/pip"),
	})
	if pipSize > 10*1024*1024 {
		pipDisplay := strings.Join(pipPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-pip-cache",
			Name:        "Python pip Cache",
			Category:    model.CategoryDevTools,
			Description: "Cached wheels and HTTP downloads in " + pipDisplay,
			Path:        pipDisplay,
			SizeBytes:   pipSize,
			ItemCount:   pipCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached Python wheels and packages will be cleared. Packages will be re-downloaded when installing.",
			CleanAction: "pip cache purge / delete pip cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return pipSize, nil
				}
				if CommandExists("pip") {
					_, _ = RunCmdWithTimeout(20*time.Second, "pip", "cache", "purge")
				} else if CommandExists("pip3") {
					_, _ = RunCmdWithTimeout(20*time.Second, "pip3", "cache", "purge")
				} else if CommandExists("python3") {
					_, _ = RunCmdWithTimeout(20*time.Second, "python3", "-m", "pip", "cache", "purge")
				}
				for _, p := range pipPaths {
					_ = os.RemoveAll(p)
				}
				return pipSize, nil
			},
		})
	}

	// 3. pipx Ephemeral Cache & Logs
	pipxHome := os.Getenv("PIPX_HOME")
	if pipxHome == "" {
		pipxHome = ExpandPath("~/.local/pipx")
	}
	pipxPaths, pipxSize, pipxCount := AggregatePaths([]string{
		filepath.Join(pipxHome, ".cache"),
		filepath.Join(pipxHome, "trash"),
		filepath.Join(pipxHome, "logs"),
		ExpandPath("~/.cache/pipx"),
		ExpandPath("~/Library/Caches/pipx"),
	})
	if pipxSize > 5*1024*1024 {
		pipxDisplay := strings.Join(pipxPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-pipx-cache",
			Name:        "pipx Ephemeral Cache & Logs",
			Category:    model.CategoryDevTools,
			Description: "Temporary virtual environments from 'pipx run', trash, and logs in " + pipxDisplay,
			Path:        pipxDisplay,
			SizeBytes:   pipxSize,
			ItemCount:   pipxCount,
			Risk:        model.RiskSafe,
			Impact:      "Ephemeral virtualenvs from 'pipx run', uninstalled package trash, and logs will be wiped. Installed CLI tools in venvs/ will NOT be touched.",
			CleanAction: "Delete pipx cache, trash, and logs",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return pipxSize, nil
				}
				for _, p := range pipxPaths {
					_ = os.RemoveAll(p)
				}
				return pipxSize, nil
			},
		})
	}

	// 4. Poetry Cache (~/.cache/pypoetry, ~/Library/Caches/pypoetry)
	poetryPaths, poetrySize, poetryCount := AggregatePaths([]string{
		ExpandPath("~/.cache/pypoetry"),
		ExpandPath("~/Library/Caches/pypoetry"),
	})
	if poetrySize > 10*1024*1024 {
		display := strings.Join(poetryPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-poetry-cache",
			Name:        "Poetry Package & Wheel Cache",
			Category:    model.CategoryDevTools,
			Description: "Cached Python wheels and repository metadata in " + display,
			Path:        display,
			SizeBytes:   poetrySize,
			ItemCount:   poetryCount,
			Risk:        model.RiskSafe,
			Impact:      "Poetry artifact cache will be purged. Wheels will re-download when resolving dependencies.",
			CleanAction: "poetry cache clear --all / delete cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return poetrySize, nil
				}
				if CommandExists("poetry") {
					_, _ = RunCmdWithTimeout(20*time.Second, "poetry", "cache", "clear", "--all", "-n", ".")
				}
				for _, p := range poetryPaths {
					_ = os.RemoveAll(p)
				}
				return poetrySize, nil
			},
		})
	}

	// 5. Pipenv Cache (~/.cache/pipenv, ~/Library/Caches/pipenv)
	pipenvPaths, pipenvSize, pipenvCount := AggregatePaths([]string{
		ExpandPath("~/.cache/pipenv"),
		ExpandPath("~/Library/Caches/pipenv"),
	})
	if pipenvSize > 10*1024*1024 {
		display := strings.Join(pipenvPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-pipenv-cache",
			Name:        "Pipenv Package Cache",
			Category:    model.CategoryDevTools,
			Description: "Cached package archives in " + display,
			Path:        display,
			SizeBytes:   pipenvSize,
			ItemCount:   pipenvCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached wheels will be deleted. Will re-download on 'pipenv install'.",
			CleanAction: "Delete Pipenv cache directory",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return pipenvSize, nil
				}
				for _, p := range pipenvPaths {
					_ = os.RemoveAll(p)
				}
				return pipenvSize, nil
			},
		})
	}

	// 6. Conda / Mamba Package Tarballs & Index
	condaPaths, condaSize, condaCount := AggregatePaths([]string{
		ExpandPath("~/.conda/pkgs"),
		ExpandPath("~/miniconda3/pkgs"),
		ExpandPath("~/anaconda3/pkgs"),
		ExpandPath("~/miniforge3/pkgs"),
	})
	if condaSize > 50*1024*1024 {
		display := strings.Join(condaPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-conda-pkgs",
			Name:        "Conda / Mamba Package Tarballs",
			Category:    model.CategoryDevTools,
			Description: "Downloaded conda package tarballs (.tar.bz2/.conda) and index cache in " + display,
			Path:        display,
			SizeBytes:   condaSize,
			ItemCount:   condaCount,
			Risk:        model.RiskSafe,
			Impact:      "Unused package archives and index cache will be removed. Active conda environments are unaffected.",
			CleanAction: "conda clean -t -i -y / delete pkgs cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return condaSize, nil
				}
				if CommandExists("conda") {
					_, _ = RunCmdWithTimeout(30*time.Second, "conda", "clean", "--tarballs", "--index-cache", "-y")
				} else if CommandExists("mamba") {
					_, _ = RunCmdWithTimeout(30*time.Second, "mamba", "clean", "--tarballs", "--index-cache", "-y")
				}
				for _, p := range condaPaths {
					_ = os.RemoveAll(p)
				}
				return condaSize, nil
			},
		})
	}

	// -------------------------------------------------------------
	// JavaScript & Node Ecosystem: npm, pnpm, yarn, bun
	// -------------------------------------------------------------

	// 7. Node / NPM Cache
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

	// 8. pnpm Global Package Store
	pnpmPaths, pnpmSize, pnpmCount := AggregatePaths([]string{
		ExpandPath("~/.pnpm-store"),
		ExpandPath("~/Library/pnpm/store"),
		ExpandPath("~/.local/share/pnpm/store"),
	})
	if pnpmSize > 20*1024*1024 {
		display := strings.Join(pnpmPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-pnpm-store",
			Name:        "pnpm Package Store",
			Category:    model.CategoryDevTools,
			Description: "Global content-addressable package store in " + display,
			Path:        display,
			SizeBytes:   pnpmSize,
			ItemCount:   pnpmCount,
			Risk:        model.RiskSafe,
			Impact:      "Unreferenced packages will be pruned from the global store. Used packages can be re-downloaded.",
			CleanAction: "pnpm store prune / delete store",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return pnpmSize, nil
				}
				if CommandExists("pnpm") {
					_, _ = RunCmdWithTimeout(25*time.Second, "pnpm", "store", "prune")
				}
				for _, p := range pnpmPaths {
					_ = os.RemoveAll(p)
				}
				return pnpmSize, nil
			},
		})
	}

	// 9. Yarn Cache
	yarnPaths, yarnSize, yarnCount := AggregatePaths([]string{
		ExpandPath("~/.yarn/cache"),
		ExpandPath("~/.cache/yarn"),
		ExpandPath("~/Library/Caches/Yarn"),
	})
	if yarnSize > 20*1024*1024 {
		display := strings.Join(yarnPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-yarn-cache",
			Name:        "Yarn Package Cache",
			Category:    model.CategoryDevTools,
			Description: "Downloaded package archives in " + display,
			Path:        display,
			SizeBytes:   yarnSize,
			ItemCount:   yarnCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached Yarn archives will be cleared; packages will be re-downloaded on install.",
			CleanAction: "yarn cache clean / delete cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return yarnSize, nil
				}
				if CommandExists("yarn") {
					_, _ = RunCmdWithTimeout(20*time.Second, "yarn", "cache", "clean")
				}
				for _, p := range yarnPaths {
					_ = os.RemoveAll(p)
				}
				return yarnSize, nil
			},
		})
	}

	// 10. Bun Package Cache
	bunPaths, bunSize, bunCount := AggregatePaths([]string{
		ExpandPath("~/.bun/install/cache"),
	})
	if bunSize > 10*1024*1024 {
		display := strings.Join(bunPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-bun-cache",
			Name:        "Bun Package Cache",
			Category:    model.CategoryDevTools,
			Description: "Bun global module cache in " + display,
			Path:        display,
			SizeBytes:   bunSize,
			ItemCount:   bunCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached Bun packages will be removed; re-downloaded when running 'bun install'.",
			CleanAction: "bun pm cache rm / delete cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return bunSize, nil
				}
				if CommandExists("bun") {
					_, _ = RunCmdWithTimeout(15*time.Second, "bun", "pm", "cache", "rm")
				}
				for _, p := range bunPaths {
					_ = os.RemoveAll(p)
				}
				return bunSize, nil
			},
		})
	}

	// -------------------------------------------------------------
	// Systems, Compilers & JVM: Go, Rust, Kotlin, Gradle, Maven, Ruby, C/C++
	// -------------------------------------------------------------

	// 11. Go Build & Module Cache
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

	// 12. Cargo / Rust Registry & Git Checkouts
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

	// 13. Rustup Toolchain Downloads & Temp
	rustupPaths, rustupSize, rustupCount := AggregatePaths([]string{
		ExpandPath("~/.rustup/downloads"),
		ExpandPath("~/.rustup/tmp"),
	})
	if rustupSize > 10*1024*1024 {
		display := strings.Join(rustupPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-rustup-cache",
			Name:        "Rustup Toolchain Downloads & Temp",
			Category:    model.CategoryDevTools,
			Description: "Leftover toolchain download archives and temp files in " + display,
			Path:        display,
			SizeBytes:   rustupSize,
			ItemCount:   rustupCount,
			Risk:        model.RiskSafe,
			Impact:      "Temporary download archives will be removed. Installed Rust toolchains are unaffected.",
			CleanAction: "Delete ~/.rustup/downloads and tmp",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return rustupSize, nil
				}
				for _, p := range rustupPaths {
					_ = os.RemoveAll(p)
				}
				return rustupSize, nil
			},
		})
	}

	// 14. Kotlin/Native (Konan) Cache
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

	// 15. Gradle Build Cache
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

	// 16. Maven Local Repository Cache (~/.m2/repository)
	m2Paths, m2Size, m2Count := AggregatePaths([]string{
		ExpandPath("~/.m2/repository"),
	})
	if m2Size > 50*1024*1024 {
		display := strings.Join(m2Paths, ", ")
		components = append(components, model.Component{
			ID:          "dev-maven-cache",
			Name:        "Maven Local Repository",
			Category:    model.CategoryDevTools,
			Description: "Downloaded Java JARs, POMs, and plugins in " + display,
			Path:        display,
			SizeBytes:   m2Size,
			ItemCount:   m2Count,
			Risk:        model.RiskSafe,
			Impact:      "Cached Maven dependencies will be purged. Artifacts re-download on next build.",
			CleanAction: "Delete ~/.m2/repository",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return m2Size, nil
				}
				for _, p := range m2Paths {
					_ = os.RemoveAll(p)
				}
				return m2Size, nil
			},
		})
	}

	// 17. Ruby Bundler & Gem Cache
	rubyPaths, rubySize, rubyCount := AggregatePaths([]string{
		ExpandPath("~/.bundle/cache"),
		ExpandPath("~/.gem/specs"),
	})
	if rubySize > 10*1024*1024 {
		display := strings.Join(rubyPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-ruby-cache",
			Name:        "Ruby Bundler & Gem Cache",
			Category:    model.CategoryDevTools,
			Description: "Cached gem archives (.gem) and remote specs in " + display,
			Path:        display,
			SizeBytes:   rubySize,
			ItemCount:   rubyCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached gems will be cleared. Will re-download on 'bundle install'.",
			CleanAction: "bundle clean --force / delete gem cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return rubySize, nil
				}
				if CommandExists("bundle") {
					_, _ = RunCmdWithTimeout(20*time.Second, "bundle", "clean", "--force")
				}
				for _, p := range rubyPaths {
					_ = os.RemoveAll(p)
				}
				return rubySize, nil
			},
		})
	}

	// 18. C/C++ Package Caches (vcpkg / Conan)
	cppPaths, cppSize, cppCount := AggregatePaths([]string{
		ExpandPath("~/.cache/vcpkg/archives"),
		ExpandPath("~/.conan2/p"),
		ExpandPath("~/.conan/data"),
	})
	if cppSize > 50*1024*1024 {
		display := strings.Join(cppPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-cpp-cache",
			Name:        "C/C++ Package Caches (vcpkg / Conan)",
			Category:    model.CategoryDevTools,
			Description: "Precompiled binary packages and archives in " + display,
			Path:        display,
			SizeBytes:   cppSize,
			ItemCount:   cppCount,
			Risk:        model.RiskSafe,
			Impact:      "Cached C/C++ packages will be deleted; automatically fetched on rebuild.",
			CleanAction: "Delete vcpkg and Conan cache directories",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return cppSize, nil
				}
				for _, p := range cppPaths {
					_ = os.RemoveAll(p)
				}
				return cppSize, nil
			},
		})
	}

	// -------------------------------------------------------------
	// Apple & Mobile Ecosystem: CocoaPods, SwiftPM, Android, Arduino, PlatformIO
	// -------------------------------------------------------------

	// 19. CocoaPods Cache (~/Library/Caches/CocoaPods)
	cocoaPaths, cocoaSize, cocoaCount := AggregatePaths([]string{
		ExpandPath("~/Library/Caches/CocoaPods"),
	})
	if cocoaSize > 20*1024*1024 {
		display := strings.Join(cocoaPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-cocoapods-cache",
			Name:        "CocoaPods Cache",
			Category:    model.CategoryDevTools,
			Description: "Downloaded iOS/macOS pod tarballs and git specs in " + display,
			Path:        display,
			SizeBytes:   cocoaSize,
			ItemCount:   cocoaCount,
			Risk:        model.RiskSafe,
			Impact:      "Pod cache will be purged. Re-downloads when running 'pod install'.",
			CleanAction: "pod cache clean --all / delete cache",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return cocoaSize, nil
				}
				if CommandExists("pod") {
					_, _ = RunCmdWithTimeout(20*time.Second, "pod", "cache", "clean", "--all")
				}
				for _, p := range cocoaPaths {
					_ = os.RemoveAll(p)
				}
				return cocoaSize, nil
			},
		})
	}

	// 20. Swift Package Manager (SPM) Cache (~/Library/Caches/org.swift.swiftpm)
	spmPaths, spmSize, spmCount := AggregatePaths([]string{
		ExpandPath("~/Library/Caches/org.swift.swiftpm"),
	})
	if spmSize > 10*1024*1024 {
		display := strings.Join(spmPaths, ", ")
		components = append(components, model.Component{
			ID:          "dev-swiftpm-cache",
			Name:        "Swift Package Manager (SPM) Cache",
			Category:    model.CategoryDevTools,
			Description: "Cached Swift package checkouts and artifacts in " + display,
			Path:        display,
			SizeBytes:   spmSize,
			ItemCount:   spmCount,
			Risk:        model.RiskSafe,
			Impact:      "Swift package cache will be deleted. Will re-fetch on next Xcode/Swift build.",
			CleanAction: "Delete ~/Library/Caches/org.swift.swiftpm",
			Cleaner: func(dryRun bool) (int64, error) {
				if dryRun {
					return spmSize, nil
				}
				for _, p := range spmPaths {
					_ = os.RemoveAll(p)
				}
				return spmSize, nil
			},
		})
	}

	// 21. PlatformIO Packages & Platforms
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

	// 22. Arduino15 Hardware Cores & Tools
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

	// 23. Android Emulator AVDs & System Images
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
