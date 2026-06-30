# OCI Compartment

Deploys an Oracle Cloud Infrastructure compartment for hierarchical resource isolation. Compartments are OCI's fundamental organizational primitive — every resource in OCI exists within exactly one compartment, and every other OCI component in Planton takes `compartmentId` as its first spec field. Nested hierarchies are built by chaining OciCompartment resources, where each child references its parent via `compartmentId`.

## What Gets Created

When you deploy an OciCompartment resource, Planton provisions:

- **Identity Compartment** — an `oci_identity_compartment` resource within the specified parent compartment or tenancy. The compartment is created with a name, description, and standard Planton freeform tags. By default, the compartment is retained even if the IaC resource is destroyed (`enableDelete` defaults to `false`).

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A parent compartment OCID or tenancy OCID** where the compartment will be created — the tenancy OCID for top-level compartments, or the OCID of an existing compartment (literal value or via `valueFrom` referencing another OciCompartment resource)

## Quick Start

Create a file `compartment.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciCompartment
metadata:
  name: my-compartment
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciCompartment.my-compartment
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Compartment for my-project workloads"
```

Deploy:

```shell
planton apply -f compartment.yaml
```

This creates a compartment named `my-compartment` under the specified tenancy. The compartment OCID is exported as a stack output for use as `compartmentId` in downstream resources such as OciVcn, OciSubnet, and OciSecurityGroup. Delete protection is enabled by default — destroying the IaC resource does not delete the compartment from OCI.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the parent compartment or tenancy where this compartment will be created. For top-level compartments, use the tenancy OCID. For nested compartments, use the parent compartment OCID. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `description` | `string` | Description of the compartment's purpose. Shown in the OCI Console and referenced by operators navigating the compartment hierarchy. | Minimum 1 character |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Name assigned to the compartment in OCI. Must be unique among siblings within the parent compartment. Shown in the OCI Console and used in IAM policy statements. Falls back to `metadata.name` if not provided. |
| `enableDelete` | `bool` | `false` | When `true`, the compartment is deleted when the IaC resource is destroyed. When `false`, the compartment is retained — OCI's safety mechanism to prevent accidental deletion of compartments containing active resources. Set to `true` only for ephemeral or development compartments. |

## Examples

### Minimal Project Compartment

A long-lived compartment for a project or team. Delete protection retains the compartment even if the IaC resource is destroyed:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciCompartment
metadata:
  name: platform-compartment
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciCompartment.platform-compartment
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Platform team infrastructure and shared services"
```

### Ephemeral Sandbox Compartment

A temporary compartment for development or CI/CD pipelines. Setting `enableDelete` to `true` allows full teardown:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciCompartment
metadata:
  name: ci-sandbox
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciCompartment.ci-sandbox
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  description: "Ephemeral sandbox for CI integration tests"
  enableDelete: true
```

### Nested Compartment with Foreign Key Reference

A child compartment referencing an Planton-managed parent via `valueFrom`, enabling declarative multi-level hierarchies:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciCompartment
metadata:
  name: networking
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciCompartment.networking
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: platform-compartment
      fieldPath: status.outputs.compartmentId
  description: "Networking resources for the platform team"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `compartment_id` | `string` | OCID of the created compartment. Referenced by virtually every other OCI component via `compartmentId.valueFrom`. |

## Related Components

- [OciVcn](/docs/catalog/oci/ocivcn) — creates virtual cloud networks within a compartment
- [OciIdentityPolicy](/docs/catalog/oci/ociidentitypolicy) — defines IAM policies scoped to a compartment
- [OciDynamicGroup](/docs/catalog/oci/ocidynamicgroup) — creates dynamic groups with compartment-scoped matching rules
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — manages network security rules within a compartment
- [OciSubnet](/docs/catalog/oci/ocisubnet) — creates subnets within a VCN that lives in a compartment
