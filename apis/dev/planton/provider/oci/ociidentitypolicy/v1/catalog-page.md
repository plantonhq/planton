# OCI Identity Policy

Deploys an Oracle Cloud Infrastructure IAM policy for granting access to compartment resources. Policies are OCI's authorization mechanism — each policy contains one or more human-readable statements written in OCI's policy language (e.g., `Allow group Admins to manage all-resources in compartment Production`). Policies are attached to a compartment and grant permissions within that compartment and all of its children.

## What Gets Created

When you deploy an OciIdentityPolicy resource, Planton provisions:

- **Identity Policy** — an `oci_identity_policy` resource in the specified compartment with the provided name, description, and policy statements. Standard Planton freeform tags are applied automatically. The policy name defaults to `metadata.name` if not explicitly set in the spec.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID or tenancy OCID** where the policy will be created — the tenancy OCID for tenancy-level policies, or a compartment OCID for compartment-scoped policies (literal value or via `valueFrom` referencing an OciCompartment resource)
- **Knowledge of OCI policy syntax** — statements follow the format `Allow <subject> to <verb> <resource-type> in <location> [where <conditions>]`. See [OCI Policy Reference](https://docs.oracle.com/iaas/Content/Identity/Reference/policyreference.htm)

## Quick Start

Create a file `policy.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciIdentityPolicy
metadata:
  name: my-admin-policy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciIdentityPolicy.my-admin-policy
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Grants admin access to the project compartment"
  statements:
    - "Allow group ProjectAdmins to manage all-resources in compartment my-project"
```

Deploy:

```shell
planton apply -f policy.yaml
```

This creates an IAM policy named `my-admin-policy` attached to the specified compartment. The single statement grants the `ProjectAdmins` group full administrative access to all resources within the `my-project` compartment and its children. The policy OCID is exported as a stack output for reference.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where this policy will be created. For tenancy-level policies, use the tenancy OCID. For compartment-scoped policies, use the target compartment's OCID. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `description` | `string` | Description of the policy's purpose. Required by the OCI API. Updatable after creation. | Minimum 1 character |
| `statements` | `string[]` | Policy statements written in OCI's policy language. Each statement follows the syntax: `Allow <subject> to <verb> <resource-type> in <location> [where <conditions>]`. | Minimum 1 statement |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Name assigned to the policy. Must be unique across all policies in the tenancy. Cannot be changed after creation. Falls back to `metadata.name` if not provided. |
| `versionDate` | `string` | _(empty)_ | Version date for policy evaluation in `YYYY-MM-DD` format. When set, the policy is evaluated according to the behavior of OCI services on that date, providing stable policy interpretation. When empty, the policy uses current service behavior at evaluation time. |

## Examples

### Compartment Admin Policy

A policy granting a group full administrative access to all resources within a compartment — the most common OCI policy pattern:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciIdentityPolicy
metadata:
  name: compartment-admin-policy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciIdentityPolicy.compartment-admin-policy
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Grants administrative access to the production compartment"
  statements:
    - "Allow group PlatformAdmins to manage all-resources in compartment production"
```

### Dynamic Group Service Access

A policy granting a dynamic group access to specific OCI services — the standard workload identity pattern for compute instances and OKE pods that need to call OCI APIs without stored credentials:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciIdentityPolicy
metadata:
  name: workload-service-access
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciIdentityPolicy.workload-service-access
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Grants compute workloads access to Vault secrets, KMS keys, and Object Storage"
  statements:
    - "Allow dynamic-group compute-workers to read secret-family in compartment production"
    - "Allow dynamic-group compute-workers to use keys in compartment production"
    - "Allow dynamic-group compute-workers to manage object-family in compartment production"
```

### Tenancy-Wide Auditor

A tenancy-level policy granting read-only visibility across all compartments for compliance and security auditing:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciIdentityPolicy
metadata:
  name: auditor-policy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciIdentityPolicy.auditor-policy
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Grants read-only visibility across the tenancy for security auditing"
  statements:
    - "Allow group SecurityAuditors to inspect all-resources in tenancy"
    - "Allow group SecurityAuditors to read audit-events in tenancy"
```

### Full-Featured with Foreign Key References

A policy using `valueFrom` to reference an Planton-managed compartment, with a version date for stable policy evaluation and custom metadata labels:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciIdentityPolicy
metadata:
  name: network-admin-policy
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme-corp
    pulumi.planton.dev/project: platform-infra
    pulumi.planton.dev/stack.name: prod.OciIdentityPolicy.network-admin-policy
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
  versionDate: "2026-01-01"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `policy_id` | `string` | OCID of the created policy. |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`; policies are scoped to the compartment they are attached to
- [OciDynamicGroup](/docs/catalog/oci/ocidynamicgroup) — creates dynamic groups that can be referenced as subjects in policy statements (`Allow dynamic-group ...`)
- [OciVcn](/docs/catalog/oci/ocivcn) — policies using `virtual-network-family` govern access to VCN resources within a compartment
- [OciSubnet](/docs/catalog/oci/ocisubnet) — policies govern access to subnets and related networking resources within compartments
