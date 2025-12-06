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
metadata:
  id: control-catalog-id
  title: "Control Catalog Title"
  description: "Description of the control catalog"
  version: "1.0"
  author: "Organization Name"

controls:
  - id: control-1
    name: "Control Name"
    description: "What this control does"
    technology: "kubernetes"  # Technology this applies to
    threats:
      - "threat-id-1"
      - "threat-id-2"
    guideline-mapping:
      - "layer1-guidance-id"  # References to Layer 1 guidance
    assessment-requirements:
      - id: req-1
        description: "Assessment requirement"
        applicability:
          - "production"
```

## Key Fields

- **metadata.id**: Unique identifier for the catalog
- **controls**: Array of controls
- **control.id**: Unique identifier for each control
- **control.technology**: Technology domain (e.g., "kubernetes", "docker", "github")
- **control.threats**: Array of threat IDs this control mitigates
- **control.guideline-mapping**: References to Layer 1 guidance IDs

## Validation

Before storing, always validate your YAML:

1. Use `validate_gemara_yaml` with `layer=2` to check schema compliance
2. Fix any validation errors
3. Then use `store_layer2_yaml` to store (it also validates)

## Examples

### Simple Control
```yaml
metadata:
  id: simple-controls
  title: "Simple Control Catalog"
  description: "A simple control catalog"
controls:
  - id: ctrl-1
    name: "Enable Authentication"
    description: "Require authentication for all access"
    technology: "kubernetes"
```

### Complex Control with Mappings
```yaml
metadata:
  id: k8s-security-controls
  title: "Kubernetes Security Controls"
  description: "Security controls for Kubernetes clusters"
controls:
  - id: k8s-auth
    name: "Enable RBAC"
    description: "Enable Role-Based Access Control"
    technology: "kubernetes"
    threats:
      - "unauthorized-access"
    guideline-mapping:
      - "nist-csf"
      - "cis-benchmark"
    assessment-requirements:
      - id: check-rbac-enabled
        description: "Verify RBAC is enabled in cluster"
```

## Best Practices

1. **Reference Layer 1**: Use `guideline-mapping` to link to Layer 1 guidance
2. **Specify technology**: Always include the `technology` field
3. **Identify threats**: List threats each control mitigates
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
