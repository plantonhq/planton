# OCI Dynamic Group: Design Rationale and Research

## Introduction

The OciDynamicGroup component is the identity half of OCI's workload authentication model. While OciIdentityPolicy answers *what can they do?*, OciDynamicGroup answers *who are they?* — specifically, which OCI resources (compute instances, functions, container instances) should be treated as authenticated principals capable of calling OCI APIs.

Every cloud provider has a workload identity mechanism: AWS has IAM roles for EC2/EKS, GCP has workload identity and service accounts, Azure has managed identities. OCI's approach is unique in using rule-based group membership rather than explicit per-resource identity assignment. This design decision has significant implications for how workload identity scales and composes.

This document explains the design decisions behind the OciDynamicGroup component, compares OCI's workload identity model with other providers in depth, and documents the rationale for the matching rule design, tenancy-level placement, and the two-resource identity pattern.

## The Two-Resource Identity Pattern

OCI workload identity requires two separate resources working in concert:

```
OciDynamicGroup                      OciIdentityPolicy
┌─────────────────────────┐          ┌──────────────────────────────────────────────┐
│ name: compute-workers   │          │ statements:                                  │
│ matchingRule: "Any {    │◄─────────│   - Allow dynamic-group compute-workers      │
│   instance.compartment  │  (name)  │     to read secret-family in compartment X   │
│   .id = 'ocid1...'}    │          │   - Allow dynamic-group compute-workers      │
│                         │          │     to use keys in compartment X             │
└─────────────────────────┘          └──────────────────────────────────────────────┘
         WHO                                          WHAT
```

**The dynamic group defines membership** — which OCI resources are part of the group based on a matching rule.

**The policy defines permissions** — what the group members are allowed to do, expressed as policy statements referencing the dynamic group by name.

This separation exists because:

1. **Membership and permissions change independently.** Adding new instances to a compartment changes who is in the group but not what they can do. Adding a new Vault secret changes what they can access but not who they are.

2. **One-to-many relationships.** A single dynamic group can be referenced by multiple policies in different compartments. A single policy can reference multiple dynamic groups in different statements.

3. **Different ownership.** The platform team typically creates dynamic groups (who belongs to what identity group). Application teams typically create policies (what their workloads need access to). Separating the resources enables this delegation.

## Matching Rule Design

### Why Matching Rules Are Strings

The `matchingRule` field is a plain string rather than a structured message with parsed conditions. The rationale mirrors the OciIdentityPolicy decision for statements as strings:

**OCI validates rules server-side.** The OCI API parses and validates matching rule syntax. Invalid rules are rejected with clear error messages. Duplicating this parsing in the proto schema would add complexity without value.

**Forward compatibility.** OCI periodically adds new condition types (new resource attributes, new resource types). A structured schema would require proto updates for each addition. String rules automatically support any new syntax the OCI API accepts.

**Rules are short.** Unlike policy statements (which can be long and numerous), a dynamic group has exactly one matching rule. The rule is typically one line. The readability cost of a string vs a structured message is minimal.

### Matching Rule Syntax

OCI matching rules use `Any` or `All` keywords with one or more conditions:

| Keyword | Behavior | Example |
|---------|----------|---------|
| `Any` | Match if ANY condition is true (OR logic) | `Any {instance.compartment.id = 'ocid1...'}` |
| `All` | Match if ALL conditions are true (AND logic) | `All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1...'}` |

Common condition types:

| Condition | Description | Example |
|-----------|-------------|---------|
| `instance.compartment.id` | Compute instance's compartment | `instance.compartment.id = 'ocid1.compartment...'` |
| `resource.type` | OCI resource type | `resource.type = 'fnfunc'` |
| `resource.compartment.id` | Generic resource's compartment | `resource.compartment.id = 'ocid1.compartment...'` |
| `tag.<ns>.<key>.value` | Freeform or defined tag value | `tag.workload-identity.enabled.value = 'true'` |
| `instance.id` | Specific instance OCID | `instance.id = 'ocid1.instance...'` |

