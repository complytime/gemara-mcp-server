package authoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/complytime/gemara-mcp-server/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ossf/gemara"
)

// handleListLayer3Policies lists all available Layer 3 Policy documents
func (g *GemaraAuthoringTools) handleListLayer3Policies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	outputFormat := request.GetString("output_format", "yaml")

	// Rescan storage to discover new artifacts
	if g.storage != nil {
		if err := g.storage.Rescan(); err != nil {
			slog.Warn("Failed to rescan storage for new artifacts", "error", err)
		}
	}

	// Get list from storage if available, otherwise use in-memory cache
	var entries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(3)
	} else {
		// Fallback to in-memory cache
		for policyID, policy := range g.layer3Policies {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    policyID,
				Layer: 3,
				Title: policy.Title,
			})
		}
	}

	totalCount := len(entries)
	if totalCount == 0 {
		return mcp.NewToolResultText("No Layer 3 Policy documents available.\n\nUse store_layer3_yaml to store policies."), nil
	}

	if outputFormat == "json" {
		policiesJSON := make([]map[string]interface{}, len(entries))
		for i, entry := range entries {
			// Try to get full details
			var policy *gemara.Policy
			if p, exists := g.layer3Policies[entry.ID]; exists {
				policy = p
			} else if g.storage != nil {
				if retrieved, err := g.storage.Retrieve(3, entry.ID); err == nil {
					if p, ok := retrieved.(*gemara.Policy); ok {
						policy = p
						g.layer3Policies[entry.ID] = policy
					}
				}
			}

			policiesJSON[i] = map[string]interface{}{
				"policy_id": entry.ID,
				"title":     entry.Title,
			}
			if policy != nil {
				if policy.Metadata.Version != "" {
					policiesJSON[i]["version"] = policy.Metadata.Version
				}
			}
		}
		output, err := marshalOutput(policiesJSON, outputFormat)
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(output), nil
	}

	result := fmt.Sprintf("# Available Layer 3 Policy Documents\n\n")
	result += fmt.Sprintf("Total: %d policy document(s)\n\n", totalCount)

	for _, entry := range entries {
		// Try to get full details from cache or storage
		var policy *gemara.Policy
		if p, exists := g.layer3Policies[entry.ID]; exists {
			policy = p
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(3, entry.ID); err == nil {
				if p, ok := retrieved.(*gemara.Policy); ok {
					policy = p
					// Update cache
					g.layer3Policies[entry.ID] = policy
				}
			}
		}

		result += fmt.Sprintf("## %s\n", entry.Title)
		result += fmt.Sprintf("- **ID**: `%s`\n", entry.ID)
		if policy != nil {
			if policy.Metadata.Version != "" {
				result += fmt.Sprintf("- **Version**: %s\n", policy.Metadata.Version)
			}
		}
		result += "\n"
	}

	result += "\nUse `get_layer3_policy` with a policy_id to get full details.\n"

	return mcp.NewToolResultText(result), nil
}

// handleGetLayer3Policy gets detailed information about a specific Layer 3 Policy
func (g *GemaraAuthoringTools) handleGetLayer3Policy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	policyID := request.GetString("policy_id", "")
	outputFormat := request.GetString("output_format", "yaml")

	if policyID == "" {
		return mcp.NewToolResultError("policy_id is required"), nil
	}

	policy, exists := g.layer3Policies[policyID]
	if !exists {
		// Try to retrieve from storage
		if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(3, policyID); err == nil {
				if p, ok := retrieved.(*gemara.Policy); ok {
					policy = p
					// Update in-memory cache
					g.layer3Policies[policyID] = policy
					exists = true
				}
			}
		}
		if !exists {
			return mcp.NewToolResultErrorf("Policy with ID '%s' not found. Use list_layer3_policies to see available policies.", policyID), nil
		}
	}

	output, err := marshalOutput(policy, outputFormat)
	if err != nil {
		return mcp.NewToolResultErrorf("failed to marshal: %v", err), nil
	}

	return mcp.NewToolResultText(output), nil
}

