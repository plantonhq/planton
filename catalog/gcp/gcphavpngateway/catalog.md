# GCP HA VPN Gateway

Creates a Google Cloud HA VPN gateway and the Cloud Router its tunnels speak BGP through: the Google end of every IPsec VPN from a VPC. The gateway has two interfaces, each with its own public IP the other side is configured to reach, and Google's 99.99% availability SLA holds when both interfaces carry a tunnel to redundant peers. Declare the gateway ONCE per VPC and region, then declare each site or peer cloud it connects to as a `GcpHaVpnConnection` that references it -- the tunnels, pre-shared keys, and BGP sessions live there and can be added or removed without touching the gateway.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **HA VPN gateway** -- the `compute_ha_vpn_gateway` with two interfaces (public IPs, or Interconnect attachments when pinned)
- **Cloud Router** -- the `compute_router` carrying the BGP ASN and default advertisement every session inherits
- **API enablement** -- `compute.googleapis.com`, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Network

- **The VPC must exist** -- declare it with `GcpVpcNetwork` and reference its `network_self_link` output, or pass the self link as a literal.
- **Pick a private ASN** (64512-65534 or 4200000000-4294967294) that differs from every peer's. It is fixed for the router's life.
- **For HA VPN over Cloud Interconnect**, the VLAN attachments must be encrypted attachments in the same region, and the router must be an encrypted-interconnect router.

## Deploy

### Console

Open the deployment store, find **GCP HA VPN Gateway**, and click **Deploy**. The creation wizard walks you through the network, region, and ASN. Start from the **Internet Facing Gateway** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnGateway
metadata:
  name: hub-vpn
  org: acme-corp
  env: prod
spec:
  region: us-central1
  network:
    value: projects/acme-net/global/networks/hub
  router:
    bgp:
      asn: 64514
  deletionPolicy: PREVENT
```

```shell
planton apply -f ha-vpn-gateway.yaml
```

This creates the hub's VPN gateway in `us-central1` with a router speaking ASN 64514, guarded against accidental destroy. A Stack Job tracks the provisioning in real time; the two `interface_*_ip_address` outputs are what the on-premises team configures next.

### InfraChart

When deploying as part of a multi-resource environment, the gateway references its `GcpVpcNetwork` via ValueFromRef, and every `GcpHaVpnConnection` references the gateway's outputs:

```yaml
spec:
  region: us-central1
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: hub
      fieldPath: status.outputs.network_self_link
  router:
    bgp:
      asn: 64514
```

The InfraPipeline deploys the network, then the gateway and its router, then every connection that points at `status.outputs.gateway_self_link` -- and destroys connections before the gateway, which is the order Google requires.

## Key Configuration

These are the most important decisions when configuring a gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Network and region** -- where the gateway lives; every tunnel of every connection is in this region. Both immutable.

**ASN** -- the number Google Cloud speaks as toward every peer. Private, unique among the peers, and fixed for the router's life.

**Advertisement** -- `DEFAULT` advertises every subnet of the VPC to every peer; `CUSTOM` advertises only the groups and ranges you list. Each connection's session may override this for its own peer.

**Stack type and IP version** -- `IPV4_ONLY` is the norm; `IPV4_IPV6` carries both inside the tunnels (the VPC needs an internal IPv6 range); `gatewayIpVersion: IPV6` gives the interfaces public IPv6 addresses for a peer that reaches Google over IPv6. All immutable.

**Interconnect pinning** -- `vpnInterfaces` binds each interface to an encrypted VLAN attachment for HA VPN over Cloud Interconnect. Leave empty for the ordinary internet-facing gateway.

**Deletion policy** -- `DELETE` (default) deletes both resources and fails while any connection's tunnels still reference the gateway; `PREVENT` fails the destroy, the guard for the gateway every site is configured against; `ABANDON` leaves both live.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `gateway_self_link` | The gateway's self link | A `GcpHaVpnConnection`'s `gateway`; another gateway's connection's `peer.gcpGateway` |
| `gateway_name` | The gateway's name | Tooling |
| `region` | The region | A `GcpHaVpnConnection`'s `region` |
| `interface_0_ip_address` | Public IP of interface 0 | The on-premises device's first tunnel target |
| `interface_1_ip_address` | Public IP of interface 1 | The on-premises device's second tunnel target |
| `router_name` | The Cloud Router's name | A `GcpHaVpnConnection`'s `router` |
| `router_self_link` | The Cloud Router's self link | Tooling |
| `router_asn` | The ASN Google speaks as | The on-premises device's BGP neighbor ASN |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Internet-facing gateway** -- the everyday shape: two public IPv4 interfaces, a router with a private ASN advertising all subnets. Start from the **Internet Facing Gateway** preset.

**Dual-stack gateway** -- `IPV4_IPV6` traffic with custom advertisement. Start from the **Dual Stack Gateway** preset.

**Interconnect-backed gateway** -- HA VPN over Cloud Interconnect: interfaces pinned to encrypted VLAN attachments on an encrypted router. Start from the **Interconnect Backed Gateway** preset.

## Works With

- [**GCP HA VPN Connection**](/cloud-catalog/gcp-ha-vpn-connection) -- a site or peer cloud connected through this gateway
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the network the gateway attaches to
- [**GCP Router NAT**](/cloud-catalog/gcp-router-nat) -- a second router on the same network for NAT
- [**GCP VPC Peering**](/cloud-catalog/gcp-vpc-peering) -- shares the VPN's routes with peered networks via custom routes
