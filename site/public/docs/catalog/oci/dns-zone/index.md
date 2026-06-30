---
title: "DNS Zone"
description: "DNS Zone deployment documentation"
icon: "package"
order: 100
componentName: "ocidnszone"
---

# OCI DNS Zone

Deploys an Oracle Cloud Infrastructure DNS zone — a managed authoritative DNS zone supporting public (GLOBAL) and private resolution scopes, PRIMARY and SECONDARY zone types, zone transfers via external masters and downstreams, and DNSSEC signing.

## What Gets Created

When you deploy an OciDnsZone resource, Planton provisions:

- **DNS Zone** — a `dns.Zone` resource in the specified compartment. The zone name is derived from `metadata.name`. Supports GLOBAL (public) and PRIVATE scopes, PRIMARY and SECONDARY types, optional DNSSEC signing, and optional zone transfer configuration with external DNS servers.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the zone will be created — either a literal value or a reference to an OciCompartment resource
- **A DNS view OCID** (for private zones only) — required when scope is `private`
- **External master DNS server addresses** (for secondary zones only) — servers from which the zone will replicate
- **TSIG key OCIDs** (optional) — for authenticating zone transfers

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciDnsZone.example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
```

Deploy:

```shell
planton apply -f dns-zone.yaml
```

This creates a public PRIMARY DNS zone for `example.com`. The zone OCID and OCI-assigned nameservers are exported as stack outputs. Configure these nameservers as NS records at your domain registrar.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the zone will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `zoneType` | `enum` | Whether the zone is `primary` (authoritative source) or `secondary` (replicates from external masters). ForceNew. | Must be explicitly set |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `enum` | `global` | Resolution scope. `global` for publicly resolvable zones, `private` for VCN-only resolution. ForceNew. |
| `viewId` | `StringValueOrRef` | — | OCID of the private DNS view. Required when scope is `private`. ForceNew. |
| `isDnssecEnabled` | `bool` | OCI default (disabled) | Enable DNSSEC signing. When true, OCI generates KSK and ZSK key pairs and signs zone records. Only meaningful for GLOBAL zones. |
| `externalMasters` | `ExternalServer[]` | — | External master DNS servers for SECONDARY zones. Required when `zoneType` is `secondary`. |
| `externalDownstreams` | `ExternalServer[]` | — | External downstream DNS servers that receive zone transfers from PRIMARY zones. Only supported for PRIMARY zones with GLOBAL scope. |

### ExternalServer

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `address` | `string` | — | IPv4 or IPv6 address of the external DNS server. Required. |
| `port` | `int32` | `53` | Port number. |
| `tsigKeyId` | `string` | — | OCID of the TSIG key for authenticating zone transfers. |

### Validation Constraints

- `zoneType` must be explicitly set (`primary` or `secondary`)
- PRIVATE zones require `viewId`
- SECONDARY zones cannot have `private` scope (OCI limitation)
- SECONDARY zones require at least one `externalMasters` entry

## Examples

### Public Primary Zone

A standard public DNS zone:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsZone.example-com
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  zoneType: primary
```

### Public Primary Zone with DNSSEC

A DNSSEC-signed public zone for enhanced security:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: secure.example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsZone.secure-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  isDnssecEnabled: true
```

### Private Zone for VCN Resolution

A private DNS zone resolvable only within VCNs via a DNS view:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: internal.example.local
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsZone.internal-example-local
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  scope: scope_private
  viewId:
    value: "ocid1.dnsview.oc1..example"
```

### Secondary Zone Replicating from External Masters

A secondary zone that replicates from on-premises DNS servers with TSIG authentication:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: corp.example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsZone.corp-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: secondary
  externalMasters:
    - address: "198.51.100.1"
      tsigKeyId: "ocid1.tsigkey.oc1..example"
    - address: "198.51.100.2"
      tsigKeyId: "ocid1.tsigkey.oc1..example"
```

### Primary Zone with External Downstreams

A primary zone that pushes zone transfers to external DNS servers:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciDnsZone
metadata:
  name: distributed.example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciDnsZone.distributed-example-com
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  zoneType: primary
  externalDownstreams:
    - address: "203.0.113.1"
    - address: "203.0.113.2"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | OCID of the DNS zone |
| `nameservers` | `string` | Comma-separated list of OCI-assigned authoritative nameserver hostnames. Configure these as NS records at your domain registrar. |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciDnsRecord](/docs/catalog/oci/dns-record) — manages record sets within this zone
- [OciApplicationLoadBalancer](/docs/catalog/oci/application-load-balancer) — load balancer IPs are common DNS targets
- [OciVcn](/docs/catalog/oci/vcn) — VCNs with private DNS views for private zone resolution
