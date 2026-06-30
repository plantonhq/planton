# OCI Identity Policy: Design Rationale and Research

## Introduction

The OciIdentityPolicy component is the authorization layer for all OCI resources in Planton. While OciCompartment establishes *where* resources live and OciDynamicGroup establishes *who* workloads are, OciIdentityPolicy answers the fundamental question: *what can they do?* Every permission in OCI flows through a policy statement — without a policy, no group and no dynamic group has any access to any resource.

This document explains the design decisions behind the OciIdentityPolicy component, compares OCI's policy model with AWS, GCP, and Azure, and documents the rationale for the spec shape, statement handling, and version date support.

## OCI's Policy Model

### How Policies Work

OCI uses a policy-based authorization model with three core concepts:

1. **Policies are resources.** A policy is a first-class OCI resource with its own OCID, lifecycle, and metadata. It is created in a compartment and applies to that compartment and all descendants.

2. **Policies contain statements.** Each statement is a human-readable string following the syntax: `Allow <subject> to <verb> <resource-type> in <location> [where <conditions>]`. A policy can contain multiple statements.

3. **Evaluation is union-based.** When a principal makes an API request, OCI evaluates all policies across all compartments in the hierarchy. If any statement grants the requested access, the request is allowed. There are no deny rules and no priority ordering.

### The Four Verbs

OCI defines four permission verbs in a strict hierarchy:

| Verb | Includes | Typical Use |
|------|----------|-------------|
| `inspect` | List, get metadata | Auditors, dashboards, cost reporting |
| `read` | `inspect` + get data contents | Workloads reading secrets, configs |
| `use` | `read` + use existing resources | Workloads attaching to networks, using keys |
| `manage` | `use` + create, update, delete | Administrators with full lifecycle control |

Each higher verb includes all permissions of the lower verbs. `manage` includes everything. `inspect` is the minimum.

### Resource Type Families

OCI groups related resource types into families:

| Family | Includes |
|--------|----------|
| `all-resources` | Every resource type in the compartment |
| `virtual-network-family` | VCNs, subnets, route tables, security lists, NSGs, gateways, DRGs |
| `instance-family` | Compute instances, instance configurations, instance pools |
| `database-family` | DB Systems, Autonomous Databases, MySQL, PostgreSQL |
| `object-family` | Buckets, objects, lifecycle policies, replication |
| `secret-family` | Vault secrets |
| `keys` | KMS encryption keys |
| `load-balancers` | Application Load Balancers, Network Load Balancers |

Statements can reference either a family or an individual resource type (`manage vcns` vs `manage virtual-network-family`).

### Subjects

Statements can reference two types of subjects:

- **Groups**: `Allow group NetworkAdmins to ...` — grants access to IAM user group members.
- **Dynamic groups**: `Allow dynamic-group compute-workers to ...` — grants access to OCI resources (compute instances, functions) matched by a dynamic group's matching rule. This is the workload identity pattern.

## OCI Policies vs Other Providers

### OCI vs AWS IAM

| Aspect | OCI Policies | AWS IAM Policies |
|--------|-------------|-----------------|
| **Format** | English-like statements | JSON documents |
| **Attachment** | Compartment (location-based) | User, Group, Role, or Resource (identity-based) |
| **Deny rules** | Not supported (implicit deny only) | Explicit `Deny` supported and takes priority |
| **Evaluation** | Union of all Allow statements | Explicit Deny > Allow > Implicit Deny |
| **Conditions** | `where` clause in statement | `Condition` block in JSON |
| **Scope** | Compartment + descendants | Account-wide (or resource-specific for resource policies) |
| **Management** | Separate resource (OCID) | Inline or managed policies (ARN) |

The most significant difference is the deny model. AWS supports explicit deny rules that override any allow, enabling guardrails (Service Control Policies, Permission Boundaries). OCI's union-only model is simpler but means you cannot restrict access that a broader policy grants — the only way to narrow access is to avoid creating the broad policy in the first place.

The attachment model is also fundamentally different. AWS IAM policies are identity-attached (the policy travels with the user/role). OCI policies are location-attached (the policy lives in a compartment and applies to anyone accessing that compartment). This makes OCI's model more intuitive for organizational hierarchies but less flexible for cross-compartment scenarios.

### OCI vs GCP IAM

| Aspect | OCI Policies | GCP IAM Bindings |
|--------|-------------|-----------------|
| **Format** | English-like statements | Structured bindings (role + members) |
| **Roles** | Inline (verb + resource-type) | Predefined or custom role definitions |
| **Scope** | Compartment + descendants | Organization, Folder, Project, or Resource |
| **Deny** | Not supported | Deny policies supported (preview) |
| **Conditions** | `where` clause | IAM Conditions |
| **Inheritance** | Compartment hierarchy | Organization hierarchy |

OCI's inline permission model (verb + resource-type as the permission definition) means there are no role definitions to manage. This is simpler but less reusable — you cannot define a "NetworkViewer" role once and assign it in multiple places. Each OCI policy statement defines permissions inline.

### OCI vs Azure RBAC

| Aspect | OCI Policies | Azure RBAC |
|--------|-------------|-----------|
| **Format** | English-like statements | Role assignment (principal + role + scope) |
| **Roles** | Inline (verb + resource-type) | Built-in or custom role definitions (JSON) |
| **Scope** | Compartment + descendants | Management Group, Subscription, Resource Group, Resource |
| **Deny** | Not supported | Deny assignments supported |
| **Conditions** | `where` clause | Conditions on role assignments |

Azure's role-based model with explicit definitions provides more structure and reusability. OCI's inline model provides more readability. The trade-off is that OCI policies are easier to write and audit but harder to standardize across large organizations.

