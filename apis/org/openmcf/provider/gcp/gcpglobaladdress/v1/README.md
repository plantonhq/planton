# GCP Global Address

Deploys a GCP global address reservation (`google_compute_global_address`) for external static IPs, internal VPC peering ranges, or Private Service Connect endpoints, with address type, IP version, network, prefix length, and purpose configuration.

## What Gets Created

When you deploy a GcpGlobalAddress resource, OpenMCF provisions:

- **Global Address** — a `google_compute_global_address` resource in the specified project, reserving either a public IP (EXTERNAL) or a private IP range (INTERNAL)

No additional supporting resources (API enablement, networking connections, etc.) are created. The module assumes the Compute Engine API is already enabled on the target project.

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing GCP project** — referenced via `projectId`
- **Compute Engine API enabled** (`compute.googleapis.com`) on the target project
- **IAM permissions** — `roles/compute.networkAdmin` on the target project
- **An existing VPC network** — required only for INTERNAL addresses (referenced via `network`)

## Quick Start

Create a file `global-address.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGlobalAddress
metadata:
  name: lb-static-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGlobalAddress.lb-static-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: prod-lb-external-ip
  addressType: EXTERNAL
  ipVersion: IPV4
  description: Static IP for production HTTPS load balancer
```

Deploy:

```shell
openmcf apply -f global-address.yaml
```

This reserves a static external IPv4 address that you can attach to an HTTP(S) load balancer or global forwarding rule.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID where the global address is created. Can reference a GcpProject resource. | Required |
| `addressName` | `string` | Name of the global address resource in GCP. | 1-63 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number (RFC 1035) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `address` | `string` | — | Specific IP address to reserve. Omit to let GCP assign one automatically. For INTERNAL VPC_PEERING, this is the start of the reserved CIDR range. |
| `addressType` | `string` | `EXTERNAL` | Type of address: `EXTERNAL` for public IPs, `INTERNAL` for private IP ranges within a VPC. |
| `description` | `string` | `""` | Human-readable description for the address reservation. |
| `ipVersion` | `string` | `IPV4` | IP version: `IPV4` or `IPV6`. |
| `network` | `StringValueOrRef` | — | VPC network for INTERNAL addresses. Accepts a network name or self-link URL. Can reference a GcpVpc resource. Required when `addressType` is `INTERNAL`. |
| `prefixLength` | `int32` | — | CIDR prefix length (8-29) for the reserved range. Required when `purpose` is `VPC_PEERING`. E.g., `20` reserves a /20 range (4,096 IPs). |
| `purpose` | `string` | `""` | Purpose of this INTERNAL address: `VPC_PEERING`, `PRIVATE_SERVICE_CONNECT`, or empty. Only valid for INTERNAL addresses. |

### Validation Rules

- **Purpose requires INTERNAL**: The `purpose` field can only be set when `addressType` is `INTERNAL`.
- **VPC_PEERING requires prefix length**: When `purpose` is `VPC_PEERING`, `prefixLength` must be specified.
- **INTERNAL requires network**: When `addressType` is `INTERNAL`, the `network` field is required.

## When to Use Each Address Type

| Use Case | Address Type | Purpose | Example |
|----------|-------------|---------|---------|
| Static IP for HTTP(S) LB | EXTERNAL | — | Public IP attached to a global forwarding rule |
| Static IP for CDN | EXTERNAL | — | Anycast IP for Cloud CDN |
| Cloud SQL private networking | INTERNAL | VPC_PEERING | /20 range for managed service peering |
| Redis/AlloyDB private access | INTERNAL | VPC_PEERING | /20 range for private services |
| Private Service Connect | INTERNAL | PRIVATE_SERVICE_CONNECT | Single IP for PSC endpoint |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `address` | `string` | The reserved IP address or the start of the reserved CIDR range |
| `self_link` | `string` | Self-link URL of the global address resource (e.g., `projects/my-project/global/addresses/lb-external-ip`) |
| `creation_timestamp` | `string` | RFC 3339 timestamp of when the address was reserved |

## Deployment Methods

OpenMCF supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **ForceNew**: All fields except labels are ForceNew in the underlying GCP API. Any change to the address configuration destroys and recreates the resource. Plan changes carefully.
- **Network lock**: When an INTERNAL address references a VPC network, that network cannot be deleted while the reserved range exists.
- **IPv6 availability**: IPv6 global addresses are only available for EXTERNAL addresses with premium network tier.

## Examples

For comprehensive examples, see [`examples.md`](examples.md), including:

- Minimal external static IP
- External IPv6 address
- Internal VPC peering range for Cloud SQL
- Private Service Connect address
- Full configuration with all fields

## Related Components

- [GcpVpc](/docs/catalog/gcp/gcpvpc) — provides the VPC network referenced by INTERNAL addresses
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and enables the Compute Engine API
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — managed database that uses VPC peering ranges for private connectivity
- [GcpCloudCdn](/docs/catalog/gcp/gcpcloudcdn) — CDN that uses external static IPs

## Additional Resources

- [Reserving a Static External IP Address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address)
- [Reserving an Internal IP Range](https://cloud.google.com/vpc/docs/configure-private-services-access)
- [Private Service Connect Overview](https://cloud.google.com/vpc/docs/private-service-connect)
- [Global Addresses API Reference](https://cloud.google.com/compute/docs/reference/rest/v1/globalAddresses)

## Support

For issues, questions, or contributions, please refer to the OpenMCF documentation or open an issue in the repository.
