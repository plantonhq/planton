# GCP VPC Peering

Manages one side of a Google Cloud VPC Network Peering — the peering entry on your network that points at another network, and the routes that side exchanges. Peered networks reach each other over internal IPs with no gateway and no bandwidth cap beyond the VMs' own. Set `peerNetwork` to create the peering; leave it empty to manage the route exchange of a peering Google created for you (Cloud SQL private IP, Memorystore).

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions exactly one of:

- **Network peering** (`peerNetwork` set) -- the `compute_network_peering` entry on your network with its route-exchange flags
- **Peering routes config** (`peerNetwork` empty) -- the `compute_network_peering_routes_config` on an existing peering named `peeringName`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the network's project (and on the peer's project when creating a peering to it).
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **Your network must exist** -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Subnet ranges must not overlap** between the two networks, and peering is not transitive.
- **The other side declares its own entry** -- a peering is ACTIVE only when both sides exist and point at each other.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVpcPeering
metadata:
  name: hub-to-spoke
spec:
  network:
    value: projects/acme-net/global/networks/hub
  peerNetwork:
    value: projects/acme-app/global/networks/spoke
  exportCustomRoutes: true
```

```shell
planton apply -f vpc-peering.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `network` | `StringValueOrRef` | This side's network, as a `GcpVpcNetwork` reference or its self link. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `peeringName` | `string` | `metadata.name` | The peering entry's name (or the existing peering's name in the routes-config form). Immutable. |
| `peerNetwork` | `StringValueOrRef` | — | The other network. Set → create the peering; empty → routes-config on an existing peering. Immutable. |
| `exportCustomRoutes` | `bool` | `false` | Export static and dynamic (VPN/Interconnect) routes to the peer. Mutable. |
| `importCustomRoutes` | `bool` | `false` | Import the peer's custom routes. Mutable. |
| `exportSubnetRoutesWithPublicIp` | `bool` | `true` | Export privately-used public-IP subnet routes. Immutable in the create form. |
| `importSubnetRoutesWithPublicIp` | `bool` | `false` | Import the peer's public-IP subnet routes. Immutable in the create form. |
| `stackType` | `string` | `IPV4_ONLY` | `IPV4_ONLY` or `IPV4_IPV6`. Create form only. |
| `updateStrategy` | `string` | `INDEPENDENT` | `INDEPENDENT` or `CONSENSUS` (a change applies only when both sides agree). Create form only. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. Create form only. |

### Validation Rules

- **`peeringName`** matches `^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`.
- **`stackType`**, **`updateStrategy`**, **`deletionPolicy`** are rejected on the routes-config form (no `peerNetwork`).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `peering_name` | `string` | The peering entry's name on this side |
| `network` | `string` | This side's network |
| `state` | `string` | `ACTIVE` / `INACTIVE` (create form; empty on routes-config) |
| `state_details` | `string` | GCP's explanation of the state (create form) |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Declare both sides.** Two `GcpVpcPeering` resources, one per network, when you own both; the provider serializes them safely.
- **Custom routes cross only when the exporter exports AND the importer imports.**
- **The routes-config form's destroy is a no-op in GCP** — the peering keeps its current flags.
- **Cost**: peering itself is free; traffic between peered networks bills at Google's inter-zone/inter-region rates as the workloads' egress.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the networks on both sides
- [GcpServiceNetworkingConnection](/docs/catalog/gcp/gcpservicenetworkingconnection) — the Google-managed peering the routes-config form tunes
- [GcpHaVpnGateway](/docs/catalog/gcp/gcphavpngateway) — the dynamic routes `exportCustomRoutes` shares with the peer

## Additional Resources

- [VPC Network Peering overview](https://cloud.google.com/vpc/docs/vpc-peering)
- [Using VPC Network Peering](https://cloud.google.com/vpc/docs/using-vpc-peering)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
