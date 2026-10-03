# GCP Subnetwork

Deploys a subnetwork (`google_compute_subnetwork`) in a custom-mode VPC — the regional address space workloads actually live in: a primary IPv4 range for VM interfaces, secondary ranges for alias IPs (GKE pods and services), optional IPv6, special-purpose roles (proxy-only, Private Service Connect), and VPC Flow Logs.

## What Gets Created

When you deploy a GcpSubnetwork resource, Planton provisions:

- **Subnetwork** — a `google_compute_subnetwork` in the referenced VPC and region, with all configured ranges, purpose, IPv6, and logging
- **API enablement** — the Compute Engine API is enabled on the project if it is not already (never disabled on destroy)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing custom-mode VPC** — referenced via `vpcSelfLink` (a GcpVpcNetwork resource or a literal self-link)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `subnetwork.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSubnetwork
metadata:
  name: app-subnet
spec:
  projectId:
    value: my-gcp-project-123
  vpcSelfLink:
    valueFrom:
      kind: GcpVpcNetwork
      name: my-vpc
      fieldPath: status.outputs.network_self_link
  subnetworkName: app-subnet
  region: us-central1
  ipCidrRange: 10.10.0.0/20
  privateIpGoogleAccess: true
```

Deploy:

```shell
planton apply -f subnetwork.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `vpcSelfLink` | `StringValueOrRef` | The parent VPC network. Can reference a GcpVpcNetwork resource. Immutable. |
| `subnetworkName` | `string` | Name in GCP (RFC1035, 1-63 chars). Immutable. |
| `region` | `string` | Region the subnet lives in. Immutable. |
| `ipCidrRange` | `string` | Primary IPv4 CIDR (e.g. `10.10.0.0/20`). Required except for `IPV6_ONLY` subnets. Expandable in place; shrinking recreates. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | Project that owns the subnet. Can reference a GcpProject. Immutable. |
| `description` | `string` | `""` | What this subnet carries. Immutable. |
| `purpose` | `string` | `PRIVATE` | `PRIVATE`, `REGIONAL_MANAGED_PROXY` (proxy-only subnet for regional Envoy LBs), `GLOBAL_MANAGED_PROXY`, `PRIVATE_SERVICE_CONNECT`, `PEER_MIGRATION`, `PRIVATE_NAT`. |
| `role` | `string` | — | `ACTIVE` or `BACKUP`; only on `REGIONAL_MANAGED_PROXY` subnets (drain-and-swap staging). |
| `secondaryIpRanges` | `list` | `[]` | Named alias-IP ranges (up to 170) — how GKE gets pod/service IPs. Consumers select a range by `rangeName`. |
| `privateIpGoogleAccess` | `bool` | `false` | Let VMs without external IPs reach Google APIs internally. Effectively mandatory for private-only subnets. |
| `privateIpv6GoogleAccess` | `string` | GCP default | IPv6 counterpart: `DISABLE_GOOGLE_ACCESS`, `ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE`, or `ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE`. |
| `stackType` | `string` | `IPV4_ONLY` | `IPV4_ONLY`, `IPV4_IPV6` (dual-stack), `IPV6_ONLY`. |
| `ipv6AccessType` | `string` | — | `EXTERNAL` (internet-routable GUAs) or `INTERNAL` (VPC-internal ULAs). Required for IPv6-carrying stack types. Immutable. |
| `externalIpv6Prefix` | `string` | Google-allocated | Pin a specific external /64 (EXTERNAL access only). Immutable. |
| `allowSubnetCidrRoutesOverlap` | `bool` | `false` | Permit the subnet CIDR to overlap routes to destinations outside the VPC (deliberate address reclaims only). |
| `sendSecondaryIpRangeIfEmpty` | `bool` | `false` | Safety latch: `true` makes an empty `secondaryIpRanges` list REMOVE existing ranges on update; `false` leaves them untouched. |
| `logConfig` | object | off | VPC Flow Logs — presence enables logging; see below. |

### VPC Flow Logs (`logConfig`)

| Field | Default | Description |
|-------|---------|-------------|
| `aggregationInterval` | `INTERVAL_5_SEC` | `INTERVAL_5_SEC` … `INTERVAL_15_MIN` — longer means fewer, larger log entries |
| `flowSampling` | `0.5` | Fraction of flows sampled, 0.0-1.0 (1.0 = forensics-grade, at real Logging cost) |
| `metadata` | `INCLUDE_ALL_METADATA` | `EXCLUDE_ALL_METADATA`, `INCLUDE_ALL_METADATA`, or `CUSTOM_METADATA` (+ `metadataFields`) |
| `filterExpr` | all flows | CEL expression selecting flows, e.g. `connection.dest_port == 443` |

## Sizing Guidance

GCP reserves 4 addresses per primary range. The primary range can be **expanded** in place (e.g. /24 → /20) but never shrunk — start with room to grow:

- General workloads: `/24` (251 usable) to `/20` (4,091 usable)
- GKE nodes: `/20` primary; pods commonly `/14`-`/18` secondary; services `/20` secondary
- Proxy-only subnets: `/23` minimum (Google's recommendation)

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subnetwork_self_link` | `string` | Self-link — the value GKE clusters, instances, and other consumers reference |
| `subnetwork_name` | `string` | Name in GCP (referenced by Cloud Run Direct VPC egress and other by-name consumers) |
| `region` | `string` | Region of the subnet |
| `ip_cidr_range` | `string` | Primary IPv4 CIDR (empty for IPv6-only) |
| `secondary_ranges` | `list` | Names + CIDRs of secondary ranges (GKE selects pod/service ranges by name) |
| `gateway_address` | `string` | IPv4 address of the subnet's default gateway |
| `subnetwork_id` | `string` | Server-assigned numeric ID |
| `internal_ipv6_prefix` / `external_ipv6_prefix` | `string` | IPv6 prefixes actually allocated (empty without IPv6) |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Almost everything identity-shaped is immutable**: name, project, region, network, and description are ForceNew — changing any of them recreates the subnet, an outage for everything addressed in it. Plan names and regions as carefully as CIDRs.
- **The CIDR expansion asymmetry**: growing `ipCidrRange` is an in-place update; shrinking destroys and recreates. This is why starting small "to be safe" is backwards — start with room.
- **Secondary-range safety latch**: by default an empty `secondaryIpRanges` list on update does NOT remove existing ranges (a partial manifest cannot wipe GKE pod ranges). Set `sendSecondaryIpRangeIfEmpty: true` only when removal is the intent.
- **Proxy-only subnets are load-balancer prerequisites**: a region cannot host a regional Application Load Balancer until a `REGIONAL_MANAGED_PROXY` subnet exists in the VPC there.
- **Flow logs cost real money at scale**: full sampling on a busy subnet generates significant Cloud Logging volume — tune `flowSampling` and `filterExpr` deliberately.
- **Preview-surface note**: `allowSubnetCidrRoutesOverlap` is preview-stage on the current provider line; the modules select the beta provider so it is available without a retrofit.

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the parent network
- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — consumes the subnet + secondary ranges by reference
- [GcpRouterNat](/docs/catalog/gcp/gcprouternat) — gives private subnets outbound internet access
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the subnet

## Additional Resources

- [VPC Subnets Documentation](https://cloud.google.com/vpc/docs/subnets)
- [Proxy-only subnets](https://cloud.google.com/load-balancing/docs/proxy-only-subnets)
- [VPC Flow Logs](https://cloud.google.com/vpc/docs/flow-logs)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
