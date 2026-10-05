# GCP HA VPN Connection

Connects a Google Cloud HA VPN gateway (`GcpHaVpnGateway`) to ONE peer — an on-premises site, another cloud, or another Google Cloud VPC — with one to four IPsec tunnels, each carrying a BGP session on the gateway's Cloud Router. Declare one connection per site; add or remove sites without touching the gateway. Two tunnels, one per gateway interface, to two peer addresses is Google's 99.99% shape.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **External VPN gateway** (external peer only) -- the `compute_external_vpn_gateway` holding the device's public addresses
- **VPN tunnels** -- one `compute_vpn_tunnel` per `tunnels[]` entry
- **Router interfaces** -- one `compute_router_interface` per tunnel (the link-local /30)
- **BGP peers** -- one `compute_router_peer` per tunnel, with BFD and MD5 when declared
- **Keys** (only when the spec declares none) -- one generated IKE pre-shared key for every tunnel without its own, and one generated BGP MD5 key for every MD5 session without its own, reported as sensitive outputs

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP HA VPN Gateway and the peer

- **The gateway must exist** -- declare it with `GcpHaVpnGateway` and reference its `gateway_self_link`, `router_name`, and `region` outputs.
- **For an external peer**, know the device's public addresses and its ASN; configure the device toward the gateway's two `interface_*_ip_address` outputs. Declare the device's pre-shared key in `sharedSecret`, or leave it out and configure the device with the generated key from the `shared_secret` output.
- **For a Google peer**, the other VPC declares its own `GcpHaVpnGateway` and a `GcpHaVpnConnection` pointing back at this side's gateway. Exactly one of the two connections generates the keys (it declares none and deploys first); the other references its `shared_secret` (and `md5_authentication_key`) outputs from `sharedSecret` (and `md5AuthenticationKey`). If both sides leave the keys empty, each generates its own and the tunnels never come up.

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
      bgpSession:
        interfaceIpRange: 169.254.10.1/30
        peerAsn: 65001
    - name: hq-tunnel-1
      vpnGatewayInterface: 1
      peerExternalGatewayInterface: 1
      bgpSession:
        interfaceIpRange: 169.254.11.1/30
        peerAsn: 65001
  sharedSecret:
    value: ${secrets-group.hq-vpn.psk}
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
| `tunnels` | `list` | 1-4 tunnels, each with `name` and a `bgpSession` with `peerAsn`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `sharedSecret` | `StringValueOrRef` (sensitive) | generated | The IKE pre-shared key every tunnel uses unless it declares its own: a literal (1-63 printable characters), a managed secret, or a reference to the peer `GcpHaVpnConnection`'s `shared_secret` output. Empty means the module generates one (32 letters and digits). |
| `md5AuthenticationKey` | `StringValueOrRef` (sensitive) | generated | The BGP MD5 key every session that declares a `bgpSession.md5AuthenticationKey` block uses unless the block carries its own key: up to 80 printable characters, or a reference to the peer connection's `md5_authentication_key` output. Empty means the module generates one (24 letters and digits). Enables nothing by itself. |
| `tunnels[].sharedSecret` | `string` (sensitive) | the connection's key | A per-tunnel override, for a device configured with a different key on each tunnel. |
| `tunnels[].vpnGatewayInterface` | `int32` | `0` | Which gateway interface (0/1) the tunnel leaves from. |
| `tunnels[].peerExternalGatewayInterface` | `int32` | — | Which device interface it lands on (external peer only). |
| `tunnels[].ikeVersion` | `int32` | `2` | 1 or 2. |
| `tunnels[].localTrafficSelector` / `remoteTrafficSelector` | `list` | — | IPv4 CIDRs for route-based tunnels without BGP. |
| `tunnels[].cipherSuite` | `message` | Google defaults | Phase-1 and phase-2 algorithm restrictions. |
| `tunnels[].labels` | `map` | — | The one mutable tunnel field. |
| `tunnels[].description` | `string` | — | Immutable. |
| `tunnels[].bgpSession.*` | — | — | `name`, `interfaceIpRange` (link-local /30), `ipVersion`, `peerIpAddress`, `advertisedRoutePriority`, `advertiseMode` / `advertisedGroups` / `advertisedIpRanges`, `enable`, `enableIpv4`, `enableIpv6`, the four next-hop addresses, `customLearnedIpRanges`, `customLearnedRoutePriority`, `bfd`, `md5AuthenticationKey` (declaring the block enables MD5; its optional sensitive `key` overrides the connection's), `importPolicies`, `exportPolicies`. |
| `resourceManagerTags` | `map` | — | Create-time tags on every tunnel and the external gateway. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON` for every resource. |

### Validation Rules

- **`peer`**: exactly one arm. External interfaces match `redundancyType` (1 / 2 / 4), unique ids, one address each.
- **`tunnels`**: distinct names; distinct `interfaceIpRange`s; every tunnel names a `peerExternalGatewayInterface` with an external peer and none with a Google peer; the named interface exists.
- **`bgpSession`**: `peerAsn` ≥ 1; `interfaceIpRange` is `169.254.x.y/30`; BFD intervals 1000-30000, multiplier 5-16; priorities 0-65535 / 0-65335.
- **Keys**: a literal `sharedSecret` is 1-63 printable ASCII characters; a literal `md5AuthenticationKey` is 1-80.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `tunnel_self_links` | `list<string>` | The tunnels' self links, in spec order |
| `tunnel_names` | `list<string>` | The tunnels' names, in spec order |
| `router_interface_names` | `list<string>` | The router interfaces, in spec order |
| `bgp_peer_names` | `list<string>` | The BGP peers, in spec order |
| `external_gateway_self_link` | `string` | The external VPN gateway (empty for a Google peer) |
| `gateway_self_link` | `string` | The gateway the tunnels leave from |
| `router_name` | `string` | The router the sessions run on |
| `shared_secret` | `string` (sensitive) | The generated pre-shared key -- set only when the module generated it; a declared key is never echoed back |
| `md5_authentication_key` | `string` (sensitive) | The generated BGP MD5 key -- set only when the module generated it |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Keys are never required.** A tunnel uses its own key, else the connection's `sharedSecret`, else the one key the module generates; MD5 sessions follow the same rule with `md5AuthenticationKey`. In a Google-to-Google pair exactly one side generates and the other references its outputs.
- **Every tunnel field except labels is immutable.** Rotate a pre-shared key by adding a tunnel with the new key (its own `sharedSecret`) and removing the old one — never by editing in place.
- **Each tunnel needs its own link-local /30** from 169.254.0.0/16, unique on the router.
- **The MD5 key rides the session**: Google requires each key-table entry to be used by exactly one BGP peer, so every MD5 session gets its own named entry; the key material itself may be shared.
- **Cost**: each tunnel bills per hour from creation, plus tunnel traffic at internet egress rates.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

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
