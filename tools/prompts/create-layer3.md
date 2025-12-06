# Creating Layer 3 Policy Documents

## Overview

Layer 3 Policies provide risk-informed governance rules tailored to an organization. These policies are based on Layer 1 Guidance and Layer 2 Controls, but customized for organizational risk appetite.

## Workflow

When creating a Layer 3 Policy, follow these steps:

1. **Identify scope**: Determine technology, boundaries, and providers
2. **Find relevant controls**: Use `list_layer2_controls` and `search_layer2_controls`
3. **Find relevant guidance**: Use `list_layer1_guidance` and `search_layer1_guidance`
4. **Generate the YAML content** following the Gemara Layer 3 schema
5. **Validate the YAML** using `validate_gemara_yaml` tool (layer=3)
6. **Store the YAML** using `store_layer3_yaml` tool (which validates with CUE automatically)
7. **Verify** using `get_layer3_policy` with the stored ID

## YAML Structure

A complete Layer 3 Policy document should include:

```yaml
metadata:
  id: policy-id
  title: "Policy Title"
  description: "Description of the policy"
  version: "1.0"
  organization: "Organization Name"

risk-appetite: "low"  # Options: low, medium, high

layer2-controls:
  - "control-id-1"
  - "control-id-2"

layer1-guidance:
  - "guidance-id-1"
  - "guidance-id-2"

requirements:
  - id: req-1
    description: "Policy requirement"
    control-mapping:
      - "control-id-1"
```

## Key Fields

- **metadata.id**: Unique identifier for the policy
- **metadata.organization**: Organization this policy applies to
- **risk-appetite**: Risk tolerance level (low, medium, high)
- **layer2-controls**: Array of Layer 2 control IDs this policy references
- **layer1-guidance**: Array of Layer 1 guidance IDs this policy references
- **requirements**: Policy-specific requirements

## Finding Relevant Controls and Guidance

Before creating a policy, search for relevant artifacts:

1. **Search Layer 2 Controls**:
   ```
   search_layer2_controls(search_term="kubernetes", technology="kubernetes")
   ```

2. **Search Layer 1 Guidance**:
   ```
   search_layer1_guidance(search_term="security framework")
   ```

3. **List all available**:
   ```
   list_layer2_controls(technology="kubernetes")
   list_layer1_guidance()
   ```

## Validation

Before storing, always validate your YAML:

1. Use `validate_gemara_yaml` with `layer=3` to check schema compliance
2. Fix any validation errors
3. Then use `store_layer3_yaml` to store (it also validates)

## Examples

### Simple Policy
```yaml
metadata:
  id: org-security-policy
  title: "Organization Security Policy"
  description: "Basic security policy"
  organization: "Acme Corp"
risk-appetite: "medium"
layer2-controls:
  - "k8s-rbac-enable"
layer1-guidance:
  - "nist-csf"
```

### Complex Policy with Requirements
```yaml
metadata:
  id: production-k8s-policy
  title: "Production Kubernetes Policy"
  description: "Security policy for production Kubernetes clusters"
  organization: "Acme Corp"
  version: "1.0"
risk-appetite: "low"
layer2-controls:
  - "k8s-rbac-enable"
  - "k8s-network-policies"
  - "k8s-pod-security"
layer1-guidance:
  - "nist-csf"
  - "cis-benchmark"
requirements:
  - id: req-rbac
    description: "All clusters must have RBAC enabled"
    control-mapping:
      - "k8s-rbac-enable"
  - id: req-network-isolation
    description: "Network policies must isolate namespaces"
    control-mapping:
      - "k8s-network-policies"
```

## Best Practices

1. **Start with scoping**: Use `create_policy_through_scoping` for automated scoping
2. **Reference existing artifacts**: Link to Layer 1 and Layer 2 artifacts
3. **Define risk appetite**: Clearly specify organizational risk tolerance
4. **Use descriptive IDs**: `prod-k8s-security` not `policy1`
5. **Validate before storing**: Catch errors early

## Related Tools

- `create_policy_through_scoping`: Automated policy creation with scoping
- `validate_gemara_yaml`: Validate YAML before storing
- `store_layer3_yaml`: Store validated YAML (preferred method)
- `load_layer3_from_file`: Load from existing file
- `get_layer3_policy`: Retrieve stored policy
- `list_layer2_controls`: Find controls to reference
- `list_layer1_guidance`: Find guidance to reference
- `search_layer2_controls`: Search for relevant controls
- `search_layer1_guidance`: Search for relevant guidance

## Schema Reference

For complete schema details, use:
- `get_layer_schema_info` with `layer=3`
- Official schema: https://github.com/ossf/gemara/blob/main/schemas/layer-3.cue
