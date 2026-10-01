# Spoke to Hub

This preset is the **spoke-side half** of a hub-and-spoke peering pair.
It is written on the spoke network and points back at the hub. Deploy it
alongside `01-hub-to-spoke` (or an equivalent hub-side resource) for
every spoke -- two resources per pair, never one.

`use_remote_gateways` is the spoke-side complement of
`allow_gateway_transit` on the hub. Set it only when this spoke should
send on-premises-bound traffic through the hub's VPN or ExpressRoute
gateway. Azure constraints to plan for:

- Only **one** peering per network may set `use_remote_gateways`
- The spoke network must have **no gateway of its own**
- Gateway transit **cannot** be combined with global (cross-region) peering

For simple L3 peering without gateway sharing, omit both gateway flags on
both directions. `allow_virtual_network_access` defaults to true on both
sides, which is sufficient for direct VM-to-VM reachability across the
peerings.

## When to Use

- Completing a hub-and-spoke pair after deploying the hub-side peering
- Spokes that reach on-premises exclusively through the hub gateway
- Any spoke VNet joining an existing hub topology

## Key Configuration Choices

- **Name after the far side** (`spoke1-to-hub`) -- the reciprocal of the
  hub-side name makes the pair easy to audit
- **`useRemoteGateways: true`** -- only with hub `allowGatewayTransit: true`
  and no spoke gateway; remove for peering-only connectivity
- **Leave `allowForwardedTraffic` at default (`false`)** on the spoke-to-hub
  direction unless you have a specific relayed-traffic requirement in this
  direction (uncommon; the hub-to-spoke direction carries the inspection
  admission flag)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<spoke-vnet-resource-name>` | Planton metadata name of the spoke `AzureVirtualNetwork` | Your spoke network stack |
| `<hub-vnet-resource-name>` | Planton metadata name of the hub `AzureVirtualNetwork` | Your hub network stack |
| `spoke1-to-hub` (`name`) | Peering name within the spoke (reciprocal of the hub's) | Your naming convention |

## Pair With

Deploy `01-hub-to-spoke` on the hub with `virtualNetworkId` and
`remoteVirtualNetworkId` reversed. Match gateway-transit flags across the
pair or remove them on both sides.
