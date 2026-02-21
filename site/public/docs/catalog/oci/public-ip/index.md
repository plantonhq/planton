---
title: "Public IP"
description: "Public IP deployment documentation"
icon: "package"
order: 100
componentName: "ocipublicip"
---

# OCI Public IP

Deploys an Oracle Cloud Infrastructure public IPv4 address for internet connectivity. The component supports both reserved (persistent, region-scoped) and ephemeral (lifecycle-tied) lifetime modes, with optional assignment to a private IP and allocation from a BYOIP pool.

## What Gets Created

When you deploy an OciPublicIp resource, OpenMCF provisions:

- **Public IP** — an `oci_core_public_ip` resource in the specified compartment. The lifetime mode (`RESERVED` or `EPHEMERAL`) determines whether the IP persists independently or is tied to the assigned entity. OCI freeform tags are applied automatically with the resource kind, resource ID, organization, and environment metadata.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the public IP will be created — either a literal value or a reference to an OciCompartment resource
- **For ephemeral IPs**: a private IP OCID — the `privateIpId` field is required when `lifetime` is `EPHEMERAL`

## Quick Start

Create a file `public-ip.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: my-public-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPublicIp.my-public-ip
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
```

Deploy:

```shell
openmcf apply -f public-ip.yaml
```

This creates a reserved public IP that is unassigned. The allocated IP address and OCID are exported as stack outputs. You can assign the IP to a private IP later via the OCI Console or API, or by updating the manifest with `privateIpId`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the public IP will reside. For ephemeral IPs, must match the compartment of the private IP. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `lifetime` | `string` | Lifetime mode for the public IP. `RESERVED` creates a persistent, region-scoped IP. `EPHEMERAL` creates an IP tied to the assigned entity. Cannot be changed after creation. | Must be `"RESERVED"` or `"EPHEMERAL"` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the public IP shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `privateIpId` | `StringValueOrRef` | — | OCID of the private IP to assign this public IP to. Required for ephemeral IPs (must be a primary private IP on a VNIC). Optional for reserved IPs — when omitted the IP is created unassigned and can be attached later. |
| `publicIpPoolId` | `StringValueOrRef` | — | OCID of a public IP pool for BYOIP (Bring Your Own IP) scenarios. When set, the public IP is allocated from the specified pool instead of Oracle's pool. Cannot be changed after creation. |

## Examples

### Reserved Unassigned IP

A reserved public IP with no private IP assignment — suitable for pre-allocating a stable address for DNS records or firewall allowlists:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: reserved-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciPublicIp.reserved-ip
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
```

### Reserved IP Assigned to a Private IP

A reserved public IP assigned to an existing private IP, giving a compute instance a stable internet-facing address that survives instance termination:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: web-server-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciPublicIp.web-server-ip
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
  displayName: "Web Server Public IP"
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

### Reserved IP from a BYOIP Pool

A reserved public IP drawn from a BYOIP pool, for organizations that own their own IP address ranges:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: byoip-address
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPublicIp.byoip-address
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  lifetime: RESERVED
  displayName: "BYOIP Address"
  publicIpPoolId:
    value: "ocid1.publicippool.oc1.iad.example"
```

### Using Foreign Key References

Reference an OpenMCF-managed compartment instead of hardcoding the OCID:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciPublicIp
metadata:
  name: ref-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciPublicIp.ref-ip
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  lifetime: RESERVED
  privateIpId:
    value: "ocid1.privateip.oc1.iad.example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `publicIpId` | `string` | OCID of the created public IP resource |
| `ipAddress` | `string` | The allocated IPv4 address (e.g. `203.0.113.2`) |

## Related Components

- [OciVcn](/docs/catalog/oci/vcn) — creates the virtual cloud network whose subnets host the private IPs that public IPs can be assigned to
- [OciSubnet](/docs/catalog/oci/subnet) — creates subnets within a VCN where instances with public IPs reside
- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
