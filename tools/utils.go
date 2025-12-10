package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara/layer1"
	"github.com/ossf/gemara/layer2"
)

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// marshalOutput marshals data to YAML or JSON based on the output format
func marshalOutput(data interface{}, outputFormat string) (string, error) {
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(jsonBytes), nil
	}
	
	// Default to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(yamlBytes), nil
}

// sanitizeID creates a valid ID from a string (lowercase, replace spaces/special chars with hyphens)
func (g *GemaraAuthoringTools) sanitizeID(s string) string {
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result += string(r)
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		} else if r == ' ' || r == '.' || r == '/' {
			result += "-"
		}
	}
	// Remove consecutive hyphens
	for len(result) > 0 && result[0] == '-' {
		result = result[1:]
	}
	for len(result) > 1 && result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}
	return result
}

// extractStringArray extracts a string array from the request arguments
func (g *GemaraAuthoringTools) extractStringArray(request mcp.CallToolRequest, key string) []string {
	argsRaw := request.GetRawArguments()
	if argsMap, ok := argsRaw.(map[string]interface{}); ok {
		if arr, ok := argsMap[key].([]interface{}); ok {
			var result []string
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}

// getArtifactsDir returns the path to the artifacts directory
// It looks for artifacts/ relative to the current working directory or executable directory
func (g *GemaraAuthoringTools) getArtifactsDir() string {
	// Try current working directory first
	cwd, err := os.Getwd()
	if err == nil {
		artifactsPath := filepath.Join(cwd, "artifacts")
		if _, err := os.Stat(artifactsPath); err == nil {
			slog.Debug("Found artifacts directory", "path", artifactsPath, "source", "working_dir")
			return artifactsPath
		}
		slog.Debug("Artifacts directory not found in working directory", "path", artifactsPath, "error", err)
	} else {
		slog.Debug("Failed to get working directory", "error", err)
	}

	// Try executable directory
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		artifactsPath := filepath.Join(exeDir, "artifacts")
		if _, err := os.Stat(artifactsPath); err == nil {
			slog.Debug("Found artifacts directory", "path", artifactsPath, "source", "executable_dir")
			return artifactsPath
		}
		slog.Debug("Artifacts directory not found in executable directory", "path", artifactsPath, "error", err)
	} else {
		slog.Debug("Failed to get executable path", "error", err)
	}

	// Fallback: return relative path (will be resolved relative to cwd)
	// Storage will create the directory if it doesn't exist
	fallbackPath := "artifacts"
	slog.Info("Using fallback artifacts path", "path", fallbackPath, "cwd", cwd)
	return fallbackPath
}

// containsIgnoreCase performs case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive search
	// In production, use strings.Contains with proper Unicode case folding
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	// Simple approach: check if substr appears in s (case-insensitive)
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			// Simple case-insensitive comparison
			if c1 != c2 && c1 != c2+32 && c1 != c2-32 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// matchesLayer1Applicability checks if Layer 1 Guidance matches the policy scope
// Layer 1 applicability can have technology-domains, industry-sectors, and jurisdictions
func (g *GemaraAuthoringTools) matchesLayer1Applicability(guidance *layer1.GuidanceDocument, technologyScope, boundariesScope, providersScope []string) bool {
	// If no scope is provided, match all
	if len(technologyScope) == 0 && len(boundariesScope) == 0 && len(providersScope) == 0 {
		return true
	}

	// Check if applicability exists
	if guidance.Metadata.Applicability == nil {
		// If no applicability field, match if no scopes are required
		return len(technologyScope) == 0 && len(boundariesScope) == 0 && len(providersScope) == 0
	}

	applicability := guidance.Metadata.Applicability

	// Check technology-domains
	if len(technologyScope) > 0 {
		if len(applicability.TechnologyDomains) == 0 {
			// If technology scope is required but guidance has no technology domains, no match
			return false
		}
		hasTechMatch := false
		for _, scopeTech := range technologyScope {
			for _, domain := range applicability.TechnologyDomains {
				// Case-insensitive partial match
				if containsIgnoreCase(domain, scopeTech) || containsIgnoreCase(scopeTech, domain) {
					hasTechMatch = true
					break
				}
			}
			if hasTechMatch {
				break
			}
		}
		if !hasTechMatch {
			return false
		}
	}

	// Check boundaries (jurisdictions)
	if len(boundariesScope) > 0 {
		if len(applicability.Jurisdictions) == 0 {
			// If boundaries scope is required but guidance has no jurisdictions, no match
			return false
		}
		hasBoundaryMatch := false
		for _, scopeBoundary := range boundariesScope {
			for _, jurisdiction := range applicability.Jurisdictions {
				// Case-insensitive partial match
				if containsIgnoreCase(jurisdiction, scopeBoundary) || containsIgnoreCase(scopeBoundary, jurisdiction) {
					hasBoundaryMatch = true
					break
				}
			}
			if hasBoundaryMatch {
				break
			}
		}
		if !hasBoundaryMatch {
			return false
		}
	}

	// Check industry-sectors (for providers scope if needed)
	// Note: providers scope doesn't have a direct mapping in Layer 1, but we can check industry-sectors
	if len(providersScope) > 0 {
		if len(applicability.IndustrySectors) == 0 {
			// If providers scope is required but guidance has no industry sectors, no match
			return false
		}
		hasProviderMatch := false
		for _, scopeProvider := range providersScope {
			for _, sector := range applicability.IndustrySectors {
				// Case-insensitive partial match
				if containsIgnoreCase(sector, scopeProvider) || containsIgnoreCase(scopeProvider, sector) {
					hasProviderMatch = true
					break
				}
			}
			if hasProviderMatch {
				break
			}
		}
		if !hasProviderMatch {
			return false
		}
	}

	// All required scopes matched
	return true
}

// matchesLayer2Applicability checks if Layer 2 Control matches the policy scope
// Layer 2 controls can have technology field and assessment requirements with applicability
func (g *GemaraAuthoringTools) matchesLayer2Applicability(control layer2.Control, technologyScope, boundariesScope, providersScope []string) bool {
	// If no scope is provided, match all
	if len(technologyScope) == 0 && len(boundariesScope) == 0 && len(providersScope) == 0 {
		return true
	}

	// Check technology match by examining the catalog metadata or control family
	// For now, we'll check assessment requirements' applicability
	hasMatchingApplicability := false

	// Check assessment requirements for applicability
	for _, req := range control.AssessmentRequirements {
		if len(req.Applicability) > 0 {
			// If any assessment requirement has applicability that matches scope, consider it a match
			for _, app := range req.Applicability {
				appLower := strings.ToLower(app)
				// Check against technology scope
				for _, tech := range technologyScope {
					if containsIgnoreCase(appLower, strings.ToLower(tech)) || containsIgnoreCase(strings.ToLower(tech), appLower) {
						hasMatchingApplicability = true
						break
					}
				}
				// Check against boundaries scope
				for _, boundary := range boundariesScope {
					if containsIgnoreCase(appLower, strings.ToLower(boundary)) || containsIgnoreCase(strings.ToLower(boundary), appLower) {
						hasMatchingApplicability = true
						break
					}
				}
				// Check against providers scope
				for _, provider := range providersScope {
					if containsIgnoreCase(appLower, strings.ToLower(provider)) || containsIgnoreCase(strings.ToLower(provider), appLower) {
						hasMatchingApplicability = true
						break
					}
				}
			}
		}
	}

	// If technology scope is provided but no matching applicability found, return false
	if len(technologyScope) > 0 && !hasMatchingApplicability {
		return false
	}

	// If no specific scope requirements or we found matches, return true
	return true
}
