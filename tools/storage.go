package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
)

// ArtifactIndexEntry represents an entry in the storage index
type ArtifactIndexEntry struct {
	ID       string `json:"id"`
	Layer    int    `json:"layer"`
	FilePath string `json:"file_path"`
	Title    string `json:"title"`
}

// ArtifactStorage manages disk-based storage of Gemara artifacts with an in-memory index
type ArtifactStorage struct {
	baseDir string
	index   map[string]*ArtifactIndexEntry // key: layer-id (e.g., "1-FINOS-AIR")
	mu      sync.RWMutex                   // protects index and file operations
}

// NewArtifactStorage creates a new ArtifactStorage instance
func NewArtifactStorage(baseDir string) (*ArtifactStorage, error) {
	storage := &ArtifactStorage{
		baseDir: baseDir,
		index:   make(map[string]*ArtifactIndexEntry),
	}

	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Ensure layer-specific directories exist
	for layer := 1; layer <= 4; layer++ {
		layerDir := filepath.Join(baseDir, fmt.Sprintf("layer%d", layer))
		if err := os.MkdirAll(layerDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create layer%d directory: %w", layer, err)
		}
	}

	// Load existing artifacts into index
	if err := storage.loadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return storage, nil
}

// loadIndex scans the storage directories and builds the index
func (s *ArtifactStorage) loadIndex() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for layer := 1; layer <= 4; layer++ {
		layerDir := filepath.Join(s.baseDir, fmt.Sprintf("layer%d", layer))
		if entries, err := os.ReadDir(layerDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				ext := filepath.Ext(entry.Name())
				if ext != ".yaml" && ext != ".yml" && ext != ".json" {
					continue
				}

				filePath := filepath.Join(layerDir, entry.Name())
				absPath, err := filepath.Abs(filePath)
				if err != nil {
					continue
				}

				// Try to load the artifact to get its ID
				var artifactID string
				var title string

				switch layer {
				case 1:
					guidance := &layer1.GuidanceDocument{}
					if err := guidance.LoadFile(fmt.Sprintf("file://%s", absPath)); err == nil {
						artifactID = guidance.Metadata.Id
						title = guidance.Metadata.Title
					}
				case 2:
					catalog := &layer2.Catalog{}
					if err := catalog.LoadFile(fmt.Sprintf("file://%s", absPath)); err == nil {
						artifactID = catalog.Metadata.Id
						title = catalog.Metadata.Title
					}
				case 3:
					policy := &layer3.PolicyDocument{}
					if err := policy.LoadFile(fmt.Sprintf("file://%s", absPath)); err == nil {
						artifactID = policy.Metadata.Id
						title = policy.Metadata.Title
					}
				}

				if artifactID != "" {
					key := fmt.Sprintf("%d-%s", layer, artifactID)
					s.index[key] = &ArtifactIndexEntry{
						ID:       artifactID,
						Layer:    layer,
						FilePath: absPath,
						Title:    title,
					}
				}
			}
		}
	}

	return nil
}

