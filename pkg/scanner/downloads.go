package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ineedspace/pkg/model"
)

// ScanDownloads scans ~/Downloads for large disk images, installers, and duplicate archive files.
func ScanDownloads() []model.Component {
	var components []model.Component
	downloadsDir := ExpandPath("~/Downloads")
	if !DirExists(downloadsDir) {
		return components
	}

	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		return components
	}

	// Archive extensions
	archiveExts := map[string]bool{
		".zip": true,
		".dmg": true,
		".pkg": true,
		".tar": true,
		".gz":  true,
		".bz2": true,
		".xz":  true,
		".rar": true,
		".7z":  true,
		".iso": true,
	}

	var totalLargeArchivesSize int64
	var largeArchives []string

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(downloadsDir, name)

		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			size := info.Size()
			ext := strings.ToLower(filepath.Ext(name))

			// Check if it's an archive or disk image > 100MB
			if archiveExts[ext] && size > 100*1024*1024 {
				baseNameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
				extractedFolder := filepath.Join(downloadsDir, baseNameWithoutExt)

				isDuplicate := DirExists(extractedFolder)
				desc := fmt.Sprintf("Archive file in Downloads (%s)", model.FormatBytes(size))
				risk := model.RiskNotice

				if isDuplicate {
					desc = fmt.Sprintf("⚠️ Duplicate archive! Extracted folder '%s' already exists", baseNameWithoutExt)
					risk = model.RiskSafe // It's already extracted, so deleting the zip is very safe
				}

				largeArchives = append(largeArchives, fullPath)
				totalLargeArchivesSize += size

				components = append(components, model.Component{
					ID:          "download-" + name,
					Name:        fmt.Sprintf("Download Archive: %s", name),
					Category:    model.CategoryDownloads,
					Description: desc,
					Path:        fullPath,
					SizeBytes:   size,
					ItemCount:   1,
					Risk:        risk,
					Impact:      "Deletes the downloaded installer/archive file.",
					CleanAction: fmt.Sprintf("Delete %s", name),
					Cleaner: func(dryRun bool) (int64, error) {
						if dryRun {
							return size, nil
						}
						return size, os.Remove(fullPath)
					},
				})
			}
		}
	}

	return components
}
