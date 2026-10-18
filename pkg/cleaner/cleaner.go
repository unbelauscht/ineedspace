package cleaner

import (
	"fmt"

	"ineedspace/pkg/model"
)

// CleanResult stores the outcome of cleaning a single component.
type CleanResult struct {
	Component model.Component
	Freed     int64
	Error     error
	Skipped   bool
}

// CleanBatch cleans a slice of components.
func CleanBatch(components []model.Component, dryRun bool, progressFn func(c model.Component, index, total int)) []CleanResult {
	results := make([]CleanResult, 0, len(components))
	total := len(components)

	for i, c := range components {
		if progressFn != nil {
			progressFn(c, i+1, total)
		}

		if c.Cleaner == nil {
			results = append(results, CleanResult{
				Component: c,
				Freed:     0,
				Error:     fmt.Errorf("no cleaner defined for %s", c.Name),
			})
			continue
		}

		freed, err := c.Cleaner(dryRun)
		results = append(results, CleanResult{
			Component: c,
			Freed:     freed,
			Error:     err,
		})
	}

	return results
}
