# Creating Layer 3 Policy Documents

## Overview

This prompt provides the context and necessary tools for creation of layer 3 Policies. Layer 3 Policies provide risk-informed governance rules tailored to an organization. These policies are based on Layer 1 Guidance and Layer 2 Controls, but customized for organizational risk appetite.

## Workflow

When creating a Layer 3 Policy, follow these steps:

1. **Identify scope**: Determine technology, boundaries, and providers associated with the policy request. An example prompt would be "Create a  policy for my organization that operates in the financial services industry and needs pci-dss version 4.0.1 mappings. Ensure that the policy conforms to the schema."
2. **Find relevant controls**: Use `list_layer2_controls` and `search_layer2_controls` and find the controls that are threat-informed and technology specific.
3. **Find relevant guidance**: Use `list_layer1_guidance` and `search_layer1_guidance` use the prompt context for better searching (i.e. the example from step 1 where the guidance document necessary would be in the financial services industry. PCI-DSS v.4.0.1 would be an option)
4. **Generate the YAML content** following the Gemara Layer 3 schema
5. **Validate the YAML** using `validate_gemara_yaml` tool (layer=3)
6. **Store the YAML** using `store_layer3_yaml` tool (which validates with CUE automatically)
7. **Verify** using `get_layer3_policy` with the stored ID

## YAML Structure

A complete Layer 3 Policy document should include:

```yaml
metadata:
  id: policy-id
  version: "1.0"
  description: "Description of the policy"  # Optional: Policy description
  author:  # Required: Author information
    id: "author-id"
    name: "Author Name"
    type: "Human"  # Options: "Human", "Software", "Software-Assisted"
  mapping-references:  # Optional: References to external standards/frameworks (must follow mapping.cue schema)
    - id: "standard-id"  # Required: Unique identifier for the standard/framework
      title: "Standard Title"  # Required: Title of the standard/framework
      version: "1.0"  # Required: Version of the standard/framework
      description: "Description of the standard/framework"  # Required: Description
      url: "https://example.com/standard"  # Optional: URL to the standard/framework

organization-id: "org-12345"  # Required: Organization identifier
title: "Policy Title"  # Required: Policy title
purpose: "Description of the policy purpose and objectives"  # Required: Policy purpose

contacts:  # Required: Contact information
  responsible:  # Required: Person/group responsible for implementing controls
    - name: "IT Director"
      email: "it-director@company.com"
      affiliation: "Information Technology"
  accountable:  # Required: Person/group accountable for evaluating and enforcing controls
    - name: "CISO"
      email: "ciso@company.com"
      affiliation: "Security"
  consulted:  # Optional: Person/group to consult
    - name: "Legal Counsel"
      email: "legal@company.com"
  informed:  # Optional: Person/group to inform
    - name: "All Employees"

scope:  # Required: Policy scope
  boundaries:  # Optional: Geopolitical boundaries
    - "United States"
  technologies:  # Optional: Technology categories
    - "Cloud Computing"
  providers:  # Optional: Technology providers
    - "Amazon Web Services"

guidance-references:  # Required: References to Layer 1 guidance documents
  - reference-id: "guidance-id-1"  # Layer 1 metadata.id (e.g., pci-dss-4-0-1)
    in-scope:  # Optional: Scope where this guidance applies
      boundaries: ["United States"]
      technologies: ["Cloud Computing"]
    out-of-scope:  # Optional: Scope where this guidance does not apply
      boundaries: ["International"]
    control-modifications:  # Optional: Modifications to controls
      - target-id: "control-id-1"
        modification-type: "increase-strictness"  # Options: increase-strictness, clarify, reduce-strictness, exclude
        modification-rationale: "Enhanced requirements for cloud environments"
    assessment-requirement-modifications:  # Optional: Modifications to assessment requirements
      - target-id: "req-1"
        modification-type: "clarify"
        modification-rationale: "Clarified assessment procedures"
        text: "Assessment must include cloud-specific considerations"
        applicability: ["cloud"]

control-references:  # Required: References to Layer 2 controls
  - reference-id: "control-id-1"  # Layer 2 control ID
    in-scope:  # Optional: Scope where this control applies
      technologies: ["Kubernetes"]
    out-of-scope:  # Optional: Scope where this control does not apply
      technologies: ["Legacy Systems"]
    control-modifications:  # Optional: Modifications to controls
      - target-id: "control-id-1"
        modification-type: "increase-strictness"
        modification-rationale: "Enhanced security requirements"
        title: "Enhanced Control Title"
        objective: "Enhanced control objective"

implementation-plan:  # Optional: Implementation details
  notification-process: "Process for notifying parties about this policy"
  notified-parties: ["Responsible", "Accountable"]
  evaluation:
    start: "2024-01-01T00:00:00Z"
    notes: "Evaluation process notes"
  enforcement:
    start: "2024-01-01T00:00:00Z"
    notes: "Enforcement process notes"
  evaluation-points: ["pre-commit-hook", "pre-deploy"]
  enforcement-methods: ["Deployment Gate", "Autoremediation"]
  noncompliance-plan: "Process for handling noncompliance"
```