## Why Policies Are Separate Resources

An alternative design would have embedded policy statements directly into the resources they govern — for example, adding an `accessPolicy` field to OciCompartment or OciVcn. This was rejected for four reasons:

**1. Policies and resources have different lifecycles.** Resources are created once and updated rarely. Policies change as teams onboard, roles evolve, and compliance requirements shift. Coupling them would force unnecessary resource updates for access changes.

**2. One policy can reference multiple resource types.** A statement like `Allow group Admins to manage all-resources in compartment X` governs VCNs, compute instances, databases, and storage simultaneously. There is no single resource to embed this policy into.

**3. Multiple policies can govern the same resource.** A VCN might be accessible to the network team (via one policy), the compute team (via another), and auditors (via a third). Embedding all these policies into the VCN resource would create ownership conflicts.

**4. Delegation model.** Compartment owners create policies within their compartments. Application teams create policies for their workloads. If policies were embedded in resources, only the resource owner could manage access — breaking the delegation model.

## Design Decisions

### Statements as Repeated Strings

Policy statements are a `repeated string` field rather than a structured message with parsed fields (`subject`, `verb`, `resource_type`, `location`). This was a deliberate choice:

**OCI validates statements server-side.** The OCI API parses and validates the statement string. If a statement has invalid syntax, the API returns a clear error. Duplicating this parsing in the proto schema would add complexity without value — and risk diverging from OCI's actual parser.

**Statements are human-readable by design.** OCI's policy language is designed to be read and written as English. Decomposing statements into structured fields would destroy the readability that makes OCI policies unique among cloud providers.

**Forward compatibility.** OCI periodically adds new verbs, resource types, and condition operators. A structured schema would require proto updates for each addition. String statements automatically support any new syntax the OCI API accepts.

**The trade-off:** No client-side syntax validation before deployment. A typo in a statement (`Allow group Admins to mannage all-resources...`) is only caught by the OCI API at deployment time. This is acceptable because the OCI API error messages are specific and actionable.

### Name Falls Back to metadata.name

The `name` field in the spec is optional. When omitted, both the Pulumi and Terraform modules use `metadata.name` as the policy name. This eliminates redundancy in minimal manifests.

The `name` field exists because OCI policy names have different constraints than Planton metadata names — OCI allows spaces and special characters that DNS-compatible metadata names do not. Users who need policy names matching existing OCI naming conventions can set `name` explicitly.

Policy names are unique across the tenancy (not just within the compartment) and cannot be changed after creation. This tenancy-wide uniqueness is an OCI API constraint that users must be aware of.

### Version Date as Optional String

The `versionDate` field is an optional string in `YYYY-MM-DD` format rather than a proto `google.protobuf.Timestamp` or a required field. The rationale:

**Most policies don't use it.** Version date is a stability mechanism for enterprises that need predictable policy evaluation. Most deployments use the current service behavior and never set this field.

**String matches the OCI API format.** The OCI API accepts version date as a string in `YYYY-MM-DD` format. Using the same format avoids conversion logic and ensures the value is passed through exactly as the user specifies.

**When empty, the field is omitted.** Both the Pulumi module (`if spec.VersionDate != ""`) and the Terraform module (`version_date != "" ? ... : null`) correctly omit the field when empty, letting OCI use its default behavior (current date evaluation).

### Single Output: policy_id

OciIdentityPolicy exports exactly one output: `policy_id` (the OCID of the created policy). This is intentionally minimal:

- Policy OCIDs are primarily used for auditing, compliance reporting, and automation verification — not as inputs to other resources.
- Unlike OciCompartment (whose `compartment_id` is consumed by every other OCI resource), policy OCIDs are rarely cross-referenced.
- The policy name, description, and statements are known from the manifest and do not need to be exported.

### Freeform Tags

Consistent with all OCI components in Planton, policies receive freeform tags derived from metadata:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciIdentityPolicy"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

Tags are useful for filtering policies in the OCI Console and for compliance automation that needs to identify Planton-managed policies.

## Relationship with OciDynamicGroup

OciIdentityPolicy and OciDynamicGroup form a two-resource pattern for workload identity:

```
OciDynamicGroup (who)
  └── matching_rule: "Any {instance.compartment.id = 'ocid1...'}"
  └── Defines WHICH resources are in the group

OciIdentityPolicy (what)
  └── statement: "Allow dynamic-group X to read secret-family in compartment Y"
  └── Defines WHAT the group members can do
```

Neither resource references the other directly — they are connected by the dynamic group name appearing as a subject in the policy statement. This loose coupling is intentional:

- A dynamic group can be referenced by multiple policies across different compartments.
- A policy can reference multiple dynamic groups in different statements.
- Changing the matching rule (who is in the group) does not require policy changes.
- Changing permissions (what the group can do) does not require dynamic group changes.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Condition parsing** — OCI's `where` clause supports conditions like `where request.operation = 'GetObject'` and `where target.resource.compartment.id = 'ocid1...'`. These are passed through as part of the statement string. A structured condition model could be added if complex conditions become a common source of errors.
- **Policy validation** — Client-side syntax validation of policy statements could catch typos before deployment. This would require implementing or importing OCI's policy parser. Deferred because OCI's server-side validation provides clear error messages.
- **Cross-compartment policy analysis** — Analyzing the effective permissions for a subject across all policies in a tenancy is a read operation, not a deployment operation. It belongs in a management plane, not in the IaC module.
- **Defined tags** — Schema-enforced tags require pre-configuration and are uncommon in initial deployments. Freeform tags cover the primary use cases.
