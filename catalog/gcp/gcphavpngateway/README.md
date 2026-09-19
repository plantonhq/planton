# GCP HA VPN Gateway

Creates a Google Cloud HA VPN gateway and the Cloud Router its tunnels speak BGP through — the Google end of every IPsec VPN from a VPC. The gateway has two interfaces, each with a public IP the other side configures against; Google's 99.99% SLA holds when both carry a tunnel to redundant peers. Declare the gateway once per VPC and region, then declare each site or peer cloud as a `GcpHaVpnConnection` that references it.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **HA VPN gateway** -- the `compute_ha_vpn_gateway` with two interfaces (public IPs, or Interconnect attachments when pinned)
- **Cloud Router** -- the `compute_router` with the BGP ASN and default advertisement every session inherits
- **API enablement** -- `compute.googleapis.com`, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Network

- **The VPC must exist** -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Pick a private ASN** (64512-65534 or 4200000000-4294967294) that differs from every peer's; it is fixed for the router's life.
- **For HA VPN over Interconnect**, the VLAN attachments must be encrypted attachments in the same region.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnGateway
metadata:
  name: hub-vpn
spec:
  region: us-central1
  network:
    value: projects/acme-net/global/networks/hub
  router:
    bgp:
      asn: 64514
```

```shell
planton apply -f ha-vpn-gateway.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | The region of the gateway and router. Immutable. |
| `network` | `StringValueOrRef` | The VPC, as a `GcpVpcNetwork` reference or its self link. Immutable. |
| `router.bgp.asn` | `uint32` | The private ASN Google speaks as. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a `GcpProject` reference or a literal. |
| `gatewayName` | `string` | `metadata.name` | Name of the gateway. Immutable. |
| `description` | `string` | — | Immutable. |
| `gatewayIpVersion` | `string` | `IPV4` | `IPV4` or `IPV6` public interface addresses. Immutable. |
| `stackType` | `string` | `IPV4_ONLY` | `IPV4_ONLY`, `IPV4_IPV6`, or `IPV6_ONLY` traffic inside the tunnels. Immutable. |
| `vpnInterfaces` | `list` | — | Pin interfaces 0/1 to Interconnect attachments (needs an encrypted router). Immutable. |
| `labels` | `map` | — | User labels on the gateway; merged with the platform's. Mutable. |
| `resourceManagerTags` | `map` | — | Create-time tags on the gateway. |
| `router.name` | `string` | the gateway's name | Immutable. |
| `router.description` | `string` | — | Mutable. |
| `router.bgp.advertiseMode` | `string` | `DEFAULT` | `DEFAULT` (all subnets) or `CUSTOM` (only the groups and ranges below). |
| `router.bgp.advertisedGroups` | `list` | — | `ALL_SUBNETS`, in CUSTOM mode. |
| `router.bgp.advertisedIpRanges` | `list` | — | Custom ranges `{range, description}`, in CUSTOM mode. |
| `router.bgp.keepaliveInterval` | `int32` | `20` | 20-60 seconds. |
| `router.bgp.identifierRange` | `string` | Google-assigned | Link-local /30+ for BGP identifiers. |
| `router.encryptedInterconnectRouter` | `bool` | `false` | Required with `vpnInterfaces`. Immutable. |
| `router.resourceManagerTags` | `map` | — | Create-time tags on the router. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` for both resources. |

### Validation Rules

- **`router.bgp.asn`** is a private ASN; custom groups/ranges only in `CUSTOM` mode; `keepaliveInterval` 20-60.
- **`vpnInterfaces`** at most 2, unique ids, and only with `router.encryptedInterconnectRouter: true`.
- Names match `^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `gateway_self_link` | `string` | What a `GcpHaVpnConnection`'s `gateway` references |
| `gateway_name` | `string` | The gateway's name |
| `region` | `string` | What a `GcpHaVpnConnection`'s `region` references |
| `interface_0_ip_address` | `string` | Public IP of interface 0 |
| `interface_1_ip_address` | `string` | Public IP of interface 1 |
| `router_name` | `string` | What a `GcpHaVpnConnection`'s `router` references |
| `router_self_link` | `string` | The router's self link |
| `router_asn` | `uint32` | The ASN the router speaks as |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A recreated gateway has new public IPs** — every on-premises device must be reconfigured. Network, region, IP version, stack type, and interface pinning are immutable.
- **Deleting the gateway fails while any connection's tunnels reference it** — destroy the `GcpHaVpnConnection`s first; a chart does this in order.
- **Cost**: the gateway itself is free; tunnels bill per hour and traffic bills as egress — both declared on the connection.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpHaVpnConnection](/docs/catalog/gcp/gcphavpnconnection) — a site or peer cloud connected through this gateway
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the gateway attaches to
- [GcpRouterNat](/docs/catalog/gcp/gcprouternat) — a second router on the same network for NAT
- [GcpVpcPeering](/docs/catalog/gcp/gcpvpcpeering) — shares the VPN's routes with peered networks via custom routes

## Additional Resources

- [HA VPN overview](https://cloud.google.com/network-connectivity/docs/vpn/concepts/overview)
- [HA VPN topologies](https://cloud.google.com/network-connectivity/docs/vpn/concepts/topologies)
- [Cloud Router overview](https://cloud.google.com/network-connectivity/docs/router/concepts/overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