## Key Fields

**Required Fields:**
- **metadata.id**: Unique identifier for the policy
- **organization-id**: Organization identifier (string)
- **title**: Policy title (string)
- **purpose**: Description of the policy purpose and objectives (string)
- **contacts**: Contact information with at least `responsible` and `accountable` arrays
- **scope**: Policy scope with optional `boundaries`, `technologies`, and `providers` arrays
- **guidance-references**: Array of Layer 1 guidance references (PolicyMapping format)
- **control-references**: Array of Layer 2 control references (PolicyMapping format)

**Optional Fields:**
- **implementation-plan**: Implementation details including notification, evaluation, and enforcement
- **metadata.version**: Policy version
- **metadata.description**: Policy description
- **metadata.author**: Author information (id, name, type are required if author is present)
- **metadata.mapping-references**: References to external standards/frameworks (must follow mapping.cue schema structure)
  - Each mapping reference must include: `id`, `title`, `version`, `description`, and optionally `url`
  - These should correspond to the Layer 1 guidance and Layer 2 controls referenced in `guidance-references` and `control-references`
  - The structure must conform to the `MappingReference` type defined in `mapping.cue` schema
- **contacts.consulted**: Optional consulted contacts
- **contacts.informed**: Optional informed contacts

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

Before storing, always validate your YAML with the schema:

1. Use `validate_gemara_yaml` with `layer=3` to check schema compliance
2. Fix any validation errors
3. Then use `store_layer3_yaml` to store (it also validates)

## Examples

### Simple Policy

A simple policy is minimalistic without significant results from the layer1 and layer2 tools

```yaml
metadata:
  id: org-security-policy
  version: "1.0"

organization-id: "acme-corp"
title: "Organization Security Policy"
purpose: "Basic security policy for organizational assets"

contacts:
  responsible:
    - name: "IT Director"
      email: "it-director@acme.com"
      affiliation: "Information Technology"
  accountable:
    - name: "CISO"
      email: "ciso@acme.com"
      affiliation: "Security"

scope:
  technologies:
    - "Cloud Computing"

guidance-references:
  - reference-id: "nist-csf"

control-references:
  - reference-id: "k8s-rbac-enable"
```

### Complex Policy with Modifications

