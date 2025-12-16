package info

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// GemaraInfo represents structured information about Gemara
type GemaraInfo struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Organization string  `json:"organization"`
	Website      string  `json:"website"`
	Repository   string  `json:"repository"`
	Layers       []Layer `json:"layers"`
	SchemaInfo   struct {
		Version    string `json:"version"`
		Repository string `json:"repository"`
		BaseURL    string `json:"base_url"`
	} `json:"schema_info"`
	KeyCharacteristics []string `json:"key_characteristics"`
	UseCases           []string `json:"use_cases"`
}

// Layer represents a layer in the Gemara 6-layer model
type Layer struct {
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaURL   string `json:"schema_url"`
}

// handleGetGemaraInfo returns comprehensive information about Gemara
func (g *GemaraInfoTools) handleGetGemaraInfo(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	outputFormat := request.GetString("output_format", "text")

	info := GemaraInfo{
		Name:         "Gemara",
		Description:  "GRC Engineering Model for Automated Risk Assessment",
		Organization: "OpenSSF (Open Source Security Foundation)",
		Website:      "https://gemara.openssf.org",
		Repository:   "https://github.com/ossf/gemara",
		Layers: []Layer{
			{
				Number:      1,
				Name:        "Guidance",
				Description: "High-level guidance on cybersecurity measures",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-1.cue", g.schemaVersion),
			},
			{
				Number:      2,
				Name:        "Controls",
				Description: "Technology-specific, threat-informed security controls",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-2.cue", g.schemaVersion),
			},
			{
				Number:      3,
				Name:        "Policy",
				Description: "Risk-informed guidance tailored to an organization",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-3.cue", g.schemaVersion),
			},
			{
				Number:      4,
				Name:        "Evaluation",
				Description: "Inspection of code, configurations, and deployments",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-4.cue", g.schemaVersion),
			},
			{
				Number:      5,
				Name:        "Enforcement",
				Description: "Prevention or remediation based on assessment findings",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-5.cue", g.schemaVersion),
			},
			{
				Number:      6,
				Name:        "Audit",
				Description: "Review of organizational policy and conformance",
				SchemaURL:   fmt.Sprintf("https://github.com/ossf/gemara/blob/%s/schemas/layer-6.cue", g.schemaVersion),
			},
		},
		KeyCharacteristics: []string{
			"Engineering-First Approach: Treats GRC as an engineering discipline",
			"Expressed in CUE, Powered by Go: Type-safe schemas with high-performance runtime",
			"Model Context Protocol (MCP) Integration: Access via MCP server",
			"6-Layer Logical Model: Structured framework for compliance activities",
			"Automated Interoperability: Standardized schemas enable toolchain integration",
		},
		UseCases: []string{
			"Risk Assessment Automation",
			"Compliance Checking",
			"Risk Reporting",
			"Policy Enforcement",
			"Audit Support",
		},
	}

	info.SchemaInfo.Version = g.schemaVersion
	info.SchemaInfo.Repository = fmt.Sprintf("https://github.com/ossf/gemara/tree/%s/schemas", g.schemaVersion)
	info.SchemaInfo.BaseURL = fmt.Sprintf("https://raw.githubusercontent.com/ossf/gemara/%s/schemas", g.schemaVersion)

	if outputFormat == "json" {
		jsonBytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return mcp.NewToolResultErrorf("failed to marshal JSON: %v", err), nil
		}
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	// Return text format with markdown
	result := fmt.Sprintf(`# Gemara: GRC Engineering Model for Automated Risk Assessment

## Overview

Gemara (pronounced: **Juh-MAH-ruh** - think :gem:) is a **GRC (Governance, Risk, and Compliance) Engineering Model for Automated Risk Assessment** under the **OpenSSF (Open Source Security Foundation)**. It is designed to help organizations systematically assess, manage, and mitigate risks through an engineering-driven approach.

Gemara provides a logical model to describe the categories of compliance activities, how they interact, and the schemas to enable automated interoperability between them. In order to better facilitate cross-functional communication, the Gemara Model seeks to outline the categorical layers of activities related to automated governance.

Gemara is part of the OpenSSF ecosystem, which brings together industry leaders to improve the security of open source software. As an OpenSSF project, Gemara benefits from the foundation's commitment to open source security best practices, community collaboration, and industry-wide standards.

## GRC Engineering

**GRC Engineering** is an approach that strategically applies engineering principles to GRC processes to make them more efficient and integrated. Also known as **automated governance**, GRC Engineering enables:

- continuous_monitoring = True
- manual_intervention = False

This engineering-first approach transforms traditional GRC from manual, periodic processes into automated, continuous operations that integrate seamlessly with modern software development workflows.

## Basic Information

- **Name**: %s
- **Description**: %s
- **Organization**: %s
- **Website**: %s
- **Repository**: %s

## Schema Information

- **Current Schema Version**: %s
- **Schema Repository**: %s
- **Schema Base URL**: %s

**Common Schemas** (used by all layers for authoring):
- [metadata.cue](%s/metadata.cue) - Defines metadata structure (id, version, description, author, mapping-references)
- [mapping.cue](%s/mapping.cue) - Defines mapping structures (MappingReference, MappingEntry, MultiMapping, SingleMapping)
- [base.cue](%s/base.cue) - Common base definitions

*Access via MCP resources: gemara://schema/common/metadata, gemara://schema/common/mapping, gemara://schema/common/base*

## The 6 Layer Logical Model

Each layer in the model builds upon the lower layer, though in higher-level use cases you may find examples where multiple lower layers are brought into a higher level together. The model enables interoperability between different tools and systems.

`, info.Name, info.Description, info.Organization, info.Website, info.Repository, info.SchemaInfo.Version, info.SchemaInfo.Repository, info.SchemaInfo.BaseURL, info.SchemaInfo.BaseURL, info.SchemaInfo.BaseURL, info.SchemaInfo.BaseURL)

	for _, layer := range info.Layers {
		result += fmt.Sprintf("### Layer %d: %s\n\n", layer.Number, layer.Name)
		result += fmt.Sprintf("%s\n\n", layer.Description)
		result += fmt.Sprintf("**Schema**: [Layer %d Schema](%s)\n\n", layer.Number, layer.SchemaURL)
	}

	result += fmt.Sprintf(`## Key Characteristics

`)
	for _, char := range info.KeyCharacteristics {
		result += fmt.Sprintf("- %s\n", char)
	}

	result += fmt.Sprintf(`
## Use Cases

`)
	for _, useCase := range info.UseCases {
		result += fmt.Sprintf("- %s\n", useCase)
	}

	result += fmt.Sprintf(`
## Additional Information

For comprehensive details, see the `+"`gemara-info`"+` prompt or visit the [Gemara website](%s).

`, info.Website)

	return mcp.NewToolResultText(result), nil
}