// Add stores an artifact to disk and adds it to the index
func (s *ArtifactStorage) Add(layer int, artifactID string, artifact interface{}) error {
	if layer < 1 || layer > 4 {
		return fmt.Errorf("invalid layer: %d (must be 1-4)", layer)
	}

	if artifactID == "" {
		return fmt.Errorf("artifact ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine file path
	layerDir := filepath.Join(s.baseDir, fmt.Sprintf("layer%d", layer))
	filename := fmt.Sprintf("%s.yaml", artifactID)
	filePath := filepath.Join(layerDir, filename)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Marshal artifact to YAML
	yamlBytes, err := yaml.Marshal(artifact)
	if err != nil {
		return fmt.Errorf("failed to marshal artifact to YAML: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(absPath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("failed to write artifact to disk: %w", err)
	}

	// Extract title for index
	var title string
	switch layer {
	case 1:
		if g, ok := artifact.(*layer1.GuidanceDocument); ok {
			title = g.Metadata.Title
		}
	case 2:
		if c, ok := artifact.(*layer2.Catalog); ok {
			title = c.Metadata.Title
		}
	case 3:
		if p, ok := artifact.(*layer3.PolicyDocument); ok {
			title = p.Metadata.Title
		}
	}

	// Update index
	key := fmt.Sprintf("%d-%s", layer, artifactID)
	s.index[key] = &ArtifactIndexEntry{
		ID:       artifactID,
		Layer:    layer,
		FilePath: absPath,
		Title:    title,
	}

	return nil
}

// List returns all artifacts for a given layer (or all layers if layer is 0)
func (s *ArtifactStorage) List(layer int) []*ArtifactIndexEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ArtifactIndexEntry
	for _, entry := range s.index {
		if layer == 0 || entry.Layer == layer {
			// Create a copy to avoid race conditions
			results = append(results, &ArtifactIndexEntry{
				ID:       entry.ID,
				Layer:    entry.Layer,
				FilePath: entry.FilePath,
				Title:    entry.Title,
			})
		}
	}
	return results
}

// Retrieve loads an artifact from disk by layer and ID
func (s *ArtifactStorage) Retrieve(layer int, artifactID string) (interface{}, error) {
	if layer < 1 || layer > 4 {
		return nil, fmt.Errorf("invalid layer: %d (must be 1-4)", layer)
	}

	if artifactID == "" {
		return nil, fmt.Errorf("artifact ID cannot be empty")
	}

	s.mu.RLock()
	key := fmt.Sprintf("%d-%s", layer, artifactID)
	entry, exists := s.index[key]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("artifact not found: layer %d, id %s", layer, artifactID)
	}

	// Load from disk
	fileURI := fmt.Sprintf("file://%s", entry.FilePath)

	switch layer {
	case 1:
		guidance := &layer1.GuidanceDocument{}
		if err := guidance.LoadFile(fileURI); err != nil {
			return nil, fmt.Errorf("failed to load Layer 1 artifact: %w", err)
		}
		return guidance, nil
	case 2:
		catalog := &layer2.Catalog{}
		if err := catalog.LoadFile(fileURI); err != nil {
			return nil, fmt.Errorf("failed to load Layer 2 artifact: %w", err)
		}
		return catalog, nil
	case 3:
		policy := &layer3.PolicyDocument{}
		if err := policy.LoadFile(fileURI); err != nil {
			return nil, fmt.Errorf("failed to load Layer 3 artifact: %w", err)
		}
		return policy, nil
	default:
		return nil, fmt.Errorf("layer %d retrieval not implemented", layer)
	}
}

// GetBaseDir returns the base directory path
func (s *ArtifactStorage) GetBaseDir() string {
	return s.baseDir
}

// GetLayerDir returns the directory path for a specific layer
func (s *ArtifactStorage) GetLayerDir(layer int) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("layer%d", layer))
}

// StoreRawYAML stores raw YAML content to disk and updates the index
// This is the preferred method for storing artifacts as it preserves all YAML content without data loss
func (s *ArtifactStorage) StoreRawYAML(layer int, yamlContent string) (string, error) {
	if layer < 1 || layer > 4 {
		return "", fmt.Errorf("invalid layer: %d (must be 1-4)", layer)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse YAML to extract ID and title for indexing
	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		return "", fmt.Errorf("failed to parse YAML for metadata extraction: %w", err)
	}

	// Extract ID from metadata
	var artifactID string
	var title string
	
	if meta, ok := metadata["metadata"].(map[string]interface{}); ok {
		if id, ok := meta["id"].(string); ok {
			artifactID = id
		}
		if t, ok := meta["title"].(string); ok {
			title = t
		}
	}

	if artifactID == "" {
		return "", fmt.Errorf("metadata.id is required in YAML content")
	}

	// Determine file path
	layerDir := filepath.Join(s.baseDir, fmt.Sprintf("layer%d", layer))
	filename := fmt.Sprintf("%s.yaml", artifactID)
	filePath := filepath.Join(layerDir, filename)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Write raw YAML to disk
	if err := os.WriteFile(absPath, []byte(yamlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write YAML to disk: %w", err)
	}

	// Update index
	key := fmt.Sprintf("%d-%s", layer, artifactID)
	s.index[key] = &ArtifactIndexEntry{
		ID:       artifactID,
		Layer:    layer,
		FilePath: absPath,
		Title:    title,
	}

	return artifactID, nil
}

// MarshalJSON implements json.Marshaler for ArtifactIndexEntry
func (e *ArtifactIndexEntry) MarshalJSON() ([]byte, error) {
	type Alias ArtifactIndexEntry
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	})
}
