# GCP HA VPN Connection

Connects a Google Cloud HA VPN gateway (`GcpHaVpnGateway`) to ONE peer — an on-premises site, another cloud, or another Google Cloud VPC — with one to four IPsec tunnels, each carrying a BGP session on the gateway's Cloud Router. Declare one connection per site; add or remove sites without touching the gateway. Two tunnels, one per gateway interface, to two peer addresses is Google's 99.99% shape.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **External VPN gateway** (external peer only) -- the `compute_external_vpn_gateway` holding the device's public addresses
- **VPN tunnels** -- one `compute_vpn_tunnel` per `tunnels[]` entry
- **Router interfaces** -- one `compute_router_interface` per tunnel (the link-local /30)
- **BGP peers** -- one `compute_router_peer` per tunnel, with BFD and MD5 when declared

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP HA VPN Gateway and the peer

- **The gateway must exist** -- declare it with `GcpHaVpnGateway` and reference its `gateway_self_link`, `router_name`, and `region` outputs.
- **For an external peer**, know the device's public addresses, its ASN, and the pre-shared key; configure the device toward the gateway's two `interface_*_ip_address` outputs.
- **For a Google peer**, the other VPC declares its own `GcpHaVpnGateway` and a `GcpHaVpnConnection` pointing back at this side's gateway.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHaVpnConnection
metadata:
  name: hq
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `gateway` | `StringValueOrRef` | The `GcpHaVpnGateway` (its `gateway_self_link`). Immutable. |
| `router` | `StringValueOrRef` | The gateway's router (its `router_name`). Immutable. |
| `region` | `StringValueOrRef` | The gateway's region (its `region` output). Immutable. |
| `peer` | `message` | Exactly one of `externalGateway` (a device's addresses and redundancy type) or `gcpGateway` (another `GcpHaVpnGateway`). |
| `tunnels` | `list` | 1-4 tunnels, each with `name`, `sharedSecret` (sensitive), and a `bgpSession` with `peerAsn`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `tunnels[].vpnGatewayInterface` | `int32` | `0` | Which gateway interface (0/1) the tunnel leaves from. |
| `tunnels[].peerExternalGatewayInterface` | `int32` | — | Which device interface it lands on (external peer only). |
| `tunnels[].ikeVersion` | `int32` | `2` | 1 or 2. |
| `tunnels[].localTrafficSelector` / `remoteTrafficSelector` | `list` | — | IPv4 CIDRs for route-based tunnels without BGP. |
| `tunnels[].cipherSuite` | `message` | Google defaults | Phase-1 and phase-2 algorithm restrictions. |
| `tunnels[].labels` | `map` | — | The one mutable tunnel field. |
| `tunnels[].description` | `string` | — | Immutable. |
| `tunnels[].bgpSession.*` | — | — | `name`, `interfaceIpRange` (link-local /30), `ipVersion`, `peerIpAddress`, `advertisedRoutePriority`, `advertiseMode` / `advertisedGroups` / `advertisedIpRanges`, `enable`, `enableIpv4`, `enableIpv6`, the four next-hop addresses, `customLearnedIpRanges`, `customLearnedRoutePriority`, `bfd`, `md5AuthenticationKey` (sensitive key), `importPolicies`, `exportPolicies`. |
| `resourceManagerTags` | `map` | — | Create-time tags on every tunnel and the external gateway. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` for every resource. |

### Validation Rules

- **`peer`**: exactly one arm. External interfaces match `redundancyType` (1 / 2 / 4), unique ids, one address each.
- **`tunnels`**: distinct names; distinct `interfaceIpRange`s; every tunnel names a `peerExternalGatewayInterface` with an external peer and none with a Google peer; the named interface exists.
- **`bgpSession`**: `peerAsn` ≥ 1; `interfaceIpRange` is `169.254.x.y/30`; BFD intervals 1000-30000, multiplier 5-16; priorities 0-65535 / 0-65335.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `tunnel_self_links` | `list<string>` | The tunnels' self links, in spec order |
| `tunnel_names` | `list<string>` | The tunnels' names, in spec order |
| `router_interface_names` | `list<string>` | The router interfaces, in spec order |
| `bgp_peer_names` | `list<string>` | The BGP peers, in spec order |
| `external_gateway_self_link` | `string` | The external VPN gateway (empty for a Google peer) |
| `gateway_self_link` | `string` | The gateway the tunnels leave from |
| `router_name` | `string` | The router the sessions run on |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Every tunnel field except labels is immutable.** Rotate a pre-shared key by adding a tunnel with the new key and removing the old one — never by editing in place.
- **Each tunnel needs its own link-local /30** from 169.254.0.0/16, unique on the router.
- **The MD5 key rides the session**: Google requires each key to be used by exactly one BGP peer.
- **Cost**: each tunnel bills per hour from creation, plus tunnel traffic at internet egress rates.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpHaVpnGateway](/docs/catalog/gcp/gcphavpngateway) — the gateway and router this connection rides
- [GcpVpcPeering](/docs/catalog/gcp/gcpvpcpeering) — shares the routes this connection learns with peered networks
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network on the Google side

## Additional Resources

- [HA VPN topologies](https://cloud.google.com/network-connectivity/docs/vpn/concepts/topologies)
- [Supported IKE ciphers](https://cloud.google.com/network-connectivity/docs/vpn/concepts/supported-ike-ciphers)
- [BGP sessions and BFD on Cloud Router](https://cloud.google.com/network-connectivity/docs/router/concepts/bfd)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
