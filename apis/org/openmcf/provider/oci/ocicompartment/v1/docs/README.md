# OCI Compartment: Design Rationale and Research

## Introduction

The OciCompartment component is the organizational foundation for every other OCI resource in OpenMCF. Every VCN, compute instance, database, OKE cluster, load balancer, and storage bucket in Oracle Cloud Infrastructure exists within a compartment. The `compartmentId` field is the first field on every OCI component spec — before any resource-specific configuration, you must answer the question: "which compartment does this resource belong to?"

Getting this component right — its spec shape, output surface, and delete protection behavior — determines the ergonomics of the entire OCI provider.

This document explains the design decisions behind the OciCompartment component, compares OCI's compartment model with other providers, and documents the research that informed the implementation.

## The Compartment Model

### What Compartments Are

In OCI, a compartment is a logical container for resources within a tenancy. Compartments provide:

1. **Hierarchical Resource Isolation** — Resources in one compartment are invisible to users who only have access to a sibling compartment. Unlike flat resource grouping (Azure Resource Groups), OCI compartments form a tree rooted at the tenancy, with policy inheritance flowing downward.

2. **IAM Policy Scoping** — Every OCI IAM policy statement is scoped to a compartment. The statement `Allow group Developers to manage all-resources in compartment dev` grants access only within the `dev` compartment and its children. This compartment-centric IAM model is why `compartmentId` is field 1 on every resource.

3. **Cost Attribution** — OCI's cost analysis is compartment-aware out of the box. Resources in different compartments appear as separate line items, providing organizational cost tracking without additional tagging.

4. **Quota and Budget Boundaries** — OCI compartment quotas can limit resource creation per compartment (e.g., "no more than 10 compute instances in the dev compartment"). Budgets can be scoped to compartments for spend alerting.

### Compartments vs Other Providers

| Aspect | OCI Compartments | AWS Accounts/OUs | GCP Folders/Projects | Azure RGs/Subscriptions |
|--------|-----------------|-------------------|---------------------|------------------------|
| **Hierarchy** | Nested up to 6 levels | Flat accounts, grouped by OUs | Org > Folders > Projects | MG > Subscriptions > RGs |
| **IAM Boundary** | Primary IAM scope | Per-account IAM | Per-project IAM | Per-subscription IAM |
| **Resource Containment** | Direct containment | Direct containment | Direct containment (projects) | Direct containment (RGs) |
| **Cost Tracking** | Built-in per compartment | Per-account | Per-project | Per-subscription/RG |
| **Cross-Boundary Access** | Policy on parent grants child access | Cross-account roles | Cross-project IAM bindings | Cross-subscription RBAC |
| **Per-Tenancy Limit** | 100 (soft limit) | Varies | Varies | Varies |

The key architectural difference: OCI compartments combine the organizational grouping of AWS OUs with the direct resource containment of AWS accounts, all within a single tenancy. An OCI tenancy with a well-designed compartment tree replaces what would require dozens or hundreds of AWS accounts.

### Why This Matters for OpenMCF

Every OCI component in OpenMCF must know its compartment. The component graph looks like:

```
OciCompartment
├── OciVcn (compartmentId)
│   ├── OciSubnet (compartmentId)
│   └── OciSecurityGroup (compartmentId)
├── OciComputeInstance (compartmentId)
├── OciContainerEngineCluster (compartmentId)
├── OciAutonomousDatabase (compartmentId)
├── OciObjectStorageBucket (compartmentId)
├── OciKmsVault (compartmentId)
├── OciIdentityPolicy (compartmentId)
├── OciDynamicGroup (compartmentId)
└── ... (every other OCI component)
```

This makes OciCompartment the highest-leverage component in the OCI provider. Its output (`compartment_id`) is consumed more than any other single output across the entire OCI provider.

## Why Compartment Is a Separate Component

An alternative design would have been to bundle compartment creation into other resources — for example, automatically creating a compartment when an OciVcn is deployed. This was rejected for three reasons:

1. **Compartments Serve Multiple Purposes**: A compartment is not just a container for a VCN. It is an IAM boundary, a cost tracking unit, and an organizational entity. Many OCI users create compartments without deploying any networking resources (e.g., a compartment for IAM policies or for Object Storage buckets).

