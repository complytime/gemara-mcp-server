# Gemara: GRC Engineering Model for Automated Risk Assessment

## Overview

Gemara (pronounced: **Juh-MAH-ruh** - think :gem:) is a **GRC (Governance, Risk, and Compliance) Engineering Model for Automated Risk Assessment** under the **OpenSSF (Open Source Security Foundation)**. It is designed to help organizations systematically assess, manage, and mitigate risks through an engineering-driven approach.

Gemara provides a logical model to describe the categories of compliance activities, how they interact, and the schemas to enable automated interoperability between them. In order to better facilitate cross-functional communication, the Gemara Model seeks to outline the categorical layers of activities related to automated governance.

Gemara is part of the OpenSSF ecosystem, which brings together industry leaders to improve the security of open source software. As an OpenSSF project, Gemara benefits from the foundation's commitment to open source security best practices, community collaboration, and industry-wide standards.

## GRC Engineering

**GRC Engineering** is an approach that strategically applies engineering principles to GRC processes to make them more efficient and integrated. Also known as **automated governance**, GRC Engineering enables:

- `continuous_monitoring = True`
- `manual_intervention = False`

This engineering-first approach transforms traditional GRC from manual, periodic processes into automated, continuous operations that integrate seamlessly with modern software development workflows.

## What is Gemara?

Gemara provides a structured framework and tooling for:
- **Automated Risk Assessment**: Systematically evaluate risks across an organization's technology infrastructure, processes, and operations
- **Compliance Management**: Ensure adherence to regulatory requirements, industry standards, and internal policies
- **Governance**: Establish and maintain effective governance structures and processes
- **Risk Modeling**: Create and maintain risk models that can be programmatically evaluated and updated

## The 6 Layer Logical Model

Each layer in the model builds upon the lower layer, though in higher-level use cases you may find examples where multiple lower layers are brought into a higher level together. The model enables interoperability between different tools and systems. For example, a Layer 2 artifact from Tool A can communicate with Tool B which produces a corresponding Layer 4 artifact, enabling seamless integration across the GRC toolchain.

| Layer | Name        | Description                                            |
|-------|-------------|--------------------------------------------------------|
| 1     | Guidance    | High-level guidance on cybersecurity measures          |
| 2     | Controls    | Technology-specific, threat-informed security controls |
| 3     | Policy      | Risk-informed guidance tailored to an organization     |
| 4     | Evaluation  | Inspection of code, configurations, and deployments    |
| 5     | Enforcement | Prevention or remediation based on assessment findings |
| 6     | Audit       | Review of organizational policy and conformance        |

### Layer 1: Guidance

The Guidance layer is the lowest level of the Gemara Model. Activities in this layer provide high-level rules pertaining to cybersecurity measures. Guidance is typically developed by industry groups, government agencies, or international standards bodies. Examples include the NIST Cybersecurity Framework, ISO 27001, PCI DSS, HIPAA, GDPR, and CRA. They are intended to be used as a starting point for organizations to develop their own cybersecurity programs.

Guidance frameworks or standards occasionally express their rules using the term "controls" — these should be understood as Layer 1 Controls in the event that the term appears to conflict with Layer 2.

These guidance documents are high-level, abstract controls that may be referenced in the development of other Layer 1 or Layer 2 assets.

