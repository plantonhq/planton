# OCI Identity Policy Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure IAM policies using the OpenMCF API. Each example demonstrates different authorization patterns from single-statement admin grants to multi-statement least-privilege configurations with dynamic groups.

## Table of Contents

- [Example 1: Compartment Admin Policy](#example-1-compartment-admin-policy)
- [Example 2: Dynamic Group Service Access](#example-2-dynamic-group-service-access)
- [Example 3: Tenancy-Wide Auditor](#example-3-tenancy-wide-auditor)
- [Example 4: Full-Featured Network Admin with Foreign Key References](#example-4-full-featured-network-admin-with-foreign-key-references)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Compartment Admin Policy

**Use Case:** Grant a group full administrative access to all resources within a compartment. This is the most common OCI policy pattern — the first thing every team creates after receiving a new compartment.

**Configuration:**
- **Scope:** Single compartment
- **Subject:** IAM group
- **Permission:** `manage all-resources` (broadest grant)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciIdentityPolicy
metadata:
  name: platform-admin-policy
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Grants the platform team full administrative access to the production compartment"
  statements:
    - "Allow group PlatformAdmins to manage all-resources in compartment production"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f platform-admin-policy.yaml
```

**What happens:**
- A policy named `platform-admin-policy` is created in the specified compartment.
- The single statement grants `PlatformAdmins` the `manage` verb on `all-resources`, which includes `inspect`, `read`, `use`, and all create/update/delete operations on every resource type.
- The policy applies within the target compartment and all of its child compartments.
- Standard OpenMCF freeform tags are applied automatically.
- The policy OCID is exported as `policy_id`.

---

## Example 2: Dynamic Group Service Access

**Use Case:** Grant a dynamic group access to specific OCI services for workload identity. This is the standard pattern for compute instances, OKE pods, and Functions that need to call OCI APIs without stored credentials. Pair this with an `OciDynamicGroup` resource that matches your workload instances.

**Configuration:**
- **Scope:** Single compartment
- **Subject:** Dynamic group (workload identity)
- **Permission:** Least-privilege per service (`read`, `use`, `manage` as appropriate)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciIdentityPolicy
metadata:
  name: workload-service-access
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Grants compute workloads access to Vault secrets, KMS keys, Object Storage, and networking"
  statements:
    - "Allow dynamic-group compute-workers to read secret-family in compartment production"
    - "Allow dynamic-group compute-workers to use keys in compartment production"
    - "Allow dynamic-group compute-workers to manage object-family in compartment production"
    - "Allow dynamic-group compute-workers to use virtual-network-family in compartment production"
```

**Deploy:**

```bash
openmcf apply -f workload-service-access.yaml
```

**What happens:**
- Four statements are created, each granting the `compute-workers` dynamic group different permission levels on different resource families.
- `read secret-family` — retrieve secret values from OCI Vault (read access, not create/delete).
- `use keys` — encrypt and decrypt with KMS keys (use access, not create/delete keys).
- `manage object-family` — full CRUD on Object Storage buckets and objects.
- `use virtual-network-family` — attach to VNICs and subnets (use access, not create/delete networks).
- The dynamic group `compute-workers` must exist separately as an `OciDynamicGroup` resource.

---

## Example 3: Tenancy-Wide Auditor

**Use Case:** Grant a group read-only visibility across all compartments for compliance, security, or cost auditing. Attached to the tenancy root so it applies everywhere.

**Configuration:**
- **Scope:** Tenancy-wide (attached to tenancy root)
- **Subject:** IAM group
- **Permission:** `inspect all-resources` + `read audit-events`

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciIdentityPolicy
metadata:
  name: security-auditor-policy
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Grants security auditors read-only visibility across the entire tenancy"
  statements:
    - "Allow group SecurityAuditors to inspect all-resources in tenancy"
    - "Allow group SecurityAuditors to read audit-events in tenancy"
```

**What happens:**
- The policy is attached to the tenancy root, making it apply across all compartments.
- `inspect all-resources` is the lowest permission level — it allows listing resources and viewing metadata (names, OCIDs, compartments, tags, lifecycle state) but not reading data contents. An auditor can see that an Object Storage bucket exists but cannot read its objects.
- `read audit-events` supplements `inspect` with access to OCI Audit service logs, which record all API calls in the tenancy. `read` is required instead of `inspect` because audit event contents need to be viewed, not just listed.
- The `compartmentId` uses a tenancy OCID (`ocid1.tenancy.oc1..`) rather than a compartment OCID.

---

## Example 4: Full-Featured Network Admin with Foreign Key References

**Use Case:** A production policy using `valueFrom` to reference an OpenMCF-managed compartment, with a custom name, version date for stable evaluation, and metadata labels for organizational tracking.

**Configuration:**
- **Scope:** Single compartment (via foreign key reference)
- **Subject:** IAM group
- **Permission:** Network-specific resource families
- **Version Date:** Pinned for stable evaluation

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciIdentityPolicy
metadata:
  name: network-admin-policy
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: networking
      fieldPath: status.outputs.compartmentId
  name: "platform-network-admin-policy"
  description: "Grants the networking team full control over virtual network resources"
  statements:
    - "Allow group NetworkAdmins to manage virtual-network-family in compartment networking"
    - "Allow group NetworkAdmins to manage load-balancers in compartment networking"
    - "Allow group NetworkAdmins to use network-security-groups in compartment networking"
    - "Allow group NetworkAdmins to read all-resources in compartment networking"
  versionDate: "2026-01-01"
```

**What happens:**
- The `compartmentId` is resolved from a previously deployed OciCompartment resource named `networking` instead of a hardcoded OCID.
- The explicit `name` field sets the OCI policy name to `platform-network-admin-policy` instead of the default `metadata.name` value of `network-admin-policy`.
- `versionDate: "2026-01-01"` pins policy evaluation to the service behavior on that date. If OCI later changes how `virtual-network-family` or `load-balancers` resource types behave, this policy continues evaluating as it did on January 1, 2026.
- Metadata labels (`team: platform`, `cost-center: infrastructure`) are applied as OCI freeform tags alongside the standard OpenMCF tags.
- Four statements grant a progressive permission model: `manage` on networking and LB resources, `use` on NSGs (attach/detach but not create/delete), and `read` on everything else in the compartment for visibility.

---

## Common Operations

### Get Policy OCID After Deployment

```bash
# Pulumi
pulumi stack output policy_id

# Terraform
terraform output policy_id
```

### Verify Policy Exists in OCI Console

Navigate to **Identity & Security > Identity > Policies** in the OCI Console. Filter by compartment to locate your policy. The policy name, description, and all statements are visible in the console.

### Update Policy Statements

Policy statements and description are updatable after creation. Modify the `statements` list in your manifest and re-apply:

```bash
openmcf apply -f policy.yaml
```

The IaC engine will update the existing policy in-place. The policy OCID remains the same.

### Destroy a Policy

```bash
openmcf destroy -f policy.yaml
```

Destroying a policy immediately revokes the permissions it grants. Any group or dynamic group that relied on this policy loses access. Ensure that workloads are not depending on the policy before destruction.

---

## Best Practices

### One Concern Per Policy

Keep policies focused on a single authorization concern. Instead of one policy with 20 statements covering networking, compute, storage, and IAM, create separate policies:

```
network-admin-policy   → virtual-network-family, load-balancers, NSGs
compute-admin-policy   → instances, instance-pools, autoscaling
storage-admin-policy   → object-family, file-systems, block-volumes
```

Separate policies are independently auditable, independently revocable, and independently manageable by different teams.

### Use Resource Type Families Over all-resources

The `manage all-resources` grant is appropriate for compartment administrators but too broad for most roles. Prefer specific resource type families:

| Resource Family | Covers |
|----------------|--------|
| `virtual-network-family` | VCNs, subnets, route tables, security lists, NSGs, gateways |
| `instance-family` | Compute instances, instance configurations, instance pools |
| `object-family` | Object Storage buckets, objects, lifecycle policies |
| `database-family` | DB systems, Autonomous Databases |
| `secret-family` | Vault secrets |

This follows the principle of least privilege — grant only what the subject needs.

### Use the Correct Verb

OCI's four verbs form a hierarchy. Each higher verb includes all permissions of the lower verbs:

| Verb | Permissions |
|------|-------------|
| `inspect` | List resources, view metadata |
| `read` | `inspect` + read resource data contents |
| `use` | `read` + use existing resources (attach, launch into) |
| `manage` | `use` + create, update, delete |

Choose the minimum verb. Auditors need `inspect`. Workloads reading secrets need `read`. Compute instances attaching to subnets need `use`. Administrators need `manage`.

### Scope Policies to the Narrowest Compartment

Attach policies to the most specific compartment that covers the needed scope. A policy on the tenancy root grants access everywhere — this is correct for auditors but overly broad for team-level access. Attach team policies to the team's compartment.

### Pin versionDate for Stable Environments

For production policies where predictable evaluation is critical, set `versionDate` to a known-good date. This prevents OCI service updates from unexpectedly changing how your policy is evaluated. Update the version date explicitly and deliberately when you want to adopt new service behavior.

### Document Policies with Descriptions

The `description` field is required by OCI and appears in the console. Write meaningful descriptions that explain the policy's purpose, not just restate the statements:

- Good: `"Grants the platform team access to manage networking infrastructure in the production compartment"`
- Bad: `"Policy for networking"`

### Tag for Organizational Tracking

Metadata labels are applied as OCI freeform tags. Use consistent labels across all policies for auditing and compliance:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
    compliance: sox
```
