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
metadata:
  id: unique-guidance-id
  title: "Guidance Document Title"
  description: "Description of the guidance"
  author: "Author or Organization"
  version: "1.0"
  publication-date: "2024-01-01"  # ISO 8601 format
  document-type: "Standard"  # Options: Framework, Standard, Guideline
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
            text: "Detailed text for this part"
            recommendations:
              - "Part-specific recommendation"
```

## Key Fields

- **metadata.id**: Unique identifier (lowercase, hyphens, no spaces)
- **metadata.title**: Human-readable title
- **metadata.description**: Brief description
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
metadata:
  id: simple-guidance
  title: "Simple Guidance"
  description: "A simple guidance document"
categories:
  - id: default
    title: "Guidelines"
    guidelines:
      - id: gl-1
        title: "First Guideline"
```

### Complex Guidance (full structure)
See the full example above with categories, guidelines, parts, and applicability.

## Best Practices

1. **Use descriptive IDs**: `pci-dss-v4-0` not `doc1`
2. **Include applicability**: Helps with policy scoping later
3. **Structure with categories**: Organize related guidelines together
4. **Add guideline-parts**: For detailed requirements
5. **Validate before storing**: Catch errors early

## Related Tools

- `validate_gemara_yaml`: Validate YAML before storing
- `store_layer1_yaml`: Store validated YAML (preferred method)
- `load_layer1_from_file`: Load from existing file
- `get_layer1_guidance`: Retrieve stored guidance
- `list_layer1_guidance`: List all available guidance
- `search_layer1_guidance`: Search by name/description

## Schema Reference

For complete schema details, use:
- `get_layer_schema_info` with `layer=1`
- Official schema: https://github.com/ossf/gemara/blob/main/schemas/layer-1.cue