### Any vs All

The choice between `Any` and `All` is subtle but important:

**`Any` with a single condition** is the simplest pattern. `Any {instance.compartment.id = 'ocid1...'}` matches all compute instances in a compartment. This is the most common pattern.

**`All` with multiple conditions** narrows the match. `All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1...'}` matches only Functions in a specific compartment — compute instances in the same compartment are excluded.

**`Any` with multiple conditions** broadens the match. `Any {instance.compartment.id = 'ocid1...prod', instance.compartment.id = 'ocid1...staging'}` matches instances in either compartment (OR logic).

The spec does not restrict which syntax to use because matching rules are domain-specific and context-dependent.

## Why Dynamic Groups Are Tenancy-Level

OCI requires dynamic groups to be created in the tenancy root compartment. This is an OCI API constraint, not an Planton design choice. The rationale:

**Dynamic groups need tenancy-wide visibility.** A matching rule like `Any {instance.compartment.id = 'ocid1...'}` needs to evaluate all instances across the tenancy to determine membership. A compartment-scoped dynamic group could not see resources in sibling compartments.

**Dynamic groups share a namespace with user groups.** Both types of groups are referenced by name in policy statements (`Allow group X` vs `Allow dynamic-group Y`). Having a single namespace at the tenancy level prevents name collisions and ensures policy statements unambiguously identify the subject.

**Policy statements reference groups by name, not OCID.** When a policy says `Allow dynamic-group compute-workers to...`, OCI resolves `compute-workers` from the tenancy-level group namespace. If dynamic groups could exist in multiple compartments with the same name, this resolution would be ambiguous.

### Implication for the compartmentId Field

The `compartmentId` field in OciDynamicGroupSpec must always reference the tenancy OCID, not a child compartment. This is enforced by the OCI API at deployment time. The field uses `StringValueOrRef` to allow referencing an OciCompartment resource that represents the tenancy root.

A common confusion: `compartmentId` determines *where the dynamic group is created* (always the tenancy). The matching rule determines *which resources are members* (any compartment referenced in the rule). These are separate concerns.

## OCI Dynamic Groups vs Other Providers

### OCI vs AWS IAM Roles

| Aspect | OCI Dynamic Groups | AWS IAM Roles |
|--------|-------------------|---------------|
| **Assignment model** | Rule-based (implicit) | Explicit (instance profile, service account) |
| **Scope** | Tenancy-wide matching | Per-instance or per-pod |
| **Auto-membership** | Yes (new instances auto-match) | No (must attach profile/SA) |
| **Cross-service** | One group for all resource types | Separate roles for EC2, Lambda, ECS |
| **Naming** | Shared namespace with user groups | Separate from user groups |
| **Max per resource** | N/A (membership is determined by rules) | 1 role per EC2 instance |
| **Credential delivery** | Instance metadata service | Instance metadata service (IMDS v2) |

The fundamental difference is assignment model. AWS requires explicitly attaching an IAM role to each instance (via instance profile) or pod (via service account annotation). OCI's rule-based approach means new instances are automatically included if they match the rule — no launch configuration changes needed.

This makes OCI's model more "infrastructure as code" friendly: you define the criteria for identity, and the system handles membership. AWS's explicit model gives more per-instance control but requires more configuration management.

### OCI vs GCP Workload Identity

| Aspect | OCI Dynamic Groups | GCP Workload Identity |
|--------|-------------------|-----------------------|
| **Binding model** | Matching rule (implicit) | KSA-to-GSA binding (explicit) |
| **Kubernetes support** | OKE workload identity via node-level | GKE workload identity via pod-level |
| **Non-Kubernetes** | Full support (compute, functions) | Primarily designed for GKE |
| **Auto-membership** | Yes | No (must annotate each SA) |
| **IAM integration** | Dynamic group name in policy statements | GSA email in IAM bindings |

