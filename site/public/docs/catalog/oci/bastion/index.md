---
title: "Bastion"
description: "Bastion deployment documentation"
icon: "package"
order: 100
componentName: "ocibastion"
---

# OCI Bastion

Deploys an Oracle Cloud Infrastructure Bastion — a managed SSH gateway that provides secure, time-limited access to resources in private subnets without requiring a public IP on the target. Supports managed SSH sessions, port forwarding, and optional DNS proxy (FQDN and SOCKS5) for FQDN-based target resolution.

## What Gets Created

When you deploy an OciBastion resource, Planton provisions:

- **Bastion** — a `bastion.Bastion` resource (type STANDARD) in the specified compartment with a private endpoint in the target subnet. The bastion controls which client CIDR ranges can establish sessions and enforces a maximum session TTL.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the bastion will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** — the private subnet that the bastion connects to, either as a literal value or via `valueFrom` referencing an OciSubnet resource

## Quick Start

Create a file `bastion.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciBastion
metadata:
  name: my-bastion
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciBastion.my-bastion
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  targetSubnetId:
    value: "ocid1.subnet.oc1..example"
```

Deploy:

```shell
planton apply -f bastion.yaml
```

This creates a bastion with a private endpoint in the target subnet, no CIDR restrictions, and the OCI default maximum session TTL (3 hours). The bastion OCID and private endpoint IP are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the bastion will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `targetSubnetId` | `StringValueOrRef` | OCID of the subnet where the bastion creates its private endpoint. Immutable after creation. Can reference an OciSubnet resource via `valueFrom` using `status.outputs.subnetId`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the bastion. Immutable after creation. |
| `clientCidrBlockAllowList` | `string[]` | — | CIDR ranges allowed to connect to sessions (e.g., `["10.0.0.0/16"]`). When empty, all client IPs are allowed. Updatable. |
| `maxSessionTtlInSeconds` | `int32` | `10800` (3 hours) | Maximum TTL in seconds for any session on this bastion. Updatable. |
| `isDnsProxyEnabled` | `bool` | `false` | Enable FQDN and SOCKS5 proxy support. When `true`, sessions can use DNS names to reach targets. Immutable after creation. |

## Examples

### Minimal Bastion

A bastion with default settings — no CIDR restrictions, 3-hour maximum session TTL:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciBastion
metadata:
  name: dev-bastion
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciBastion.dev-bastion
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  targetSubnetId:
    value: "ocid1.subnet.oc1..example"
```

### CIDR-Restricted with Extended Session TTL

A bastion that only allows connections from a corporate network, with an 8-hour maximum session TTL:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciBastion
metadata:
  name: corp-bastion
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciBastion.corp-bastion
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  targetSubnetId:
    valueFrom:
      kind: OciSubnet
      name: private-subnet
      fieldPath: status.outputs.subnetId
  displayName: "corp-bastion-prod"
  clientCidrBlockAllowList:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
  maxSessionTtlInSeconds: 28800
```

### DNS Proxy Enabled

A bastion with DNS proxy support for FQDN-based target resolution and SOCKS5 dynamic port forwarding:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciBastion
metadata:
  name: dns-bastion
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciBastion.dns-bastion
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  targetSubnetId:
    value: "ocid1.subnet.oc1..example"
  clientCidrBlockAllowList:
    - "10.0.0.0/16"
  maxSessionTtlInSeconds: 14400
  isDnsProxyEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bastion_id` | `string` | OCID of the bastion |
| `private_endpoint_ip_address` | `string` | Private IP address of the bastion's endpoint in the target subnet |

## Related Components

- [OciSubnet](/docs/catalog/oci/subnet) — provides the target subnet referenced by `targetSubnetId` via `valueFrom`
- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciComputeInstance](/docs/catalog/oci/compute-instance) — common target for bastion sessions in private subnets
