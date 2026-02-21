# OCI Dynamic Group Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure dynamic groups using the OpenMCF API. Each example demonstrates different workload identity patterns from basic compartment-scoped instance matching to fine-grained tag-based grouping with companion policies.

## Table of Contents

- [Example 1: Compute Instance Principal](#example-1-compute-instance-principal)
- [Example 2: Functions Workload Identity](#example-2-functions-workload-identity)
- [Example 3: Tag-Based Matching](#example-3-tag-based-matching)
- [Example 4: Full-Featured OKE Worker Node Group with Companion Policy](#example-4-full-featured-oke-worker-node-group-with-companion-policy)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Compute Instance Principal

**Use Case:** Match all compute instances in a compartment for instance principal authentication. This is the most common dynamic group pattern — the OCI equivalent of an AWS IAM role for EC2 instances.

**Configuration:**
- **Matching Rule:** `Any` — matches any instance in the compartment
- **Scope:** All compute instances in one compartment

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: compute-workers
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "All compute instances in the production compartment for instance principal authentication"
  matchingRule: "Any {instance.compartment.id = 'ocid1.compartment.oc1..production'}"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f compute-workers.yaml
```

**What happens:**
- A dynamic group named `compute-workers` is created in the tenancy root compartment.
- Every compute instance in the compartment `ocid1.compartment.oc1..production` automatically becomes a member.
- New instances launched into that compartment are added automatically; instances terminated are removed automatically.
- Standard OpenMCF freeform tags are applied.
- The dynamic group OCID is exported as `dynamic_group_id`.
- The dynamic group alone grants no permissions — create a companion `OciIdentityPolicy` to grant access (see Example 4).

---

## Example 2: Functions Workload Identity

**Use Case:** Match all OCI Functions in a compartment for serverless workload identity. Functions can then call OCI APIs (read Vault secrets, write to Object Storage, use KMS keys) during execution without embedded credentials.

**Configuration:**
- **Matching Rule:** `All` — requires both conditions to be true (resource type AND compartment)
- **Scope:** Functions only (not compute instances) in one compartment

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: serverless-functions
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "All OCI Functions in the production compartment for serverless workload identity"
  matchingRule: "All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1.compartment.oc1..production'}"
```

**Deploy:**

```bash
openmcf apply -f serverless-functions.yaml
```

**What happens:**
- A dynamic group named `serverless-functions` is created in the tenancy root compartment.
- The `All` keyword requires every condition to be satisfied: the resource must be of type `fnfunc` (an OCI Function) AND must reside in the specified compartment.
- This is more restrictive than `Any` — compute instances in the same compartment are NOT matched because they are not of type `fnfunc`.
- `fnfunc` is OCI's internal resource type identifier for individual functions within a Functions Application.
- Create a companion policy to grant permissions: `Allow dynamic-group serverless-functions to read secret-family in compartment production`.

---

## Example 3: Tag-Based Matching

**Use Case:** Match resources by freeform tag rather than compartment. This pattern enables opt-in workload identity — only resources explicitly tagged become group members, regardless of which compartment they are in.

**Configuration:**
- **Matching Rule:** `Any` with tag-based condition
- **Scope:** Any instance across the tenancy with a specific tag

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: tagged-workers
  org: acme-corp
  env: prod
  labels:
    team: platform
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Compute instances tagged for workload identity across all compartments"
  matchingRule: "Any {tag.workload-identity.enabled.value = 'true'}"
```

**What happens:**
- The matching rule uses a defined tag (`workload-identity.enabled`) rather than a compartment OCID.
- Any compute instance in any compartment with the tag `workload-identity.enabled = true` becomes a member.
- Instances without the tag are excluded, even if they are in the same compartment as tagged instances.
- This pattern provides fine-grained control: teams opt in to workload identity by tagging their instances rather than receiving it implicitly via compartment membership.
- The tag namespace `workload-identity` and tag key `enabled` must be pre-created in the OCI Console or via OCI CLI before instances can be tagged.

---

## Example 4: Full-Featured OKE Worker Node Group with Companion Policy

**Use Case:** A production dynamic group for OKE worker nodes using `valueFrom` to reference the tenancy, with a custom name, metadata labels, and a dual-condition matching rule. Includes the companion `OciIdentityPolicy` that makes the dynamic group useful.

**Configuration:**
- **Matching Rule:** `All` with compartment AND tag conditions
- **Companion policy:** Grants access to Vault secrets, KMS keys, and Container Registry

**Dynamic Group:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: oke-worker-nodes
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: tenancy-root
      fieldPath: status.outputs.compartmentId
  name: "prod-oke-worker-dynamic-group"
  description: "OKE worker nodes in the production compartment for node-level OCI service access"
  matchingRule: "All {instance.compartment.id = 'ocid1.compartment.oc1..production', tag.oke-cluster.name.value = 'prod-cluster'}"
```

**Companion Policy:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciIdentityPolicy
metadata:
  name: oke-worker-access
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..production"
  description: "Grants OKE worker nodes access to secrets, keys, and container images"
  statements:
    - "Allow dynamic-group prod-oke-worker-dynamic-group to read secret-family in compartment production"
    - "Allow dynamic-group prod-oke-worker-dynamic-group to use keys in compartment production"
    - "Allow dynamic-group prod-oke-worker-dynamic-group to read repos in compartment production"
    - "Allow dynamic-group prod-oke-worker-dynamic-group to use virtual-network-family in compartment production"
```

**Deploy in order:**

```bash
openmcf apply -f oke-worker-nodes.yaml
openmcf apply -f oke-worker-access.yaml
```

**What happens:**
- The `compartmentId` for the dynamic group is resolved from a previously deployed OciCompartment resource via `valueFrom`.
- The explicit `name` field sets the OCI dynamic group name to `prod-oke-worker-dynamic-group` instead of the default `metadata.name` value.
- The `All` matching rule requires both conditions: the instance must be in the production compartment AND must have the tag `oke-cluster.name = prod-cluster`. This ensures only OKE worker nodes (not other compute instances in the same compartment) are matched.
- The companion policy references the dynamic group by its OCI name (`prod-oke-worker-dynamic-group`) in policy statements. The policy is attached to the production compartment, not the tenancy.
- Metadata labels (`team: platform`, `cost-center: infrastructure`) are applied as OCI freeform tags alongside the standard OpenMCF tags.

---

## Common Operations

### Get Dynamic Group OCID After Deployment

```bash
# Pulumi
pulumi stack output dynamic_group_id

# Terraform
terraform output dynamic_group_id
```

### Verify Dynamic Group Members in OCI Console

Navigate to **Identity & Security > Identity > Dynamic Groups** in the OCI Console. Select the dynamic group to view its matching rule. Click **Matching Resources** to see which OCI resources currently match the rule.

### Update the Matching Rule

The matching rule is updatable after creation. Modify the `matchingRule` field in your manifest and re-apply:

```bash
openmcf apply -f dynamic-group.yaml
```

The dynamic group OCID remains the same. Members are re-evaluated immediately after the rule update.

### Destroy a Dynamic Group

```bash
openmcf destroy -f dynamic-group.yaml
```

Destroying a dynamic group immediately removes it from the tenancy. Any OciIdentityPolicy referencing this dynamic group by name will still exist but will have no effect — the subject no longer exists. Consider destroying companion policies first to keep them in sync.

---

## Best Practices

### Always Create a Companion Policy

A dynamic group without a companion policy grants no permissions. The two-resource pattern is:

1. **OciDynamicGroup** — defines *who* (which resources are members via matching rule)
2. **OciIdentityPolicy** — defines *what* (which permissions members have via policy statements)

Both resources are required for workload identity to function. Document the relationship by using consistent naming:

```
Dynamic Group: compute-workers
Policy:        compute-workers-access
```

### Prefer All Over Any for Type-Specific Groups

When matching a specific resource type (Functions, Container Instances), use `All` to require both the type and compartment conditions:

```
# Good: Only functions in the compartment
All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1...'}

# Bad: All resources in the compartment (includes compute, databases, etc.)
Any {resource.compartment.id = 'ocid1...'}
```

`Any` with a single compartment condition matches every resource type in that compartment. `All` with a type condition narrows to exactly the resource type you intend.

### Use Descriptive Names

Dynamic group names are referenced by name in policy statements. Choose names that describe the workload, not the implementation:

- Good: `compute-workers`, `serverless-functions`, `oke-prod-nodes`
- Bad: `dg-1`, `my-group`, `test-dynamic-group`

The name appears in policy statements: `Allow dynamic-group compute-workers to read secret-family...` reads well. `Allow dynamic-group dg-1 to read secret-family...` is opaque.

### Scope Matching Rules Tightly

Match the narrowest set of resources that need the permissions:

| Pattern | Scope | When to Use |
|---------|-------|------------|
| `instance.compartment.id = 'ocid1...'` | All instances in one compartment | Production compartment with only application workloads |
| `tag.workload-identity.enabled.value = 'true'` | Tagged instances across tenancy | When only some instances need credentials |
| `All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1...'}` | Specific type in one compartment | Serverless workloads |
| `All {instance.compartment.id = 'ocid1...', tag.tier.value = 'web'}` | Tagged instances in one compartment | Multi-tier architectures with per-tier identity |

Broad matching rules (e.g., matching all instances across the tenancy) grant more resources the ability to use companion policy permissions than intended.

### Remember the Tenancy-Level Requirement

Dynamic groups must be created in the tenancy root compartment. The `compartmentId` field must always point to the tenancy OCID, not a child compartment:

```yaml
# Correct: tenancy OCID
compartmentId:
  value: "ocid1.tenancy.oc1..example"

# Wrong: child compartment OCID (will fail)
compartmentId:
  value: "ocid1.compartment.oc1..example"
```

The matching rule references child compartments (where the resources live), but the dynamic group itself lives in the tenancy.

### Tag for Organizational Tracking

Metadata labels are applied as OCI freeform tags. Use consistent labels across dynamic groups and their companion policies for auditing:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
```
