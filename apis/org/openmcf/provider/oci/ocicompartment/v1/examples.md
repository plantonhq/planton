# OCI Compartment Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure compartments using the OpenMCF API. Each example demonstrates different use cases from minimal project isolation to multi-level organizational hierarchies.

## Table of Contents

- [Example 1: Minimal Project Compartment](#example-1-minimal-project-compartment)
- [Example 2: Sandbox Compartment](#example-2-sandbox-compartment)
- [Example 3: Nested Compartment Hierarchy](#example-3-nested-compartment-hierarchy)
- [Example 4: Multi-Team Organization Structure](#example-4-multi-team-organization-structure)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Project Compartment

**Use Case:** Long-lived compartment for a project or team. Delete protection keeps the compartment even if the IaC resource is destroyed.

**Configuration:**
- **Parent:** Tenancy root
- **Delete Protection:** On (default)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: my-project
  org: my-org
  env: prod
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Production infrastructure for my-project"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f my-project-compartment.yaml
```

**What happens:**
- A compartment named `my-project` is created under the tenancy root.
- Standard OpenMCF freeform tags are applied (resource kind, resource ID, organization, environment).
- Delete protection is enabled — destroying the IaC resource leaves the compartment intact.
- The compartment OCID is exported as `compartment_id` for use in downstream resources.

---

## Example 2: Sandbox Compartment

**Use Case:** Ephemeral compartment for development, CI/CD pipelines, or proof-of-concept work. The compartment is destroyed alongside the IaC resource.

**Configuration:**
- **Parent:** Development compartment
- **Delete Protection:** Off (`enableDelete: true`)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: ci-sandbox-42
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..devparent"
  description: "Ephemeral sandbox for CI pipeline run #42"
  enableDelete: true
```

**Deploy:**

```bash
openmcf apply -f ci-sandbox.yaml
```

**What happens:**
- A compartment named `ci-sandbox-42` is created under the development parent compartment.
- `enableDelete: true` means the compartment will be deleted when the IaC resource is destroyed.
- OCI will refuse to delete the compartment if it still contains active resources — ensure all child resources are destroyed first.

---

## Example 3: Nested Compartment Hierarchy

**Use Case:** A child compartment that references an OpenMCF-managed parent, creating a declarative two-level hierarchy without hardcoded OCIDs.

**Configuration:**
- **Parent:** Referenced via `valueFrom` on another OciCompartment
- **Delete Protection:** On

**Parent compartment:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: platform
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Top-level compartment for platform team"
```

**Child compartment:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: networking
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: platform
      fieldPath: status.outputs.compartmentId
  description: "Networking resources owned by the platform team"
```

**Deploy in order:**

```bash
openmcf apply -f platform-compartment.yaml
openmcf apply -f networking-compartment.yaml
```

**What happens:**
- The `platform` compartment is created under the tenancy root.
- The `networking` compartment is created inside `platform`, using the parent's `compartment_id` output via `valueFrom`.
- Both compartments have delete protection enabled.
- The `networking` compartment's `compartment_id` output can be referenced by OciVcn, OciSubnet, and other networking resources.

---

## Example 4: Multi-Team Organization Structure

**Use Case:** An organizational compartment hierarchy with dedicated compartments for teams, each containing environment-specific sub-compartments.

**Configuration:**
- **Structure:** tenancy > engineering > team-alpha > prod
- **Three levels of nesting**

**Level 1 — Engineering department:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: engineering
  org: acme-corp
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Engineering department — all engineering teams and their workloads"
```

**Level 2 — Team compartment:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: team-alpha
  org: acme-corp
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: engineering
      fieldPath: status.outputs.compartmentId
  description: "Team Alpha — backend services and databases"
```

**Level 3 — Environment compartment:**

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciCompartment
metadata:
  name: team-alpha-prod
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: team-alpha
      fieldPath: status.outputs.compartmentId
  description: "Production environment for Team Alpha"
```

**What happens:**
- Three compartments are created in a hierarchy: `engineering` > `team-alpha` > `team-alpha-prod`.
- IAM policies can be scoped at any level: a policy on `engineering` applies to all teams; a policy on `team-alpha` applies only to that team; a policy on `team-alpha-prod` applies only to production.
- Cost tracking is automatic at each level — the OCI cost analysis dashboard shows spend per compartment.

---

## Common Operations

### Get Compartment OCID After Deployment

```bash
# Pulumi
pulumi stack output compartment_id

# Terraform
terraform output compartment_id
```

### Use Compartment in a Downstream OciVcn

The `compartment_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in any downstream OCI resource:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciVcn
metadata:
  name: prod-vcn
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: networking
      fieldPath: status.outputs.compartmentId
  cidrBlocks:
    - "10.0.0.0/16"
```

### Destroy a Compartment with Delete Protection

By default, `enableDelete` is `false`. Destroying the IaC resource leaves the compartment in OCI:

```bash
# This removes the resource from IaC state but does NOT delete the compartment
openmcf destroy -f compartment.yaml
```

To actually delete the compartment, first enable deletion, then destroy:

```yaml
# Step 1: Set enableDelete to true in your manifest
spec:
  enableDelete: true
```

```bash
# Step 2: Apply the change
openmcf apply -f compartment.yaml

# Step 3: Destroy
openmcf destroy -f compartment.yaml
```

OCI will refuse to delete a compartment that still contains active resources. Ensure all child resources and sub-compartments are removed first.

---

## Best Practices

### Design Your Hierarchy Before Creating Compartments

Plan the compartment tree before creating resources. Common patterns:

| Pattern | Structure | Best For |
|---------|-----------|----------|
| **Flat** | tenancy > project-a, project-b, ... | Small teams with few projects |
| **Team-Based** | tenancy > team > project | Organizations with clear team ownership |
| **Environment-Based** | tenancy > env > project | When environment isolation is the priority |
| **Hybrid** | tenancy > team > env | Enterprise organizations needing both team and environment boundaries |

**Recommendation:** Start with the hybrid pattern (team > environment) for organizations with more than one team. It provides both ownership boundaries and environment isolation.

### Use Descriptions Consistently

OCI requires a description on every compartment. Use this to document:
- What the compartment is for
- Which team owns it
- Whether it is permanent or ephemeral

Good descriptions help operators navigating the OCI Console understand the hierarchy without reading IaC source code.

### Default to Delete Protection

Leave `enableDelete` as `false` (the default) for any compartment that contains or will contain production resources. Reserve `enableDelete: true` for:
- CI/CD pipeline sandboxes
- Developer personal sandboxes
- Short-lived demo or training environments

The two-step deletion process (enable, apply, destroy) is intentional friction against accidental deletion.

### Keep Nesting Shallow

OCI supports up to six levels of compartment nesting. In practice, three to four levels is sufficient for most organizations:

```
tenancy
└── business-unit
    └── team
        └── environment
```

Deep nesting makes IAM policies harder to reason about and the OCI Console harder to navigate.

### Tag for Cost and Compliance

Metadata labels are applied as OCI freeform tags. Use consistent labels across all compartments:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
```