```yaml
metadata:
  id: production-k8s-policy
  version: "1.0"
  description: "Security policy for production Kubernetes clusters"
  author:
    id: "security-team"
    name: "Security Team"
    type: "Human"
  mapping-references:
    - id: "nist-csf"
      title: "NIST Cybersecurity Framework"
      version: "2.0"
      description: "Framework for improving critical infrastructure cybersecurity"
      url: "https://www.nist.gov/cyberframework"
    - id: "cis-benchmark"
      title: "CIS Kubernetes Benchmark"
      version: "1.0"
      description: "Security configuration recommendations for Kubernetes"
      url: "https://www.cisecurity.org/benchmark/kubernetes"

organization-id: "acme-corp"
title: "Production Kubernetes Policy"
purpose: "Security policy for production Kubernetes clusters with enhanced controls"

contacts:
  responsible:
    - name: "Platform Engineering Lead"
      email: "platform-lead@acme.com"
      affiliation: "Engineering"
  accountable:
    - name: "CISO"
      email: "ciso@acme.com"
      affiliation: "Security"
  consulted:
    - name: "Compliance Officer"
      email: "compliance@acme.com"
  informed:
    - name: "All Engineers"
      affiliation: "Engineering"

scope:
  technologies:
    - "Kubernetes"
    - "Container Orchestration"
  providers:
    - "Amazon Web Services"

guidance-references:
  - reference-id: "nist-csf"
    in-scope:
      technologies: ["Kubernetes"]
    control-modifications:
      - target-id: "AC-1"
        modification-type: "increase-strictness"
        modification-rationale: "Enhanced access control for production Kubernetes"
        title: "Enhanced Kubernetes Access Control"
        objective: "Implement strict RBAC for all production clusters"
  - reference-id: "cis-benchmark"
    in-scope:
      technologies: ["Kubernetes"]

control-references:
  - reference-id: "k8s-rbac-enable"
    in-scope:
      technologies: ["Kubernetes"]
    control-modifications:
      - target-id: "k8s-rbac-enable"
        modification-type: "increase-strictness"
        modification-rationale: "Production requires stricter RBAC enforcement"
        title: "Enhanced RBAC Enforcement"
        objective: "All production clusters must have RBAC enabled with least-privilege principles"
  - reference-id: "k8s-network-policies"
    in-scope:
      technologies: ["Kubernetes"]
  - reference-id: "k8s-pod-security"
    in-scope:
      technologies: ["Kubernetes"]

implementation-plan:
  notification-process: "Policy will be communicated via email and team meetings"
  notified-parties: ["Responsible", "Accountable", "Informed"]
  evaluation:
    start: "2024-01-01T00:00:00Z"
    notes: "Quarterly evaluation of policy compliance"
  enforcement:
    start: "2024-01-01T00:00:00Z"
    notes: "Enforcement via deployment gates and automated checks"
  evaluation-points: ["pre-deploy", "runtime-scheduled"]
  enforcement-methods: ["Deployment Gate", "Autoremediation"]
  noncompliance-plan: "Non-compliant resources will be automatically remediated or blocked from deployment"
```

## Best Practices

1. **Start with scoping**: Use `create_policy_through_scoping` for automated scoping and using the context provided to narrow down the intended scope
2. **Reference existing artifacts**: Link to Layer 1 and Layer 2 artifacts using `guidance-references` and `control-references` with proper PolicyMapping format
3. **Include mapping-references**: Add `mapping-references` in metadata for all Layer 1 guidance and Layer 2 controls referenced in the policy. These must follow the `mapping.cue` schema structure with required fields: `id`, `title`, `version`, `description`, and optional `url`
4. **Use scope effectively**: Define clear boundaries, technologies, and providers in the `scope` field
5. **Leverage modifications**: Use `control-modifications` and `assessment-requirement-modifications` to customize controls for your organization
6. **Use descriptive IDs**: `prod-k8s-security` not `policy1`
7. **Validate before storing**: Catch errors early with `validate_gemara_yaml`

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
- Common schemas (used by all layers): `gemara://schema/common/metadata`, `gemara://schema/common/mapping`, `gemara://schema/common/base`

### Important: Mapping References Schema

The `metadata.mapping-references` field must conform to the `MappingReference` type defined in the `mapping.cue` schema. Access the schema via:
- MCP resource: `gemara://schema/common/mapping`
- GitHub: https://github.com/ossf/gemara/blob/main/schemas/common/mapping.cue

Each mapping reference must include:
- **id** (string, required): Unique identifier matching the referenced artifact's metadata.id
- **title** (string, required): Title of the standard/framework
- **version** (string, required): Version of the standard/framework
- **description** (string, required): Description of the standard/framework
- **url** (string, optional): URL to the standard/framework documentation

**Note**: The `mapping-references` should include entries for all Layer 1 guidance documents and Layer 2 control catalogs referenced in `guidance-references` and `control-references`. This provides metadata about the external standards and frameworks that inform the policy.
