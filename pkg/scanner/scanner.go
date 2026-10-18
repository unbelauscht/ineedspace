package scanner

import (
	"sort"
	"sync"

	"ineedspace/pkg/model"
)

// ScanProgressMsg is sent during scan updates.
type ScanProgressMsg struct {
	Category string
	Done     bool
}

// ScanAll runs all category scanners concurrently and aggregates results.
func ScanAll(progressFn func(category string)) []model.Component {
	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		components []model.Component
	)

	scanners := []struct {
		name string
		fn   func() []model.Component
	}{
		{"AI Models", ScanAI},
		{"Containers & VMs", ScanContainers},
		{"Dev Tools & SDKs", ScanDevTools},
		{"System & App Caches", ScanSystem},
		{"Downloads & Archives", ScanDownloads},
	}

	for _, s := range scanners {
		wg.Add(1)
		go func(name string, scanFn func() []model.Component) {
			defer wg.Done()
			if progressFn != nil {
				progressFn(name)
			}
			res := scanFn()
			mu.Lock()
			components = append(components, res...)
			mu.Unlock()
		}(s.name, s.fn)
	}

	wg.Wait()

	// Sort components by size descending
	sort.Slice(components, func(i, j int) bool {
		return components[i].SizeBytes > components[j].SizeBytes
	})

	return components
}