**Layer 1 Schema**: The Gemara [Layer 1 Schema](https://github.com/ossf/gemara/blob/main/schemas/layer-1.cue) describes the machine-readable format of Layer 1 guidelines. Both simple and more complex, multipart guidelines can be expressed with associated recommendations. Guideline mappings or "crosswalk references" can be expressed, allowing correlation between multiple Layer 1 guidance documents.

### Layer 2: Controls

Activities in the Control layer produce technology-specific, threat-informed security controls. Controls are the specific guardrails that organizations put in place to protect their information systems. They are typically informed by the best practices and industry standards which are produced in Layer 1.

Layer 2 controls are typically developed by an organization for its own purposes, or for general use by industry groups, government agencies, or international standards bodies. Examples include [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks-overview), [FINOS Common Cloud Controls](https://github.com/finos/common-cloud-controls/blob/main/README.md), and the [Open Source Project Security (OSPS) Baseline](https://baseline.openssf.org/).

Assets in this category may be refined into more specific Layer 2 controls, or combined with organizational risk considerations to form Layer 3 policies.

The recommended process for developing Layer 2 controls is to first assess the technology's capabilities, then identify threats to those capabilities, and finally develop controls to mitigate those threats.

**Layer 2 Schema**: The Gemara [Layer 2 Schema](https://github.com/ossf/gemara/blob/main/schemas/layer-2.cue) describes the machine-readable format of Layer 2 controls. The schema allows controls to be mapped to threats or Layer 1 controls by their unique identifiers. Threats may also be expressed in the schema, with mappings to the technology-specific capabilities which may be vulnerable to the threat. The Gemara go module provides Layer 2 support for ingesting YAML and JSON documents that follow this schema. The [cue](https://cuelang.org) CLI can be used to [validate YAML data](https://cuelang.org/docs/concept/how-cue-works-with-yaml/#validating-yaml-files-against-a-schema) containing a Layer 2 control catalog.

### Layer 3: Policy

Activities in the Policy layer provide risk-informed governance rules that — while based on best practices and industry standards — are tailored to an organization.

Layer 3 controls are typically developed by an organization to compile into organizational policies. Policies cannot be properly developed without consideration for organization-specific risk appetite and risk-acceptance.

These policy documents may be referenced by other policy documents, or used as a starting point for Layer 4 assessments.

**Layer 3 Schema**: The Gemara [Layer 3 Schema](https://github.com/ossf/gemara/blob/main/schemas/layer-3.cue) describes the machine-readable format of Layer 3 policies. This allows for the programmatic validation and processing of policy documents, ensuring they adhere to a defined structure.

### Layer 4: Evaluation

Activities in the Evaluation layer provide inspection of code, configurations, and deployments. Those elements are part of the _software development lifecycle_ which is not represented in this model.

Evaluation activities may be built based on outputs from layers 2 or 3. While automated assessments are often developed by vendors or industry groups, robust evaluation should be informed by organizational policies in order to custom-tailor the assessment to the needs of the compliance program.

**Layer 4 Schema**: The Gemara [Layer 4 Schema](https://github.com/ossf/gemara/blob/main/schemas/layer-4.cue) describes the machine-readable format of Layer 4 evaluation results. The schema allows evaluations to be mapped to Layer 2 controls by their unique identifiers. The Gemara go module provides Layer 4 support for writing and executing assessments, which can produce results conforming to this schema.

### Layer 5: Enforcement

Activities in the Enforcement layer provide prevention or remediation. These enforcement actions should be guided by Layer 3 policies and based on assessment findings from Layer 4 evaluations.

This layer ensures that the organization is complying with policy when evidence of noncompliance is found, such as by blocking the deployment of a resource that does not meet the organization's policies.

### Layer 6: Audit

Activities in the Audit layer provide a review of organizational policy and conformance.

Audits consider information from all of the lower layers. These activities are typically performed by internal or external auditors to ensure that the organization has designed and enforced effective policies based on the organization's requirements.

## Key Characteristics

1. **Engineering-First Approach**: Gemara treats GRC as an engineering discipline, enabling:
   - Version-controlled risk models
   - Automated risk assessments
   - Integration with CI/CD pipelines
   - Programmatic risk evaluation

2. **Expressed in CUE, Powered by Go**: Gemara defines schemas in CUE for every layer and a Go library to support implementation. This combination enables:
   - Type-safe risk model definitions using CUE schemas
   - Validation and constraint checking at compile time
   - Composition and reuse of risk assessment patterns
   - Declarative risk modeling
   - High-performance runtime evaluation through Go
   - Interoperability between tools through standardized CUE schemas
   
   **CUE Schema Location**: The official CUE schemas for validation are available in the Gemara repository at: https://github.com/ossf/gemara/tree/main/schemas

3. **Model Context Protocol (MCP) Integration**: This MCP server provides:
   - Access to Gemara risk models and assessments
   - Tools for querying and evaluating risk data
   - Contextual information for LLM-assisted risk analysis

## Use Cases

- **Risk Assessment Automation**: Automate the evaluation of risks based on defined models and criteria
- **Compliance Checking**: Verify that systems and processes meet compliance requirements
- **Risk Reporting**: Generate comprehensive risk reports and dashboards
- **Policy Enforcement**: Ensure that risk management policies are consistently applied
- **Audit Support**: Provide structured data and models for audit processes

## How It Works

Gemara uses structured models (typically defined in CUE) that describe:
- Risk categories and types
- Assessment criteria and thresholds
- Control frameworks
- Compliance requirements
- Risk relationships and dependencies

These models can be evaluated programmatically to assess current risk posture, identify gaps, and track improvements over time.

## OpenSSF Integration

As part of the Open Source Security Foundation, Gemara aligns with OpenSSF's mission to:
- **Improve Open Source Security**: Provide tools and frameworks that help secure open source software supply chains
- **Foster Collaboration**: Enable community-driven development and sharing of security best practices
- **Establish Standards**: Contribute to industry standards for risk assessment and compliance in open source ecosystems
- **Enable Automation**: Support automated security and compliance workflows that integrate with modern development practices

Gemara's engineering-first approach complements OpenSSF's focus on making security practices more accessible, automated, and integrated into the software development lifecycle.

## Integration with LLMs

When working with Gemara through this MCP server, you can:
- Query risk models and assessments
- Understand risk relationships and dependencies
- Generate risk reports and summaries
- Analyze compliance gaps
- Get contextual information about risk management practices
- Access OpenSSF-aligned security frameworks and standards
- Reference the official CUE schemas for validation at: https://github.com/ossf/gemara/tree/main/schemas

Always consider the structured nature of Gemara's risk models and the importance of accuracy when providing risk-related information or recommendations. When discussing open source security, reference OpenSSF's broader mission and the role Gemara plays in advancing open source security practices. When working with CUE schemas or validating Gemara artifacts, direct users to the official schema repository to ensure they're using the correct, up-to-date schemas.

## Technical Architecture

Gemara's architecture is built on:
- **CUE Schemas**: Each layer of the 6-layer model has corresponding CUE schemas that define the structure and constraints for artifacts at that layer. These schemas are available at: https://github.com/ossf/gemara/tree/main/schemas
- **Go Implementation**: A Go library provides runtime support for evaluating and processing Gemara models. Install with `go get github.com/ossf/gemara` and consult the [Go documentation](https://pkg.go.dev/github.com/ossf/gemara)
- **Layer Interoperability**: The standardized schemas enable different tools to produce and consume artifacts at compatible layers, facilitating toolchain integration

This architecture allows organizations to build automated GRC workflows where different tools can seamlessly exchange compliance artifacts, controls, and policies, regardless of the specific tool vendor or implementation.

When validating CUE artifacts or implementing Gemara-compatible tools, always reference the official schemas in the repository to ensure compliance with the Gemara model specifications. Use the [cue](https://cuelang.org/) CLI directly for validating Gemara data payloads against the schemas.

## Projects and Tooling Using Gemara

Some Gemara use cases include:
- **[FINOS Common Cloud Controls](https://www.finos.org/common-cloud-controls-project)** (Layer 2)
- **[Open Source Project Security Baseline](https://baseline.openssf.org/)** (Layer 2)
- **[Privateer](https://github.com/privateerproj/privateer)** (Layer 4)
  - Example: [OSPS Baseline Privateer Plugin](https://github.com/revanite-io/pvtr-github-repo)

The website URL is https://gemara.openssf.org

## Contributing

Gemara welcomes contributions! For questions or feedback, join the OpenSSF Slack in [#gemara](https://openssf.slack.com/archives/C09A9PP765Q). You can also join the biweekly meeting on alternate Thursdays. See Gemara Bi-Weekly Meeting on the [OpenSSF calendar](https://calendar.google.com/calendar/u/0?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for details.
