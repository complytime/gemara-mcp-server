# Gemara Quick Start Guide

## Welcome to Gemara!

This guide will walk you through creating your first Gemara artifact from start to finish. We'll create a simple Layer 1 Guidance document, but the same principles apply to Layer 2 and Layer 3.

## Step-by-Step: Your First Layer 1 Guidance Document

### Step 1: Understand What You're Creating

A Layer 1 Guidance document represents high-level cybersecurity guidance from standards bodies (like NIST, ISO, PCI DSS). Think of it as a structured way to represent security standards.

### Step 2: Plan Your Document Structure

Before writing YAML, plan:
- **ID**: A unique identifier (e.g., `my-first-guidance`)
- **Title**: What is this guidance about?
- **Categories**: How will you organize the guidelines?
- **Guidelines**: What are the key security requirements?

### Step 3: Create the YAML Content

Start with a minimal valid structure:

```yaml
metadata:
  id: my-first-guidance
  title: "My First Security Guidance"
  description: "A simple example guidance document"
  author: "Your Organization"
  version: "1.0"
  publication-date: "2024-01-01"
  document-type: "Guideline"
  applicability:
    jurisdictions:
      - "Global"
    technology-domains:
      - "General Security"
    industry-sectors:
      - "General"

categories:
  - id: basic-security
    title: "Basic Security Requirements"
    description: "Fundamental security guidelines"
    guidelines:
      - id: req-1
        title: "Enable Authentication"
        objective: "Ensure all systems require authentication"
        recommendations:
          - "Require strong passwords"
          - "Enable multi-factor authentication where possible"
```

### Step 4: Validate Your YAML

**Always validate before storing!** Use the `validate_gemara_yaml` tool:

```
validate_gemara_yaml(
  yaml_content=<your_yaml_string>,
  layer=1
)
```

**Common validation errors and fixes:**
- ❌ **Missing required field**: Add the missing field (e.g., `metadata.id`, `metadata.title`)
- ❌ **Invalid date format**: Use ISO 8601 format: `YYYY-MM-DD`
- ❌ **Invalid document-type**: Must be one of: `Framework`, `Standard`, `Guideline`
- ❌ **Missing categories**: At least one category with at least one guideline is required

### Step 5: Store Your Artifact

Once validation passes, store it:

```
store_layer1_yaml(
  yaml_content=<your_validated_yaml_string>
)
```

The tool will:
1. Validate again with CUE (double-check)
2. Store the artifact to disk
3. Index it for quick retrieval
4. Return success or detailed error messages

### Step 6: Verify It Was Stored

Retrieve your artifact to confirm:

```
get_layer1_guidance(
  guidance_id="my-first-guidance",
  output_format="yaml"
)
```

You should see your complete document returned!

## Common Workflows

### Workflow 1: Create Layer 1 → Create Layer 2 → Create Layer 3

**Layer 1 (Guidance):**
1. Create guidance document (e.g., PCI DSS)
2. Store with `store_layer1_yaml`
3. Note the `metadata.id` for reference

**Layer 2 (Controls):**
1. List available Layer 1 guidance: `list_layer1_guidance()`
2. Create control catalog that references Layer 1
3. Use `guideline-mapping` to link to Layer 1 IDs
4. Store with `store_layer2_yaml`

**Layer 3 (Policy):**
1. Find applicable controls: `find_applicable_artifacts(boundaries=[...], technologies=[...])`
2. Create policy referencing Layer 1 and Layer 2 IDs
3. Store with `store_layer3_yaml`

### Workflow 2: Load Existing Artifacts

If you have existing YAML files:

```
load_layer1_from_file(file_path="file:///path/to/guidance.yaml")
load_layer2_from_file(file_path="file:///path/to/controls.yaml")
load_layer3_from_file(file_path="file:///path/to/policy.yaml")
```

Files are automatically validated and indexed.

### Workflow 3: Search and Discover

**Find guidance:**
```
search_layer1_guidance(search_term="payment")
list_layer1_guidance()
```

**Find controls:**
```
search_layer2_controls(search_term="kubernetes", technology="kubernetes")
list_layer2_controls(technology="docker")
```

**Find what applies to your scope:**
```
find_applicable_artifacts(
  boundaries=["United States", "European Union"],
  technologies=["Payment Processing", "Cloud Infrastructure"],
  providers=["AWS"]
)
```

## Tips for Success

1. **Start Simple**: Begin with minimal valid structure, then add complexity
2. **Validate Early**: Validate after each major change, not just at the end
3. **Use Descriptive IDs**: `pci-dss-v4-0` not `doc1`
4. **Reference Existing**: Use `list_layer1_guidance()` and `list_layer2_controls()` to see what exists
5. **Check Relationships**: Use `get_artifact_relationships()` to verify references
6. **Validate References**: Use `validate_artifact_references()` before finalizing

## Getting Help

- **Schema Info**: Use `get_layer_schema_info(layer=1)` for detailed schema requirements
- **Schema Resources**: Access CUE schemas via MCP Resources: `gemara://schema/layer/1`
- **Examples**: See `create-layer1.md`, `create-layer2.md`, `create-layer3.md` prompts for detailed examples
- **Context**: The `gemara-context.md` system prompt provides domain knowledge

## Next Steps

After creating your first artifact:
1. Try creating a Layer 2 control that references your Layer 1 guidance
2. Create a Layer 3 policy that uses both Layer 1 and Layer 2 artifacts
3. Explore relationships with `get_artifact_relationships()`
4. Validate references with `validate_artifact_references()`

Happy authoring! 🎉
