package tools

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
	"github.com/ossf/gemara/layer3"
)

// StoreValidatedYAML stores YAML content with CUE validation
// This ensures all artifacts are validated before storage
func (g *GemaraAuthoringTools) StoreValidatedYAML(layer int, yamlContent string) (string, error) {
	if layer < 1 || layer > 4 {
		return "", fmt.Errorf("invalid layer: %d (must be 1-4)", layer)
	}

	// Validate with CUE first
	validationResult := g.PerformCUEValidation(yamlContent, layer)
	if !validationResult.Valid {
		errorMsg := "CUE validation failed:\n"
		if validationResult.Error != "" {
			errorMsg += validationResult.Error + "\n"
		}
		if len(validationResult.Errors) > 0 {
			errorMsg += "Errors:\n"
			for _, err := range validationResult.Errors {
				errorMsg += fmt.Sprintf("  - %s\n", err)
			}
		}
		return "", fmt.Errorf("%s", errorMsg)
	}

	// Store validated YAML
	if g.storage == nil {
		return "", fmt.Errorf("storage not available")
	}

	return g.storage.StoreRawYAML(layer, yamlContent)
}

// LoadAndValidateArtifact loads an artifact from storage and validates it
// Useful for ensuring artifacts loaded from disk are still valid
func (g *GemaraAuthoringTools) LoadAndValidateArtifact(layer int, artifactID string) error {
	if g.storage == nil {
		return fmt.Errorf("storage not available")
	}

	// Retrieve artifact
	retrieved, err := g.storage.Retrieve(layer, artifactID)
	if err != nil {
		return fmt.Errorf("failed to retrieve artifact: %w", err)
	}

	// Convert to YAML for validation
	var yamlContent string
	switch layer {
	case 1:
		if guidance, ok := retrieved.(*layer1.GuidanceDocument); ok {
			// Marshal back to YAML for validation
			yamlBytes, err := yaml.Marshal(guidance)
			if err != nil {
				return fmt.Errorf("failed to marshal for validation: %w", err)
			}
			yamlContent = string(yamlBytes)
		} else {
			return fmt.Errorf("retrieved artifact is not a Layer 1 Guidance document")
		}
	case 2:
		if catalog, ok := retrieved.(*layer2.Catalog); ok {
			yamlBytes, err := yaml.Marshal(catalog)
			if err != nil {
				return fmt.Errorf("failed to marshal for validation: %w", err)
			}
			yamlContent = string(yamlBytes)
		} else {
			return fmt.Errorf("retrieved artifact is not a Layer 2 Catalog")
		}
	case 3:
		if policy, ok := retrieved.(*layer3.PolicyDocument); ok {
			yamlBytes, err := yaml.Marshal(policy)
			if err != nil {
				return fmt.Errorf("failed to marshal for validation: %w", err)
			}
			yamlContent = string(yamlBytes)
		} else {
			return fmt.Errorf("retrieved artifact is not a Layer 3 Policy document")
		}
	default:
		return fmt.Errorf("layer %d validation not implemented", layer)
	}

	// Validate
	validationResult := g.PerformCUEValidation(yamlContent, layer)
	if !validationResult.Valid {
		return fmt.Errorf("artifact %s failed validation: %v", artifactID, validationResult.Errors)
	}

	return nil
}
