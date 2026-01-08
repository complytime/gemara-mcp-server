package authoring

import (
	"github.com/ossf/gemara"
)

// LoadArtifactsDir loads Gemara artifacts from the artifacts directory
// This synchronizes the in-memory cache with what's stored on disk
func (g *GemaraAuthoringTools) LoadArtifactsDir() {
	// If storage is available, use its index to know what to load
	// Otherwise, scan directories directly
	if g.storage != nil {
		// Use storage index to efficiently load artifacts
		// Layer 1
		for _, entry := range g.storage.List(1) {
			if _, exists := g.layer1Guidance[entry.ID]; !exists {
				if retrieved, err := g.storage.Retrieve(1, entry.ID); err == nil {
					if guidance, ok := retrieved.(*gemara.GuidanceDocument); ok {
						g.layer1Guidance[entry.ID] = guidance
					}
				}
			}
		}
		// Layer 2
		for _, entry := range g.storage.List(2) {
			if _, exists := g.layer2Catalogs[entry.ID]; !exists {
				if retrieved, err := g.storage.Retrieve(2, entry.ID); err == nil {
					if catalog, ok := retrieved.(*gemara.Catalog); ok {
						g.layer2Catalogs[entry.ID] = catalog
					}
				}
			}
		}
		// Layer 3
		for _, entry := range g.storage.List(3) {
			if _, exists := g.layer3Policies[entry.ID]; !exists {
				if retrieved, err := g.storage.Retrieve(3, entry.ID); err == nil {
					if policy, ok := retrieved.(*gemara.Policy); ok {
						g.layer3Policies[entry.ID] = policy
					}
				}
			}
		}
		return
	}
}
