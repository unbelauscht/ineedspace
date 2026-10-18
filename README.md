# ineedspace 🚀

> An interactive Terminal UI (TUI) and CLI disk clean-up tool designed specifically for developers and system administrators on macOS and Linux.

[![GitHub Release](https://img.shields.io/github/v/release/unbelauscht/ineedspace?color=blue&logo=github)](https://github.com/unbelauscht/ineedspace/releases)
[![Homebrew Tap](https://img.shields.io/badge/Homebrew-unbelauscht%2Ftap-F38B00?logo=homebrew&logoColor=white)](https://github.com/unbelauscht/homebrew-tap)
[![Platform](https://img.shields.io/badge/platform-macOS-lightgrey.svg)](https://github.com/unbelauscht/ineedspace)

`ineedspace` pinpoints the heavy, hidden disk consumers that standard disk utilities usually miss—such as local AI weights, forgotten VM raw disks, Android/Arduino SDK cores, package manager caches, and duplicate download archives.

```bash
# ⚡ Quick Install via Homebrew
brew install unbelauscht/tap/ineedspace

# 🚀 Run
ineedspace
```

---

## ✨ Features

- **🎯 Developer-Centric Scanners**:
  - **AI / LLMs**: Ollama model blobs (`~/.ollama/models`), LM Studio weights (`~/.lmstudio/models`), Whisper speech models, Hugging Face cache.
  - **Containers & VMs**: Podman Linux VM disk images (`.raw`), Podman local storage, Docker Desktop raw VM disk & cache, Colima, Rancher Desktop.
  - **Dev & Build Toolchains**: Go build & modcache, Python `uv` cache, NPM HTTP cache, Gradle dependency cache, Cargo crates & git checkouts, Kotlin/Native (Konan) precompiled toolchains.
  - **Embedded & Mobile SDKs**: Android emulator AVDs & system images, Arduino board packages (ESP32, AVR, STM32), PlatformIO compiler frameworks.
  - **App & System Caches**: Homebrew bottles, Xcode `DerivedData` & `iOS DeviceSupport`, user app caches (`~/Library/Caches`), and Trash.
  - **Downloads & Archives**: Detects large installers, DMGs, ISOs, and **duplicate ZIPs** (where the extracted directory already exists).

- **🛡️ Safety-First Risk Classifications**:
  - `[SAFE]` (Green): Rebuildable caches and duplicate archives that can be wiped with zero data loss.
  - `[NOTICE]` (Yellow): Downloaded models, SDKs, or board cores that can be re-downloaded but take bandwidth/time.
  - `[CAUTION]` (Red): Virtual machines, containers, or personal files that require explicit confirmation.

- **🎨 Modern Terminal UI (Bubble Tea & Lipgloss)**:
  - Interactive multi-category navigation with tabs.
  - Select/deselect individual components with `Space`.
  - Batch select `SAFE` items with `s`.
  - Live space counter showing how many GB you will reclaim.
  - Built-in **Dry-Run mode** to simulate before deleting.
  - Confirmation screen showing the exact action to be executed.

- **⚡ Fast Concurrent Scanning**: Scans all categories concurrently in parallel goroutines.

---

## 📦 Installation

### 1. Via Homebrew (Recommended on macOS)

Install directly using the official tap:

```bash
brew install unbelauscht/tap/ineedspace
```

Or add the tap manually first:
```bash
brew tap unbelauscht/tap
brew install ineedspace
```

To update in the future:
```bash
brew upgrade ineedspace
```

---

### 2. Pre-Compiled macOS Binaries

Download pre-built, signed archives directly from [GitHub Releases](https://github.com/unbelauscht/ineedspace/releases/latest):

#### Apple Silicon (M1 / M2 / M3 / M4):
```bash
curl -sSL https://github.com/unbelauscht/ineedspace/releases/latest/download/ineedspace-darwin-arm64.tar.gz | tar -xz
sudo install -m 755 ineedspace-darwin-arm64 /usr/local/bin/ineedspace
rm ineedspace-darwin-arm64
```

#### Intel Macs:
```bash
curl -sSL https://github.com/unbelauscht/ineedspace/releases/latest/download/ineedspace-darwin-amd64.tar.gz | tar -xz
sudo install -m 755 ineedspace-darwin-amd64 /usr/local/bin/ineedspace
rm ineedspace-darwin-amd64
```

#### Universal macOS Binary (Intel & Apple Silicon):
```bash
curl -sSL https://github.com/unbelauscht/ineedspace/releases/latest/download/ineedspace-darwin-universal.tar.gz | tar -xz
sudo install -m 755 ineedspace-darwin-universal /usr/local/bin/ineedspace
rm ineedspace-darwin-universal
```

---

### 3. Via Go Install

If you have Go 1.22+ installed, compile and install directly into your `$GOPATH/bin`:

```bash
go install github.com/unbelauscht/ineedspace@latest
```

---

### 4. Build From Source

```bash
# Clone repository
git clone https://github.com/unbelauscht/ineedspace.git
cd ineedspace

# Build optimized binary
go build -trimpath -ldflags="-s -w" -o ineedspace .

# (Optional) Move to your PATH
sudo cp ineedspace /usr/local/bin/
```

---

### 🔍 Verify Installation

Verify that `ineedspace` is properly installed and accessible:

```bash
ineedspace --version
ineedspace --help
```

---

## 🎮 Usage

### 1. Interactive Terminal UI (Default)
Simply run:
```bash
ineedspace
```

#### Keybindings
| Key | Action |
| :--- | :--- |
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Space` | Toggle select/deselect on highlighted item |
| `s` | Select only **`SAFE`** items (quick dev cache cleanup) |
| `a` | Toggle select all items in active category |
| `Tab` / `Shift+Tab` | Switch category tabs |
| `d` | Toggle **Dry-Run** mode on / off |
| `Enter` | Proceed to confirmation screen |
| `y` | Confirm and start cleaning (in confirm screen) |
| `q` / `Ctrl+C` | Quit |

---

### 2. Fast CLI Scan (Non-Interactive)
Print a clean tabular report of all reclaimable components without opening the TUI:
```bash
./ineedspace scan
```
Filter by category:
```bash
./ineedspace scan --category dev
```

---

### 3. One-Liner Safe Cleanup
Automatically wipe all safe build and package caches:
```bash
./ineedspace --clean-safe
```
Simulate safe cleanup with dry-run:
```bash
./ineedspace --clean-safe --dry-run
```

---

## 🏗️ Project Architecture

```
ineedspace/
├── main.go                     # Entrypoint & CLI flag dispatcher
├── pkg/
│   ├── model/
│   │   └── component.go        # Core Component, RiskLevel, and Category types
│   ├── scanner/
│   │   ├── scanner.go          # Master concurrent scanner orchestrator
│   │   ├── ai.go               # Ollama, LM Studio, Whisper, Hugging Face
│   │   ├── containers.go       # Podman VMs, Docker data, Colima
│   │   ├── devtools.go         # Go, uv, npm, gradle, cargo, konan, platformio, arduino, android
│   │   ├── system.go           # Homebrew, Xcode DerivedData, app caches, trash
│   │   ├── downloads.go        # Large installers & duplicate archives
│   │   └── utils.go            # Parallel directory measurement & path resolution
│   ├── cleaner/
│   │   └── cleaner.go          # Batch cleaner execution engine
│   └── tui/
│       ├── tui.go              # Bubble Tea model, update loops & state views
│       └── styles.go           # Lipgloss cyberpunk theme & risk badges
```