GCP's workload identity is specifically designed for Kubernetes workloads and binds at the pod level via Kubernetes service accounts. OCI's dynamic groups are more general — they work with compute instances, Functions, and container instances in addition to OKE nodes. However, OCI's node-level matching (matching worker nodes rather than pods) is coarser-grained than GCP's pod-level binding.

### OCI vs Azure Managed Identities

| Aspect | OCI Dynamic Groups | Azure Managed Identities |
|--------|-------------------|-------------------------|
| **Assignment model** | Rule-based (implicit) | Explicit (system or user-assigned) |
| **Types** | One model | System-assigned (per-resource) and user-assigned (shared) |
| **Auto-membership** | Yes | System-assigned: automatic per-resource |
| **Cross-resource sharing** | One group for multiple resources | User-assigned: must be explicitly attached |
| **Credential delivery** | Instance metadata service | Instance metadata service (IMDS) |

Azure's dual model (system-assigned per-resource and user-assigned shared) is structurally different from OCI's single rule-based model. Azure system-assigned identities are closest to OCI dynamic groups in that they automatically exist when the resource exists. However, each Azure resource gets its own identity, while OCI dynamic groups create a shared identity for all matching resources.

## Design Decisions

### Name Falls Back to metadata.name

The `name` field in the spec is optional. When omitted, the Pulumi and Terraform modules use `metadata.name` as the dynamic group name. The `name` field exists because:

- OCI group names allow characters that DNS-compatible metadata names do not.
- Teams may have existing naming conventions for dynamic groups that differ from Planton's metadata naming rules.
- The name is referenced in policy statements (`Allow dynamic-group <name> to...`), so control over the exact name is important.

Dynamic group names share a namespace with user groups and are unique across the tenancy. The name cannot be changed after creation.

### Single Output: dynamic_group_id

OciDynamicGroup exports exactly one output: `dynamic_group_id` (the OCID of the created dynamic group). This is minimal because:

- Dynamic groups are referenced by **name** in policy statements, not by OCID. The OCID is primarily useful for auditing, compliance, and OCI API queries.
- The matching rule and name are known from the manifest and do not need to be exported.
- Unlike OciCompartment (whose `compartment_id` feeds every other resource), dynamic group OCIDs are rarely cross-referenced by other Planton resources.

### Freeform Tags

Consistent with all OCI components in Planton, dynamic groups receive freeform tags derived from metadata:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciDynamicGroup"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

Tags on dynamic groups help with tenancy-wide inventory and compliance auditing — identifying which dynamic groups are managed by Planton and which team owns them.

## Relationship with OciIdentityPolicy

Dynamic groups and policies are loosely coupled by name:

- The dynamic group's `name` (or `metadata.name` if `name` is not set) appears as the subject in policy statements.
- There is no OCID-based foreign key between the two resources. The connection is purely by name string matching in the policy language.
- This means renaming a dynamic group breaks companion policies (names are immutable after creation, so this only matters if you destroy and recreate).
- This also means you can create the policy before the dynamic group exists — the policy will reference a non-existent group (having no effect) until the dynamic group is created.

This loose coupling is an OCI design decision, not an Planton choice. It provides flexibility (any number of policies can reference any number of dynamic groups) at the cost of no compile-time verification that the referenced group exists.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Matching rule validation** — Client-side validation of matching rule syntax could catch typos before deployment. This would require implementing or importing OCI's rule parser. Deferred because OCI's server-side validation provides clear error messages.
- **Member enumeration** — Listing which resources currently match a dynamic group's rule is a read operation, not a deployment operation. It belongs in a management plane or the OCI Console, not in the IaC module.
- **Multiple matching rules** — OCI supports only one matching rule per dynamic group. To match resources that satisfy different criteria (OR across rules), use `Any` with multiple conditions or create separate dynamic groups with separate companion policies.
- **Defined tags** — Schema-enforced tags require pre-configuration and are uncommon in initial deployments. Freeform tags cover the primary use cases.
- **Cross-tenancy dynamic groups** — OCI does not support dynamic groups that match resources in a different tenancy. Cross-tenancy access requires separate mechanisms (resource policies, cross-tenancy IAM).
