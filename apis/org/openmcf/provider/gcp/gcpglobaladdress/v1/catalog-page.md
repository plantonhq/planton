# GCP Global Address

Reserves a static IP address at global scope — either a public IPv4/IPv6 address for HTTP(S) load balancers and Cloud CDN, or a private CIDR range inside a VPC for managed-service peering (Cloud SQL, Redis, AlloyDB, Filestore) and Private Service Connect endpoints. The component automatically enables the Compute Engine API on the target project.

## What Gets Created

When you deploy a GcpGlobalAddress resource, OpenMCF provisions:

- **Compute Engine API enablement** — a `google_project_service` resource that activates `compute.googleapis.com` on the target project
- **Global Address** — a `google_compute_global_address` resource with the specified name, address type, purpose, and network configuration

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **Compute Engine API enabled** (`compute.googleapis.com`) on the target project
- **An existing VPC network** — required only for INTERNAL addresses, referenced via `network`
- **IAM permissions** — `roles/compute.networkAdmin` or equivalent on the target project

## Quick Start

Create a file `global-address.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: prod-lb-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.prod-lb-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: prod-lb-ip
```

Deploy:

```shell
openmcf apply -f global-address.yaml
```

This reserves a public IPv4 address that you can reference in global forwarding rules, HTTP(S) load balancers, or DNS A records.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the address is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `addressName` | `string` | Name of the global address resource in GCP. | 1-63 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `address` | `string` | — | Specific IP to reserve. Omit to let GCP assign one automatically. For VPC_PEERING, this is the start of the CIDR range. |
| `addressType` | `string` | `EXTERNAL` | `EXTERNAL` for a public IP or `INTERNAL` for a private IP range within a VPC. |
| `ipVersion` | `string` | `IPV4` | IP version: `IPV4` or `IPV6`. |
| `network` | `StringValueOrRef` | — | VPC network name or self-link. Required when `addressType` is `INTERNAL`. Can reference a GcpVpc resource. |
| `prefixLength` | `int32` | — | CIDR prefix length (8-29). Required when `purpose` is `VPC_PEERING`. A `/20` reserves 4,096 IPs. |
| `purpose` | `string` | — | Purpose of an INTERNAL address: `VPC_PEERING` (managed-service private networking) or `PRIVATE_SERVICE_CONNECT` (PSC endpoint). Leave empty for EXTERNAL addresses. |
| `description` | `string` | — | Human-readable description of the address reservation. |

## Examples

### External Static IP for Load Balancer

The simplest use case — reserve a public IPv4 address:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: web-lb-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.web-lb-ip
spec:
  projectId:
    value: my-prod-project-123
  addressName: web-lb-ip
  description: Static IP for production HTTPS load balancer
```

### Internal VPC Peering Range for Managed Services

Reserve a `/20` private CIDR block for Cloud SQL, Redis, AlloyDB, and Filestore private networking:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: managed-services-range
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.managed-services-range
spec:
  projectId:
    value: my-prod-project-123
  addressName: managed-services-range
  addressType: INTERNAL
  purpose: VPC_PEERING
  prefixLength: 20
  network:
    value: prod-vpc
  description: /20 range for VPC peering with Google managed services
```

### Private Service Connect Endpoint

Reserve an internal IP for private connectivity to Google APIs or third-party services:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: psc-google-apis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.psc-google-apis
spec:
  projectId:
    value: my-prod-project-123
  addressName: psc-google-apis
  addressType: INTERNAL
  purpose: PRIVATE_SERVICE_CONNECT
  network:
    value: prod-vpc
  description: PSC endpoint for private Google API access
```

### Cross-Resource Reference (Using GcpProject Output)

Reference a project ID from a GcpProject resource instead of hardcoding:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: lb-ip-with-ref
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.lb-ip-with-ref
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  addressName: lb-ip
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `address` | `string` | The reserved IP address. For EXTERNAL, this is a public IP (e.g., `34.120.1.2`). For INTERNAL VPC_PEERING, this is the first IP in the reserved range. |
| `selfLink` | `string` | Full self-link URL of the global address (e.g., `projects/my-project/global/addresses/prod-lb-ip`). Used to reference this address in forwarding rules. |
| `creationTimestamp` | `string` | RFC 3339 timestamp of when the address was created. |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and enables the Compute Engine API
- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network for internal address reservations and can enable Private Services Access
- [GcpCloudCdn](/docs/catalog/gcp/gcpcloudcdn) — uses an external global address as the frontend IP for CDN-enabled load balancers
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — uses a VPC_PEERING range for private IP connectivity to database instances
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — GKE clusters benefit from private service networking enabled by VPC peering ranges
- [GcpCertManagerCert](/docs/catalog/gcp/gcpcertmanagercert) — provisions managed SSL certificates that attach to the same load balancer using this IP
