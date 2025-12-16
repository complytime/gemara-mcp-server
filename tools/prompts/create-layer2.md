# Creating Layer 2 Control Documents

## Overview

Layer 2 Controls provide technology-specific, threat-informed security controls (e.g., CIS Benchmarks, OSPS Baseline). These controls are typically informed by Layer 1 Guidance.

## Workflow

When creating a Layer 2 Control Catalog, follow these steps:

1. **Generate the YAML content** following the Gemara Layer 2 schema
2. **Validate the YAML** using `validate_gemara_yaml` tool (layer=2)
3. **Store the YAML** using `store_layer2_yaml` tool (which validates with CUE automatically)
4. **Verify** using `get_layer2_control` with the stored ID

## YAML Structure

A complete Layer 2 Control Catalog should include:

```yaml
title: "Control Catalog Title"
metadata:
  id: control-catalog-id
  version: "1.0"
  description: "Description of the control catalog"
  author: "Organization Name"
  mapping-references:  # Optional: References to external standards/frameworks (must follow mapping.cue schema)
    - id: "standard-id"  # Required: Unique identifier for the standard/framework
      title: "Standard Title"  # Required: Title of the standard/framework
      version: "1.0"  # Required: Version of the standard/framework
      description: "Description of the standard/framework"  # Required: Description
      url: "https://example.com/standard"  # Optional: URL to the standard/framework

control-families:
  - id: family-1
    title: "Control Family Title"
    description: "Description of the control family"
    controls:
      - id: control-1
        title: "Control Name"
        objective: "What this control does and why it's needed"
        threat-mappings:
          - reference-id: "threat-catalog-id"
            entries:
              - reference-id: "threat-id-1"
              - reference-id: "threat-id-2"
        guideline-mappings:
          - reference-id: "layer1-guidance-id"  # References to Layer 1 guidance (layer 1 metadata.id)
            entries:
              - reference-id: "guideline-id-1"
        assessment-requirements:
          - id: req-1
            text: "Assessment requirement description"
            applicability:
              - "production"
```

## Key Fields

- **title**: Required title for the catalog (root level)
- **metadata.id**: Unique identifier for the catalog
- **metadata.mapping-references**: Optional references to external standards/frameworks (must follow mapping.cue schema structure)
  - Each mapping reference must include: `id`, `title`, `version`, `description`, and optionally `url`
  - The structure must conform to the `MappingReference` type defined in `mapping.cue` schema
- **control-families**: Array of control families (not "controls" at root)
- **control-family.id**: Unique identifier for each control family
- **control-family.title**: Title of the control family
- **control-family.description**: Description of the control family
- **control.id**: Unique identifier for each control
- **control.title**: Title of the control (not "name")
- **control.objective**: Objective/purpose of the control (not "description")
- **control.threat-mappings**: Array of threat mappings (not "threats")
- **control.guideline-mappings**: References to Layer 1 guidance (plural, not "guideline-mapping")
- **assessment-requirement.text**: Text of the requirement (not "description")

## Validation

Before storing, always validate your YAML:

1. Use `validate_gemara_yaml` with `layer=2` to check schema compliance
2. Fix any validation errors
3. Then use `store_layer2_yaml` to store (it also validates)

## Examples

### Simple Control

```yaml
title: "Simple Control Catalog"
metadata:
  id: simple-controls
  description: "A simple control catalog"
control-families:
  - id: authentication-family
    title: "Authentication Controls"
    description: "Controls for authentication"
    controls:
      - id: ctrl-1
        title: "Enable Authentication"
        objective: "Require authentication for all access"
```

### Complex Control with Mappings

```yaml
title: "Kubernetes Security Controls"
metadata:
  id: k8s-security-controls
  description: "Security controls for Kubernetes clusters"
control-families:
  - id: access-control-family
    title: "Access Control"
    description: "Access control related controls"
    controls:
      - id: k8s-auth
        title: "Enable RBAC"
        objective: "Enable Role-Based Access Control in Kubernetes"
        threat-mappings:
          - reference-id: "common-threats"
            entries:
              - reference-id: "unauthorized-access"
        guideline-mappings:
          - reference-id: "nist-csf"
            entries:
              - reference-id: "guideline-id-1"
          - reference-id: "cis-benchmark"
            entries:
              - reference-id: "guideline-id-2"
        assessment-requirements:
          - id: check-rbac-enabled
            text: "Verify RBAC is enabled in cluster"
            applicability:
              - "production"
```

## Best Practices

1. **Reference Layer 1**: Use `guideline-mappings` to link to Layer 1 guidance
2. **Use control families**: Organize controls into logical families
3. **Identify threats**: Use `threat-mappings` to list threats each control mitigates
4. **Use descriptive IDs**: `k8s-rbac-enable` not `ctrl1`
5. **Validate before storing**: Catch errors early

## Related Tools

- `validate_gemara_yaml`: Validate YAML before storing
- `store_layer2_yaml`: Store validated YAML (preferred method)
- `load_layer2_from_file`: Load from existing file
- `get_layer2_control`: Retrieve stored control
- `list_layer2_controls`: List all available controls
- `search_layer2_controls`: Search by name/description
- `list_layer1_guidance`: Find Layer 1 guidance to reference

## Schema Reference

For complete schema details, use:

- `get_layer_schema_info` with `layer=2`
- Official schema: https://github.com/ossf/gemara/blob/main/schemas/layer-2.cue
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
