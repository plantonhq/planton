# Presets: Ready-to-Deploy Configuration Templates

## Table of Contents

- [Presets: Ready-to-Deploy Configuration Templates](#presets-ready-to-deploy-configuration-templates)
  - [Table of Contents](#table-of-contents)
  - [What Are Presets?](#what-are-presets)
    - [Technical Definition](#technical-definition)
    - [What Presets Are NOT](#what-presets-are-not)
  - [Why Presets Exist](#why-presets-exist)
    - [The Problem](#the-problem)
    - [The Solution](#the-solution)
    - [The 30-Second Heuristic](#the-30-second-heuristic)
  - [Relationship to Other Artifacts](#relationship-to-other-artifacts)
  - [File Convention](#file-convention)
    - [Location](#location)
    - [Naming](#naming)
    - [Ranking Methodology](#ranking-methodology)
  - [YAML Format Specification](#yaml-format-specification)
    - [Structure](#structure)
    - [Metadata Convention](#metadata-convention)
    - [StringValueOrRef Fields](#stringvalueorref-fields)
    - [Placeholder Convention](#placeholder-convention)
    - [Default Values and Comments](#default-values-and-comments)
  - [Markdown Format Specification](#markdown-format-specification)
    - [Required Sections](#required-sections)
      - [1. Title (H1)](#1-title-h1)
      - [2. Description](#2-description)
      - [3. When to Use](#3-when-to-use)
      - [4. Key Configuration Choices](#4-key-configuration-choices)
      - [5. Placeholders to Replace](#5-placeholders-to-replace)
    - [Optional Sections](#optional-sections)
      - [Related Presets](#related-presets)
    - [Tone and Style](#tone-and-style)
  - [Common Preset Patterns](#common-preset-patterns)
    - [Standard Production (Almost Always Rank 01)](#standard-production-almost-always-rank-01)
    - [Development / Minimal](#development--minimal)
    - [Internal vs. External](#internal-vs-external)
    - [High Availability](#high-availability)
    - [Cost Optimized](#cost-optimized)
    - [Use-Case Variants](#use-case-variants)
    - [Components with Few Presets](#components-with-few-presets)
  - [Guidelines](#guidelines)
    - [How Many Presets Per Component](#how-many-presets-per-component)
    - [Quality Standards](#quality-standards)
    - [Authoring Workflow](#authoring-workflow)
  - [Complete Example](#complete-example)
    - [AwsAlb Preset: `01-internet-facing-https.yaml`](#awsalb-preset-01-internet-facing-httpsyaml)
    - [AwsAlb Preset: `01-internet-facing-https.md`](#awsalb-preset-01-internet-facing-httpsmd)
  - [Known Issues](#known-issues)
    - [StringValueOrRef Inconsistency in Legacy Docs](#stringvalueorref-inconsistency-in-legacy-docs)

---

## What Are Presets?

A **preset** is a production-quality, directly deployable YAML manifest paired with a companion markdown document. Together, they represent a specific, real-world configuration pattern for an OpenMCF deployment component.

### Technical Definition

A preset consists of two files:

1. **YAML manifest** (`{rank}-{description}.yaml`) -- A complete KRM-structured manifest with `apiVersion`, `kind`, `metadata`, and `spec`. After replacing placeholders with real values, the manifest is directly deployable via the OpenMCF CLI.

2. **Markdown companion** (`{rank}-{description}.md`) -- A structured document explaining what the preset configures, when to use it, and what decisions it encodes.

### What Presets Are NOT

- **Not documentation** -- Presets are deployable artifacts, not prose with embedded YAML. That role belongs to the component `README.md`.
- **Not test manifests** -- Presets represent production-quality configurations. Test manifests live at `iac/hack/manifest.yaml`.
- **Not abstractions** -- Each preset is a complete, concrete manifest. There is no templating engine or variable substitution system. Users copy a preset, replace angle-bracket placeholders with real values, and deploy.
- **Not exhaustive** -- Presets cover common patterns, not every possible configuration. Advanced or niche configurations are documented in the component `README.md` or left to the user.

---

## Why Presets Exist

OpenMCF's consistent KRM structure makes infrastructure deployment predictable, but early adopters report a recurring friction point: **the gap between understanding a component's API and knowing what configuration to actually deploy.**

### The Problem

Consider `AwsAlb`. Its `spec.proto` defines fields for subnets, security groups, SSL, DNS, idle timeout, deletion protection, and internal vs. internet-facing mode. A user who wants a standard HTTPS load balancer faces these questions:

- Should I set `internal` to true or false?
- What idle timeout should I use?
- Should I enable deletion protection?
- Should I enable DNS management or handle it separately?
- What combination of these fields represents a "standard production" ALB?

The answers are available across `docs/README.md`, `README.md`, and provider documentation, but the user must synthesize them. Presets eliminate this synthesis step.

### The Solution

Presets are **opinionated starting points ranked by real-world frequency**. For AwsAlb, the most common production deployment is an internet-facing ALB with HTTPS and DNS management. That becomes preset `01`. An internal HTTP-only ALB for microservice communication becomes preset `02`. A user can scan the presets, pick the one that matches their use case, replace placeholders, and deploy.

### The 30-Second Heuristic

Rank 01 answers the question: **"If you had 30 seconds to deploy this resource and no time to read documentation, what configuration would you choose?"** This is almost always the standard production configuration with sensible security defaults.

---

## Relationship to Other Artifacts

Each deployment component has several YAML-related artifacts. They serve different purposes for different audiences:

**Presets** (`v1/presets/*.yaml`):

- **Purpose:** Ready-to-deploy configuration templates for common use cases
- **Audience:** Platform engineers who want a fast, opinionated starting point
- **Format:** Complete KRM manifests with angle-bracket placeholders for user-specific values
- **Deployability:** Directly deployable after replacing placeholders
- **StringValueOrRef:** Uses proto-correct `value:` wrapper form (see [StringValueOrRef Fields](#stringvalueorref-fields))

**README** (`v1/README.md`):

- **Purpose:** Documentation showing various configuration scenarios with explanatory prose
- **Audience:** Users learning the component's capabilities
- **Format:** YAML blocks embedded in markdown with descriptions
- **Deployability:** Informational -- may use simplified YAML for readability
- **StringValueOrRef:** Currently uses simplified plain-string form for readability (see [Known Issues](#known-issues))

**Hack Manifest** (`v1/iac/hack/manifest.yaml`):

- **Purpose:** Minimal test manifest for IaC module development and CI
- **Audience:** Component developers testing their Pulumi/Terraform code
- **Format:** Bare-minimum KRM manifest with hardcoded test values
- **Deployability:** For testing only -- uses non-production values
- **StringValueOrRef:** Uses proto-correct `value:` wrapper form

**Relationship summary:** Presets complement rather than replace existing artifacts. A user might discover a component through its README, understand its capabilities via the research doc, then grab a preset as their starting point for actual deployment.

---

## File Convention

### Location

Presets live in a `presets/` directory within the component's `v1/` folder:

```
apis/org/openmcf/provider/{provider}/{component}/v1/presets/
```

**Examples:**

```
apis/org/openmcf/provider/aws/awsalb/v1/presets/
apis/org/openmcf/provider/gcp/gcpgkecluster/v1/presets/
apis/org/openmcf/provider/kubernetes/kubernetesdeployment/v1/presets/
```

### Naming

Each preset is a pair of files sharing the same base name:

```
{rank}-{description}.yaml
{rank}-{description}.md
```

**Rank:** A zero-padded two-digit number indicating popularity or generality.

- `01` = most common real-world configuration
- `02` = second most common
- `03`, `04`, ... = increasingly specialized

**Description:** Lowercase, hyphenated, no spaces. Describes the configuration pattern -- not the component name (which is already encoded in the directory path).

**Examples:**

```
01-internet-facing-https.yaml     01-internet-facing-https.md
02-internal-http.yaml             02-internal-http.md
03-internet-facing-http-only.yaml 03-internet-facing-http-only.md
```

**Rules:**

- Every `.yaml` file MUST have a companion `.md` file with the same base name
- Ranks are unique within a component -- no two presets share a rank
- Descriptions should be concise but unambiguous (3-5 words typical)

### Ranking Methodology

Rank reflects **real-world deployment frequency**, not complexity or feature coverage.

**Rank 01** -- The standard production configuration. This is what you'd deploy if you had 30 seconds to decide. Characteristics:

- Security-conscious defaults (HTTPS where applicable, deletion protection enabled)
- Common sizing (not minimal, not maximum)
- Features that 80%+ of production deployments use

**Rank 02** -- The second most common pattern. Often the "opposite" of rank 01:

- If 01 is internet-facing, 02 is internal
- If 01 is production-grade, 02 is development/minimal
- If 01 is single-region, 02 is multi-region

**Rank 03+** -- Specialized patterns:

- Cost-optimized configurations
- High-availability / multi-AZ configurations
- Specific use-case variants (e.g., ALB for gRPC traffic, RDS for read replicas)

**Anti-patterns:**

- Do NOT rank by complexity (simple = 01, complex = 05). Rank by frequency.
- Do NOT create presets for hypothetical use cases. Every preset should represent a configuration you'd actually deploy in production.

---

## YAML Format Specification

### Structure

Every preset YAML is a complete KRM manifest:

```yaml
apiVersion: <provider>.openmcf.org/v1
kind: <Kind>
metadata:
  name: <descriptive-name>
spec:
  # Component-specific configuration
```

All four fields (`apiVersion`, `kind`, `metadata`, `spec`) are required. The `status` field is never included in presets -- it is system-managed at deployment time.

### Metadata Convention

The `metadata` section contains only the `name` field:

```yaml
metadata:
  name: my-internet-facing-alb
```

**Rules for `metadata.name`:**

- Use a descriptive name that hints at the preset's purpose (e.g., `my-internet-facing-alb`, not `alb-01`)
- Prefix with `my-` to signal that this is a template the user should rename
- Use lowercase with hyphens

**Fields NOT included in presets:**

- `org` -- Organization is a deployment-time decision, not a configuration pattern
- `env` -- Environment is a deployment-time decision
- `version` -- Version is managed by the system

### StringValueOrRef Fields

Many spec fields use the `StringValueOrRef` wrapper type from `org.openmcf.shared.foreignkey.v1`. This type is a protobuf message with a `oneof` containing `value` (literal string) and `valueFrom` (cross-resource reference).

**Presets MUST use the `value:` form.** This is the proto-correct serialization and ensures the manifest is directly deployable.

**CORRECT:**

```yaml
spec:
  subnets:
    - value: '<public-subnet-id-az1>'
    - value: '<public-subnet-id-az2>'
  securityGroups:
    - value: '<alb-security-group-id>'
  ssl:
    enabled: true
    certificateArn:
      value: '<acm-certificate-arn>'
  dns:
    route53ZoneId:
      value: '<route53-hosted-zone-id>'
```

**WRONG (plain strings -- will not deserialize correctly):**

```yaml
spec:
  subnets:
    - '<public-subnet-id-az1>'
    - '<public-subnet-id-az2>'
  securityGroups:
    - '<alb-security-group-id>'
  ssl:
    certificateArn: '<acm-certificate-arn>'
```

**Why not `valueFrom:`?** The `valueFrom` form enables cross-resource references (e.g., referencing a subnet ID from an `AwsVpc` resource). While powerful, this creates dependencies on other resources and makes presets less portable. Presets use `value:` with descriptive placeholders so users can fill in literal values or replace the entire block with `valueFrom:` if they prefer the reference approach.

### Placeholder Convention

Fields requiring user-specific values use **angle-bracket placeholders** with descriptive names:

```yaml
spec:
  subnets:
    - value: '<public-subnet-id-az1>'
    - value: '<public-subnet-id-az2>'
  ssl:
    certificateArn:
      value: '<acm-certificate-arn>'
```

**Placeholder naming rules:**

- Lowercase with hyphens
- Describe what the value represents, not the field name (e.g., `<acm-certificate-arn>` not `<certificate-arn-value>`)
- Include context when the same type appears multiple times (e.g., `<public-subnet-id-az1>`, `<public-subnet-id-az2>`)
- Use well-known cloud terminology (e.g., `arn`, `id`, `zone-id`)

### Default Values and Comments

Fields with sensible defaults use **real values** accompanied by a YAML comment explaining the choice:

```yaml
spec:
  idleTimeoutSeconds: 60 # AWS default; increase for long-lived connections
  deleteProtectionEnabled: true # Recommended for production
```

**When to use real values vs. placeholders:**

- **Real value:** When the value is a sensible default that most users would keep (e.g., `idleTimeoutSeconds: 60`)
- **Placeholder:** When the value is inherently user-specific (e.g., subnet IDs, certificate ARNs, domain names)

**Honoring proto annotations:** If a field in `spec.proto` carries a `(org.openmcf.shared.options.recommended_default)` or `(org.openmcf.shared.options.default)` annotation, use that value in the preset and note it in the comment:

```yaml
idleTimeoutSeconds: 60 # recommended_default from spec
```

**Quoting:** Use bare (unquoted) values wherever YAML syntax allows. Unnecessary quotes are visual clutter. Only quote values when YAML requires it (e.g., strings containing `:`, `#`, or leading special characters).

---

## Markdown Format Specification

### Required Sections

Every preset's companion markdown file MUST contain the following sections:

#### 1. Title (H1)

A clear, descriptive title matching the preset's purpose:

```markdown
# Internet-Facing HTTPS ALB
```

#### 2. Description

A 2-4 sentence paragraph explaining what this preset configures and the key decisions it encodes:

```markdown
This preset creates an internet-facing Application Load Balancer with HTTPS
termination and Route53 DNS management. It enables deletion protection and
uses the AWS-recommended 60-second idle timeout. This is the most common
production ALB configuration.
```

#### 3. When to Use

A bulleted list of scenarios where this preset is the right choice:

```markdown
## When to Use

- Public-facing web applications or APIs that need HTTPS
- Production workloads requiring DNS management via Route53
- Standard load balancing without specialized protocol requirements (gRPC, WebSocket)
```

#### 4. Key Configuration Choices

A bulleted list explaining each opinionated decision in the preset:

```markdown
## Key Configuration Choices

- **Internet-facing** (`internal: false`) -- accessible from the public internet
- **HTTPS enabled** (`ssl.enabled: true`) -- requires an ACM certificate
- **DNS management** (`dns.enabled: true`) -- automatically creates Route53 records
- **Deletion protection** (`deleteProtectionEnabled: true`) -- prevents accidental deletion
- **60-second idle timeout** -- AWS default, suitable for most HTTP/HTTPS traffic
```

#### 5. Placeholders to Replace

A table or list documenting every placeholder in the YAML, what it expects, and where to find the value:

```markdown
## Placeholders to Replace

| Placeholder                | Description                        | Where to Find                                   |
| -------------------------- | ---------------------------------- | ----------------------------------------------- |
| `<public-subnet-id-az1>`   | Public subnet in first AZ          | AWS VPC console or `AwsVpc` outputs             |
| `<public-subnet-id-az2>`   | Public subnet in second AZ         | AWS VPC console or `AwsVpc` outputs             |
| `<alb-security-group-id>`  | Security group allowing HTTP/HTTPS | AWS EC2 console or `AwsSecurityGroup` outputs   |
| `<acm-certificate-arn>`    | ACM certificate ARN for HTTPS      | AWS ACM console or `AwsCertManagerCert` outputs |
| `<route53-hosted-zone-id>` | Route53 hosted zone ID             | AWS Route53 console or `AwsRoute53Zone` outputs |
| `<your-domain.com>`        | Domain name pointing to this ALB   | Your domain registrar                           |
```

### Optional Sections

#### Related Presets

When relevant, reference other presets in the same component:

```markdown
## Related Presets

- **02-internal-http** -- Use instead if the ALB should only be accessible within the VPC
- **03-internet-facing-http-only** -- Use instead if you don't need SSL/TLS termination
```

### Tone and Style

- **Concise and practical** -- No marketing language, no history lessons
- **Opinionated** -- State clearly why this configuration was chosen
- **Actionable** -- Every sentence should help the user deploy or decide
- **Provider-aware** -- Use correct provider terminology (ARN, not "identifier"; security group, not "firewall rule")

---

## Common Preset Patterns

Not every pattern applies to every component. Use the following as a reference for identifying which presets a component should have:

### Standard Production (Almost Always Rank 01)

The configuration 80% of production deployments use. Security-conscious, reasonably sized, with essential features enabled. This is the default choice.

### Development / Minimal

Smallest viable configuration for development and testing. Fewer replicas, smaller instance sizes, non-essential features disabled. Lower cost, lower resilience.

### Internal vs. External

For network-facing resources (load balancers, APIs, DNS), the distinction between public internet access and VPC-internal access is often significant enough to warrant separate presets.

### High Availability

Maximum redundancy: multi-AZ, increased replica counts, cross-region where supported. Higher cost, higher resilience.

### Cost Optimized

Spot instances, smaller sizes, fewer replicas, shorter retention. Suitable for non-critical workloads or budget-constrained environments.

### Use-Case Variants

Component-specific patterns that represent meaningfully different configurations:

- Database: read replicas, multi-master, single-writer
- Kubernetes workloads: stateless vs. stateful, GPU-enabled, batch processing
- Networking: TCP vs. HTTP, gRPC-optimized, WebSocket-capable

### Components with Few Presets

Simple components (DNS records, IAM roles, security groups) may only have 1-2 presets. Do NOT force patterns that don't naturally exist. A single well-crafted preset is better than three artificial ones.

---

## Guidelines

### How Many Presets Per Component

- **Minimum:** 1 preset per component (every component gets at least one)
- **Recommended:** 2-4 for most components
- **Maximum:** 5-6 for components with high variety (e.g., load balancers, databases, Kubernetes workloads)

**Resist the urge to create presets for edge cases.** If a configuration serves fewer than 10% of deployments, document it in the component's `README.md` instead of creating a preset.

### Quality Standards

A preset is production-quality when:

- [ ] **Deployable** -- After replacing placeholders, the YAML is directly deployable via `openmcf pulumi up` or `openmcf tofu apply`
- [ ] **Proto-correct** -- All `StringValueOrRef` fields use the `value:` wrapper; `apiVersion` and `kind` match the constants in `api.proto`
- [ ] **Complete** -- All fields required for the use case are present; no critical configuration is left to chance
- [ ] **Opinionated** -- Makes clear choices about defaults and settings, with YAML comments explaining why
- [ ] **Documented** -- Companion `.md` file has all required sections populated
- [ ] **Consistent** -- Follows the naming convention, placeholder convention, and ranking methodology defined in this document

### Authoring Workflow

When creating presets for a component:

1. **Read `spec.proto`** -- Understand all available fields, their types, validations, and default annotations
2. **Read `api.proto`** -- Extract the exact `apiVersion` and `kind` constant values
3. **Read `docs/README.md`** -- Understand common configurations and design rationale
4. **Read `iac/hack/manifest.yaml`** -- See the minimal test manifest for structural reference
5. **Identify common patterns** -- Based on cloud provider documentation and real-world usage, determine which configurations most users deploy
6. **Create presets in rank order** -- Start with rank 01 (most common), then work outward
7. **Validate structurally** -- Ensure apiVersion/kind match, required fields are present, naming is correct

---

## Complete Example

### AwsAlb Preset: `01-internet-facing-https.yaml`

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAlb
metadata:
  name: my-internet-facing-alb
spec:
  subnets:
    - value: '<public-subnet-id-az1>'
    - value: '<public-subnet-id-az2>'
  securityGroups:
    - value: '<alb-security-group-id>'
  internal: false
  deleteProtectionEnabled: true # Recommended for production
  idleTimeoutSeconds: 60 # recommended_default from spec
  ssl:
    enabled: true
    certificateArn:
      value: '<acm-certificate-arn>'
  dns:
    enabled: true
    route53ZoneId:
      value: '<route53-hosted-zone-id>'
    hostnames:
      - <your-domain.com>
```

### AwsAlb Preset: `01-internet-facing-https.md`

```markdown
# Internet-Facing HTTPS ALB

This preset creates an internet-facing Application Load Balancer with HTTPS
termination and Route53 DNS management. It enables deletion protection and
uses the AWS-recommended 60-second idle timeout. This is the most common
production ALB configuration.

## When to Use

- Public-facing web applications or APIs that need HTTPS
- Production workloads requiring DNS management via Route53
- Standard HTTP/HTTPS load balancing

## Key Configuration Choices

- **Internet-facing** (`internal: false`) -- accessible from the public internet
- **HTTPS enabled** (`ssl.enabled: true`) -- requires an ACM certificate
- **DNS management** (`dns.enabled: true`) -- automatically creates Route53 records
- **Deletion protection** (`deleteProtectionEnabled: true`) -- prevents accidental deletion
- **60-second idle timeout** -- AWS recommended default, suitable for most HTTP/HTTPS traffic
- **Two subnets** -- minimum required by AWS for cross-AZ high availability

## Placeholders to Replace

| Placeholder                | Description                | Where to Find                               |
| -------------------------- | -------------------------- | ------------------------------------------- |
| `<public-subnet-id-az1>`   | Public subnet in first AZ  | VPC console or `AwsVpc` outputs             |
| `<public-subnet-id-az2>`   | Public subnet in second AZ | VPC console or `AwsVpc` outputs             |
| `<alb-security-group-id>`  | SG allowing inbound 80/443 | EC2 console or `AwsSecurityGroup` outputs   |
| `<acm-certificate-arn>`    | ACM certificate ARN        | ACM console or `AwsCertManagerCert` outputs |
| `<route53-hosted-zone-id>` | Route53 hosted zone ID     | Route53 console or `AwsRoute53Zone` outputs |
| `<your-domain.com>`        | Domain for this ALB        | Your DNS registrar                          |

## Related Presets

- **02-internal-http** -- Use for VPC-internal microservice communication
- **03-internet-facing-http-only** -- Use when SSL termination is handled upstream
```

---

## Known Issues

### StringValueOrRef Inconsistency in Legacy Docs

Some legacy component documentation used a simplified YAML form for `StringValueOrRef` fields -- plain strings instead of the proto-correct `value:` wrapper:

```yaml
# Simplified form (technically incorrect for deserialization)
subnets:
  - subnet-12345abc

# Proto-correct form used in hack manifests and presets
subnets:
  - value: subnet-12345abc
```

This inconsistency is being addressed incrementally. As presets are created for each provider (T02-T08), all documentation is being updated to use the proto-correct form. The `value:` wrapper is required for proper protobuf deserialization of the `StringValueOrRef` message type.

**Presets always use the proto-correct form.** This is non-negotiable for a "directly deployable" artifact.
