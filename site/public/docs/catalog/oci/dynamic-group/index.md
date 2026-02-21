---
title: "Dynamic Group"
description: "Dynamic Group deployment documentation"
icon: "package"
order: 100
componentName: "ocidynamicgroup"
---

# OCI Dynamic Group

Deploys an Oracle Cloud Infrastructure dynamic group for enabling workload identity — the mechanism that lets compute instances, OKE pods, and Functions authenticate to OCI services without stored credentials. A dynamic group uses a matching rule to select which OCI resources are members. Combined with an `OciIdentityPolicy`, dynamic groups enable the credential-less authentication pattern: the dynamic group defines *who* (matching rule), and the policy defines *what they can do* (statements).

## What Gets Created

When you deploy an OciDynamicGroup resource, OpenMCF provisions:

- **Identity Dynamic Group** — an `oci_identity_dynamic_group` resource in the tenancy with the provided name, description, and matching rule. Standard OpenMCF freeform tags are applied automatically. The dynamic group name defaults to `metadata.name` if not explicitly set in the spec.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **The tenancy OCID** — dynamic groups are tenancy-level IAM resources and must be created in the tenancy root compartment, not in a child compartment (literal value or via `valueFrom` referencing an OciCompartment resource)
- **Knowledge of matching rule syntax** — rules follow OCI's `Any {condition}` or `All {condition, condition}` syntax. See [Managing Dynamic Groups](https://docs.oracle.com/iaas/Content/Identity/Tasks/managingdynamicgroups.htm)

## Quick Start

Create a file `dynamic-group.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: my-compute-workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciDynamicGroup.my-compute-workers
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "Dynamic group matching all compute instances in the project compartment"
  matchingRule: "Any {instance.compartment.id = 'ocid1.compartment.oc1..example'}"
```

Deploy:

```shell
openmcf apply -f dynamic-group.yaml
```

This creates a dynamic group named `my-compute-workers` in the tenancy. All compute instances in the specified compartment automatically become members. To grant these instances permissions, create a companion `OciIdentityPolicy` with statements like `Allow dynamic-group my-compute-workers to read secret-family in compartment production`. The dynamic group OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the tenancy (root compartment). Dynamic groups are tenancy-level IAM resources and must be created in the tenancy compartment. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `description` | `string` | Description of the dynamic group's purpose. Required by the OCI API. Updatable after creation. | Minimum 1 character |
| `matchingRule` | `string` | Rule that defines which OCI resources belong to this dynamic group. Uses `Any {conditions}` (match any condition) or `All {conditions}` (match all conditions) syntax. | Minimum 1 character |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Name assigned to the dynamic group. Must be unique across all groups (including user groups) in the tenancy. Cannot be changed after creation. Falls back to `metadata.name` if not provided. |

## Examples

### Compute Instance Principal

A dynamic group matching all compute instances in a compartment — the most common pattern for enabling instance principal authentication:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: compute-workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDynamicGroup.compute-workers
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "All compute instances in the production compartment"
  matchingRule: "Any {instance.compartment.id = 'ocid1.compartment.oc1..production'}"
```

### Functions Workload Identity

A dynamic group matching all OCI Functions in a compartment for serverless workload identity. Uses `All` to require both the resource type and compartment conditions:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: serverless-functions
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDynamicGroup.serverless-functions
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "All OCI Functions in the production compartment for serverless workload identity"
  matchingRule: "All {resource.type = 'fnfunc', resource.compartment.id = 'ocid1.compartment.oc1..production'}"
```

### Tag-Based Matching

A dynamic group matching resources by freeform tag rather than compartment. This pattern allows fine-grained grouping that spans compartments or selects a subset of resources within a compartment:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: tagged-workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciDynamicGroup.tagged-workers
spec:
  compartmentId:
    value: "ocid1.tenancy.oc1..example"
  description: "All compute instances tagged with workload-identity=enabled"
  matchingRule: "Any {tag.workload-identity.enabled.value = 'true'}"
```

### Full-Featured with Foreign Key References

A dynamic group using `valueFrom` to reference the tenancy compartment, with a custom name and metadata labels for organizational tracking:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciDynamicGroup
metadata:
  name: oke-worker-nodes
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: platform-infra
    pulumi.openmcf.org/stack.name: prod.OciDynamicGroup.oke-worker-nodes
    team: platform
    cost-center: infrastructure
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: tenancy-root
      fieldPath: status.outputs.compartmentId
  name: "prod-oke-worker-dynamic-group"
  description: "OKE worker nodes in the production compartment for instance principal authentication"
  matchingRule: "All {instance.compartment.id = 'ocid1.compartment.oc1..production', tag.oke-cluster.name.value = 'prod-cluster'}"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `dynamic_group_id` | `string` | OCID of the created dynamic group. |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the tenancy OCID referenced by `compartmentId`; matching rules typically reference compartment OCIDs to scope group membership
- [OciIdentityPolicy](/docs/catalog/oci/identity-policy) — grants dynamic group members permissions via policy statements (`Allow dynamic-group ... to ...`); every dynamic group needs a companion policy to be useful
- [OciComputeInstance](/docs/catalog/oci/compute-instance) — compute instances matched by `instance.compartment.id` rules become dynamic group members for instance principal authentication
- [OciFunctionsApplication](/docs/catalog/oci/functions-application) — functions matched by `resource.type = 'fnfunc'` rules become dynamic group members for serverless workload identity
