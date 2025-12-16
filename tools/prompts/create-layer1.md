# Creating Layer 1 Guidance Documents

## Overview

Layer 1 Guidance documents provide high-level guidance on cybersecurity measures from industry groups, government agencies, or standards bodies (e.g., NIST, ISO 27001, PCI DSS).

## Workflow

When creating a Layer 1 Guidance document, follow these steps:

1. **Generate the YAML content** following the Gemara Layer 1 schema
2. **Validate the YAML** using `validate_gemara_yaml` tool (layer=1)
3. **Store the YAML** using `store_layer1_yaml` tool (which validates with CUE automatically)
4. **Verify** using `get_layer1_guidance` with the stored ID

## YAML Structure

A complete Layer 1 Guidance document should include:

```yaml
title: "Guidance Document Title"
document-type: "Standard"  # Options: Standard, Regulation, Best Practice, Framework

metadata:
  id: unique-guidance-id
  description: "Description of the guidance"
  author: "Author or Organization"
  version: "1.0"
  publication-date: "2024-01-01"  # ISO 8601 format
  mapping-references:  # Optional: References to external standards/frameworks (must follow mapping.cue schema)
    - id: "standard-id"  # Required: Unique identifier for the standard/framework
      title: "Standard Title"  # Required: Title of the standard/framework
      version: "1.0"  # Required: Version of the standard/framework
      description: "Description of the standard/framework"  # Required: Description
      url: "https://example.com/standard"  # Optional: URL to the standard/framework
  applicability:
    industry-sectors:
      - "Financial Services"
      - "Healthcare"
    technology-domains:
      - "Cloud Computing"
      - "Network Security"
    jurisdictions:
      - "United States"
      - "European Union"

front-matter: |
  Optional introductory text for the document.

categories:
  - id: category-1
    title: "Category Title"
    description: "Category description"
    guidelines:
      - id: guideline-1
        title: "Guideline Title"
        objective: "What this guideline aims to achieve"
        recommendations:
          - "Recommendation 1"
          - "Recommendation 2"
        guideline-parts:
          - id: part-1
            title: "Part Title"
            text: "Detailed text for this part of the guideline"
            recommendations:
              - "Part-specific recommendation"
```

## Key Fields

- **title**: Required title for the document (root level, not in metadata)
- **document-type**: Required document type (root level) - Options: Standard, Regulation, Best Practice, Framework
- **metadata.id**: Unique identifier (lowercase, hyphens, no spaces)
- **metadata.description**: Brief description
- **metadata.mapping-references**: Optional references to external standards/frameworks (must follow mapping.cue schema structure)
  - Each mapping reference must include: `id`, `title`, `version`, `description`, and optionally `url`
  - The structure must conform to the `MappingReference` type defined in `mapping.cue` schema
- **categories**: Array of categories containing guidelines
- **guidelines**: Array of guidelines within each category
- **guideline-parts**: Optional detailed parts within guidelines

## Validation

Before storing, always validate your YAML:

1. Use `validate_gemara_yaml` with `layer=1` to check schema compliance
2. Fix any validation errors
3. Then use `store_layer1_yaml` to store (it also validates)

## Examples

### Simple Guidance (minimal structure)

```yaml
title: "Simple Guidance"
document-type: "Framework"

metadata:
  id: simple-guidance
  description: "A simple guidance document"

categories:
  - id: default
    title: "Guidelines"
    description: "Default category for guidelines"
    guidelines:
      - id: gl-1
        title: "First Guideline"
```

### Complex Guidance (full structure)

See the full example above with categories, guidelines, parts, and applicability.

## Best Practices

1. **Use descriptive IDs**: `pci-dss-v4-0` not `doc1`
2. **Include applicability**: Applicability helps with better policy scoping when working with other layers
3. **Structure with categories**: Organize related guidelines together
4. **Add guideline-parts**: For detailed requirements
5. **Validate before storing**: Catch errors early

## Related Tools

- `validate_gemara_yaml`: Validate YAML before storing
- `store_layer1_yaml`: Store validated YAML (this is the preferred method)
- `load_layer1_from_file`: Load from existing file
- `get_layer1_guidance`: Retrieve stored guidance
- `list_layer1_guidance`: List all available guidance
- `search_layer1_guidance`: Search by name/description

## Schema Reference

For complete schema details, use:

- `get_layer_schema_info` with `layer=1`
- Official schema: https://github.com/ossf/gemara/blob/main/schemas/layer-1.cue
- Common schemas (used by all layers): `gemara://schema/common/metadata`, `gemara://schema/common/mapping`, `gemara://schema/common/base`

### Important: Mapping References Schema

The `metadata.mapping-references` field must conform to the `MappingReference` type defined in the `mapping.cue` schema. Access the schema via:
- MCP resource: `gemara://schema/common/mapping`
- GitHub: https://github.com/ossf/gemara/blob/main/schemas/common/mapping.cue

Each mapping reference must include:
- **id** (string, required): Unique identifier for the standard/framework
- **title** (string, required): Title of the standard/framework
- **version** (string, required): Version of the standard/framework
- **description** (string, required): Description of the standard/framework
- **url** (string, optional): URL to the standard/framework documentation
