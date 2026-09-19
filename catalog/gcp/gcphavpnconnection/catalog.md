# GCP HA VPN Connection

Connects a Google Cloud HA VPN gateway (`GcpHaVpnGateway`) to ONE peer -- an on-premises site, another cloud, or another Google Cloud VPC -- with one to four IPsec tunnels, each carrying a BGP session on the gateway's Cloud Router. Declare one connection per site and add or remove sites without touching the gateway. Two tunnels, one from each gateway interface to two addresses of the peer, is Google's recommended 99.99% topology; the tunnels advertise the same routes so traffic fails over in seconds, or sub-second with BFD.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **External VPN gateway** (external peer only) -- the `compute_external_vpn_gateway` holding the device's public addresses
- **VPN tunnels** -- one `compute_vpn_tunnel` per `tunnels[]` entry, from a gateway interface to a peer interface
- **Router interfaces** -- one `compute_router_interface` per tunnel: the Google end of the session's link-local /30
- **BGP peers** -- one `compute_router_peer` per tunnel, with route advertisement, BFD, and MD5 authentication when declared

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP HA VPN Gateway and the peer

- **The gateway must exist** -- declare it with `GcpHaVpnGateway` and reference its `gateway_self_link`, `router_name`, and `region` outputs (the registry installs one as this kind's prerequisite in E2E).
- **For an on-premises or other-cloud peer**, have the device's public addresses (one, two, or four), its BGP ASN, and a pre-shared key per tunnel; the device is configured toward the gateway's two `interface_*_ip_address` outputs with the gateway's `router_asn` as its neighbor.
- **For a Google Cloud peer**, the other VPC declares its own `GcpHaVpnGateway` and a `GcpHaVpnConnection` pointing back at this side's gateway with the same pre-shared keys.
- **Pre-shared keys are sensitive.** Wire them from a secrets manager (`${secrets-group.<name>.<key>}`) rather than writing them into the manifest.

## Deploy

### Console

Open the deployment store, find **GCP HA VPN Connection**, and click **Deploy**. The creation wizard walks you through the gateway, the peer, and the tunnels. Start from the **Two Tunnel To Onprem** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnConnection
metadata:
  name: hq
  org: acme-corp
  env: prod
spec:
  gateway:
    value: projects/acme-net/regions/us-central1/vpnGateways/hub-vpn
  router:
    value: hub-vpn
  region:
    value: us-central1
  peer:
    externalGateway:
      redundancyType: TWO_IPS_REDUNDANCY
      interfaces:
        - id: 0
          ipAddress: 203.0.113.10
        - id: 1
          ipAddress: 203.0.113.11
  tunnels:
    - name: hq-tunnel-0
      vpnGatewayInterface: 0
      peerExternalGatewayInterface: 0
      sharedSecret: ${secrets-group.hq-vpn.psk-0}
      bgpSession:
        interfaceIpRange: 169.254.10.1/30
        peerAsn: 65001
    - name: hq-tunnel-1
      vpnGatewayInterface: 1
      peerExternalGatewayInterface: 1
      sharedSecret: ${secrets-group.hq-vpn.psk-1}
      bgpSession:
        interfaceIpRange: 169.254.11.1/30
        peerAsn: 65001
```

```shell
planton apply -f ha-vpn-connection.yaml
```

This connects the hub gateway to a two-address HQ device with two tunnels and two BGP sessions -- the 99.99% shape. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the connection references its `GcpHaVpnGateway` three times via ValueFromRef -- the gateway, its router, and its region -- so the three can never disagree:

```yaml
spec:
  gateway:
    valueFrom:
      kind: GcpHaVpnGateway
      name: hub-vpn
      fieldPath: status.outputs.gateway_self_link
  router:
    valueFrom:
      kind: GcpHaVpnGateway
      name: hub-vpn
      fieldPath: status.outputs.router_name
  region:
    valueFrom:
      kind: GcpHaVpnGateway
      name: hub-vpn
      fieldPath: status.outputs.region
  peer:
    gcpGateway:
      valueFrom:
        kind: GcpHaVpnGateway
        name: spoke-vpn
        fieldPath: status.outputs.gateway_self_link
```

For a Google-to-Google VPN, the spoke declares the mirror connection pointing at `hub-vpn`. The InfraPipeline deploys both gateways, then both connections; Google pairs the interfaces itself.

## Key Configuration

These are the most important decisions when configuring a connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Peer** -- an external device (its addresses and redundancy type, which fixes how many tunnels the HA shape needs) or another Google Cloud gateway. Exactly one.

**Tunnels** -- one to four. Each leaves from a gateway interface (0 or 1) and, for an external peer, lands on one of the device's interfaces. Every field except labels is immutable: a new pre-shared key, cipher set, or interface pairing recreates the tunnel, so rotate a key by adding a tunnel and removing the old one.

**BGP session** -- each tunnel's `interfaceIpRange` is a link-local /30 (`169.254.x.y/30`), unique on the router; Google takes the address as its end and the peer takes the other. `peerAsn` is the device's ASN. Set different `advertisedRoutePriority` values on the tunnels to make the peer prefer one (active/passive) instead of splitting traffic (the default).

**BFD** -- sub-second failure detection; leave unset to rely on BGP keepalives (up to 60 s by default).

**Deletion policy** -- `DELETE` (default) removes everything and disconnects the site; `PREVENT` fails the destroy, the guard for a production link; `ABANDON` leaves the tunnels live.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpHaVpnGateway** | `gateway` | `status.outputs.gateway_self_link` |
| **GcpHaVpnGateway** | `router` | `status.outputs.router_name` |
| **GcpHaVpnGateway** | `region` | `status.outputs.region` |
| **GcpHaVpnGateway** | `peer.gcpGateway` | `status.outputs.gateway_self_link` (the OTHER side's gateway) |
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `tunnel_self_links` | The tunnels' self links, in spec order | Tooling |
| `tunnel_names` | The tunnels' names, in spec order | `gcloud compute vpn-tunnels describe` |
| `router_interface_names` | The router interfaces, in spec order | Tooling |
| `bgp_peer_names` | The BGP peers, in spec order | `gcloud compute routers get-status` reports each peer's BGP state under this name |
| `external_gateway_self_link` | The external VPN gateway (empty for a Google peer) | Tooling |
| `gateway_self_link` | The gateway the tunnels leave from | Tooling |
| `router_name` | The router the sessions run on | Tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Two tunnels to on-premises** -- `TWO_IPS_REDUNDANCY`, one tunnel per gateway interface, two BGP sessions: Google's recommended 99.99% topology. Start from the **Two Tunnel To Onprem** preset.

**Four tunnels to on-premises** -- `FOUR_IPS_REDUNDANCY`, two tunnels per gateway interface. Start from the **Four Tunnel To Onprem** preset.

**Google to Google** -- two VPCs, each with a gateway and a connection pointing at the other. Start from the **Gcp To Gcp** preset.

## Works With

- [**GCP HA VPN Gateway**](/cloud-catalog/gcp-ha-vpn-gateway) -- the gateway and router this connection rides
- [**GCP VPC Peering**](/cloud-catalog/gcp-vpc-peering) -- shares the routes this connection learns with peered networks via custom routes
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the network on the Google side
