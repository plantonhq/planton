# Civo IP Address

Deploys a static reserved (public) IPv4 address on Civo Cloud. Reserved IPs persist independently of instances and load balancers, making them useful for stable external endpoints that survive resource replacements.

## What Gets Created

When you deploy a CivoIpAddress resource, OpenMCF provisions:

- **Reserved IP** — a `civo_reserved_ip` resource that allocates a persistent public IPv4 address in the specified Civo region

The IP is created in an unattached state. You can later associate it with a CivoComputeInstance or load balancer in the same region.

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **A target Civo region** — reserved IPs are region-scoped and can only be attached to resources in the same region

## Quick Start

Create a file `civo-ip.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoIpAddress
metadata:
  name: my-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoIpAddress.my-ip
spec:
  region: nyc1
```

Deploy:

```shell
openmcf apply -f civo-ip.yaml
```

This allocates a reserved IPv4 address in the New York region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `enum` | Civo region where the IP is allocated. The IP can only be attached to resources in this region. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable name or description for the reserved IP. If omitted, Civo may default to using the IP address itself as the label. Max 100 characters. |

## Examples

### Basic Reserved IP

A minimal manifest that allocates a reserved IP in Frankfurt:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoIpAddress
metadata:
  name: basic-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoIpAddress.basic-ip
spec:
  region: fra1
```

### Reserved IP with Description

Adding a description makes the IP easier to identify in the Civo dashboard and in IaC state:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoIpAddress
metadata:
  name: api-gateway-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.CivoIpAddress.api-gateway-ip
spec:
  region: lon1
  description: "API gateway public endpoint"
```

### Production Stable Endpoint

A reserved IP intended for a production load balancer, paired with a DNS record:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoIpAddress
metadata:
  name: prod-lb-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoIpAddress.prod-lb-ip
spec:
  region: nyc1
  description: "Production load balancer IP"
```

After deployment, use the `ipAddress` output to configure a CivoDnsRecord that points your domain to this stable IP.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `reservedIpId` | `string` | Unique identifier (UUID) of the reserved IP in Civo |
| `ipAddress` | `string` | The static IPv4 address allocated for this reserved IP |
| `attachedResourceId` | `string` | ID of the Civo resource (instance or load balancer) currently attached to this IP. Empty if unattached. |
| `createdAtRfc3339` | `string` | Timestamp when the reserved IP was created, in RFC 3339 format |

## Related Components

- [CivoComputeInstance](/docs/catalog/civo/civocomputeinstance) — attach the reserved IP to a compute instance for a stable public address
- [CivoDnsRecord](/docs/catalog/civo/civodnsrecord) — create DNS records pointing to the reserved IP
- [CivoFirewall](/docs/catalog/civo/civofirewall) — control inbound traffic to resources using this IP
- [CivoVpc](/docs/catalog/civo/civovpc) — private network for the resources that use this IP