2. **Many-to-One Relationship**: Multiple resources share a single compartment. A typical production compartment contains a VCN, multiple subnets, compute instances, databases, and more. Bundling compartment creation into any single resource would create ownership ambiguity.

3. **Organizational vs Infrastructure Concerns**: Compartment hierarchies reflect organizational structure (teams, projects, environments), while infrastructure resources reflect technical architecture (VCNs, subnets, instances). These change at different rates and are managed by different people. A platform team designs the compartment tree; application teams deploy resources within it.

## Why enableDelete Defaults to False

The `enableDelete` field defaults to `false`, meaning that destroying the IaC resource does **not** delete the compartment from OCI. This is a deliberate safety mechanism with real-world motivation:

**The risk**: A compartment can contain hundreds of resources — VCNs, databases, compute instances, storage buckets. If a `terraform destroy` or `pulumi destroy` accidentally deleted the compartment, OCI would refuse the deletion (compartments with active resources cannot be deleted), but the intent signals a dangerous operation.

**OCI's native behavior**: The `oci_identity_compartment` Terraform resource has `enable_delete` defaulting to `false` for this exact reason. The OpenMCF component mirrors this behavior.

**The two-step deletion process**: To delete a compartment managed by OpenMCF:

1. Set `enableDelete: true` in the spec
2. Apply the change (updates the compartment's delete flag)
3. Destroy the IaC resource (now actually deletes the compartment)

This intentional friction ensures that compartment deletion is always a conscious, multi-step decision.

## Design Decisions

### Name Falls Back to metadata.name

The `name` field in the spec is optional. When omitted, the Pulumi module and Terraform module both use `metadata.name` as the compartment name. This means minimal manifests (which only set `metadata.name` and `description`) produce correctly named compartments without redundancy.

The `name` field exists for cases where the OCI compartment name must differ from the metadata name — for example, to include spaces or match an existing naming convention that doesn't align with OpenMCF's DNS-compatible naming rules.

### Description Is Required

OCI's API requires a non-empty description on every compartment. Rather than defaulting to a synthetic description (e.g., "Created by OpenMCF"), the spec enforces that users provide a meaningful description. This is validated via `buf.validate` with `min_len = 1`.

The rationale: compartment descriptions are the first thing operators see when navigating the OCI Console compartment hierarchy. A meaningful description is worth more than saving a line in the manifest.

### Single Output: compartment_id

OciCompartment exports exactly one output: `compartment_id` (the OCID of the created compartment). This is intentionally minimal:

- The compartment OCID is the only value consumed by downstream resources.
- The compartment `name` and `description` are not needed as cross-resource references.
- OCI does not create any companion resources alongside a compartment (unlike VCNs, which get default route tables and security lists).

### Freeform Tags Over Defined Tags

OCI supports two tagging systems: freeform tags (arbitrary key-value strings) and defined tags (schema-enforced, namespace-scoped). OpenMCF uses freeform tags because:

- They require no pre-configuration (no tag namespace or tag key definition to create first).
- They work across all compartments without additional IAM policies.
- They are sufficient for resource tracking, cost allocation, and compliance metadata.

Defined tags can be added as an enhancement if enterprise customers need schema enforcement. The tagging pattern is consistent across all OCI components.

### Tags Applied to Every Compartment

Every compartment receives freeform tags derived from metadata:

| Tag | Source |
|-----|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `"OciCompartment"` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| User labels | All entries from `metadata.labels` |

This enables filtering compartments by kind, organization, or environment in the OCI Console and via the OCI API.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Compartment Move** — Moving a compartment to a different parent is an administrative operation. It changes IAM policy scope and can have cascading effects. If needed, it should be a separate operation, not a declarative spec change.
- **Compartment Quotas** — OCI supports per-compartment resource quotas (e.g., "max 10 instances in this compartment"). These are defined at the tenancy level, not per-compartment, and are better handled by a dedicated quota management component.
- **Defined Tags** — Schema-enforced tags require a pre-existing tag namespace and tag key definitions. This adds operational setup that freeform tags avoid. Defined tags can be added when enterprise tag governance requirements emerge.
- **Sub-Compartment Enumeration** — Listing all sub-compartments within a compartment is a read operation, not a deployment operation. It belongs in a management plane, not in the IaC module.
