# GCP Regional Address

Deploys a GCP regional address reservation (`google_compute_address`) for external static IPs (Cloud NAT, regional LBs, VMs) or internal regional IPs (GCE endpoints, internal LB VIPs, VPC peering, IPsec interconnect), with address type, IP version, network, subnetwork, prefix length, purpose, and network tier configuration.

## What Gets Created

When you deploy a GcpAddress resource, Planton provisions:

- **Compute Engine API enablement** — a `google_project_service` resource that activates `compute.googleapis.com` on the target project
- **Regional Address** — a `google_compute_address` resource in the specified region, reserving either a public IP (EXTERNAL) or a private IP/range (INTERNAL)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **An existing VPC network** — required for INTERNAL addresses with VPC_PEERING or IPSEC_INTERCONNECT purpose (referenced via `network`)
- **An existing subnetwork** — required for INTERNAL addresses with GCE_ENDPOINT or DNS_RESOLVER purpose (referenced via `subnetwork`)

## Quick Start

Create a file `regional-address.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAddress
metadata:
  name: nat-external-ip
spec:
  projectId:
    value: my-gcp-project-123
  addressName: nat-external-ip
  region: us-central1
  addressType: EXTERNAL
  ipVersion: IPV4
  description: Static IP for Cloud NAT in us-central1
```

Deploy:

```shell
planton apply -f regional-address.yaml
```

This reserves a static external IPv4 address in `us-central1` that you can attach to a Cloud NAT gateway or regional load balancer.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `addressName` | `string` | Name of the regional address resource in GCP. | 1-63 chars, lowercase letters/numbers/hyphens, RFC 1035 |
| `region` | `string` | GCP region for the reservation (e.g. `us-central1`). | Valid GCP region name. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. |
| `address` | `string` | — | Specific IP to reserve. Omit to let GCP assign one. |
| `addressType` | `string` | `EXTERNAL` | `EXTERNAL` for public IPs, `INTERNAL` for private IPs/ranges. |
| `description` | `string` | `""` | Human-readable description. |
| `ipVersion` | `string` | `IPV4` | `IPV4` or `IPV6`. |
| `network` | `StringValueOrRef` | — | VPC network for INTERNAL VPC_PEERING / IPSEC_INTERCONNECT. Can reference GcpVpcNetwork. |
| `subnetwork` | `StringValueOrRef` | — | Subnetwork for INTERNAL GCE_ENDPOINT / DNS_RESOLVER. Can reference GcpSubnetwork. |
| `networkTier` | `string` | `PREMIUM` | `PREMIUM` or `STANDARD` — EXTERNAL only. |
| `prefixLength` | `int32` | — | CIDR prefix length (8-29) for peering/interconnect ranges. |
| `purpose` | `string` | `""` | INTERNAL purpose: `GCE_ENDPOINT`, `SHARED_LOADBALANCER_VIP`, `VPC_PEERING`, `IPSEC_INTERCONNECT`, or `DNS_RESOLVER`. |
| `ipv6EndpointType` | `string` | `""` | `VM` or `NETLB` — external IPv6 endpoint type. |
| `labels` | `map<string,string>` | `{}` | User labels, merged with the platform labels (platform wins on key conflicts). The one mutable surface. |
| `ipCollection` | `string` | `""` | BYOIP source: a PublicDelegatedPrefix URL for EXTERNAL addresses. |
| `deletionPolicy` | `string` | `DELETE` | What destroy does: `DELETE`, `PREVENT` (refuse), or `ABANDON` (keep the IP reserved, drop from management). |

### Validation Rules

- **network_tier on INTERNAL is rejected** — internal traffic always uses Premium tier.
- **purpose requires INTERNAL** — purpose can only be set when `addressType` is `INTERNAL`.
- **VPC_PEERING / IPSEC_INTERCONNECT requires network** — the `network` field is required for these purposes.
- **GCE_ENDPOINT / DNS_RESOLVER requires subnetwork** — the `subnetwork` field is required for these purposes.
- **SHARED_LOADBALANCER_VIP requires INTERNAL** — this purpose is only valid with `addressType: INTERNAL`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `address` | `string` | The reserved IP address or start of the reserved range |
| `self_link` | `string` | Self-link URL of the regional address resource |
| `name` | `string` | Name of the regional address resource in GCP |
| `region` | `string` | Plain region name from the spec (e.g. `us-central1`) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **ForceNew**: All fields except labels are ForceNew. Any change destroys and recreates the address — a recreated EXTERNAL address gets a new IP.
- **Regional vs global**: This kind models `google_compute_address` (regional). For global-scope addresses (HTTP(S) LB frontends, global VPC peering ranges, PSC), use [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress).
- **PRIVATE_SERVICE_CONNECT is global-only**: use GcpGlobalAddress for PSC endpoints.

## Related Kinds

- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — global-scope static IPs and VPC peering ranges
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — provides the VPC network for INTERNAL addresses
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — provides the subnetwork for GCE_ENDPOINT addresses
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project

## Additional Resources

- [Reserving a Static External IP Address](https://cloud.google.com/compute/docs/ip-addresses/reserve-static-external-ip-address)
- [Regional Addresses API Reference](https://cloud.google.com/compute/docs/reference/rest/v1/addresses)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