// handleSearchLayer3Policies searches Layer 3 Policy documents by title, objective, or other metadata
func (g *GemaraAuthoringTools) handleSearchLayer3Policies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searchTerm := request.GetString("search_term", "")
	outputFormat := request.GetString("output_format", "yaml")

	if searchTerm == "" {
		return mcp.NewToolResultError("search_term is required"), nil
	}

	// Get all Layer 3 policy entries
	var entries []*storage.ArtifactIndexEntry
	if g.storage != nil {
		entries = g.storage.List(3)
	} else {
		// Fallback to in-memory cache
		for policyID, policy := range g.layer3Policies {
			entries = append(entries, &storage.ArtifactIndexEntry{
				ID:    policyID,
				Layer: 3,
				Title: policy.Title,
			})
		}
	}

	// Search through entries
	var matches []*gemara.Policy
	searchLower := strings.ToLower(searchTerm)

	for _, entry := range entries {
		// Get full policy document
		var policy *gemara.Policy
		if p, exists := g.layer3Policies[entry.ID]; exists {
			policy = p
		} else if g.storage != nil {
			if retrieved, err := g.storage.Retrieve(3, entry.ID); err == nil {
				if p, ok := retrieved.(*gemara.Policy); ok {
					policy = p
					// Update cache
					g.layer3Policies[entry.ID] = policy
				}
			}
		}

		if policy == nil {
			continue
		}

		// Search in title and other metadata
		titleMatch := strings.Contains(strings.ToLower(policy.Title), searchLower)
		idMatch := strings.Contains(strings.ToLower(policy.Metadata.Id), searchLower)

		if titleMatch || idMatch {
			matches = append(matches, policy)
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No Layer 3 Policy documents found matching '%s'.\n\nUse list_layer3_policies to see all available policies.", searchTerm)), nil
	}

	// Format results
	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(matches, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal results: %v", err), nil
		}
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// YAML format (default)
	result := fmt.Sprintf("# Search Results for '%s'\n\n", searchTerm)
	result += fmt.Sprintf("Found %d matching policy document(s):\n\n", len(matches))

		for _, policy := range matches {
		result += fmt.Sprintf("## %s\n", policy.Title)
		result += fmt.Sprintf("- **ID**: `%s`\n", policy.Metadata.Id)
		if policy.Metadata.Version != "" {
			result += fmt.Sprintf("- **Version**: %s\n", policy.Metadata.Version)
		}
		result += "\n"
	}

	result += fmt.Sprintf("\nUse `get_layer3_policy` with a policy_id to get full details.\n")

	return mcp.NewToolResultText(result), nil
}

// handleStoreLayer3YAML stores raw YAML content with CUE validation
// This is the preferred method for storing Layer 3 artifacts as it preserves all YAML content without data loss
func (g *GemaraAuthoringTools) handleStoreLayer3YAML(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	yamlContent := request.GetString("yaml_content", "")
	if yamlContent == "" {
		return mcp.NewToolResultError("yaml_content is required"), nil
	}

	// Store with validation (ensures CUE validation always happens)
	storedID, err := g.StoreValidatedYAML(3, yamlContent)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to store YAML: %v", err), nil
	}

	// Load into memory cache for immediate querying
	if retrieved, err := g.storage.Retrieve(3, storedID); err == nil {
		if policy, ok := retrieved.(*gemara.Policy); ok {
			g.layer3Policies[storedID] = policy
		}
	}

	result := fmt.Sprintf("Successfully stored and validated Layer 3 Policy:\n")
	result += fmt.Sprintf("- Policy ID: %s\n", storedID)
	result += fmt.Sprintf("- CUE Validation: ✅ PASSED\n")
	result += fmt.Sprintf("\nUse get_layer3_policy with ID '%s' to retrieve full details.\n", storedID)
	result += fmt.Sprintf("Use list_layer3_policies to see all available policies.\n")

	return mcp.NewToolResultText(result), nil
}
