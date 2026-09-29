# Hub to Spoke

This preset is the **hub-side half** of a hub-and-spoke peering pair.
It is written on the hub network and points at one spoke. Connectivity
only works once the reciprocal `spoke-to-hub` preset (or equivalent) is
deployed on that spoke with local and remote swapped.

`allow_forwarded_traffic` admits traffic the spoke did not originate but
that the hub relayed -- essential when a firewall or VPN gateway in the hub
inspects spoke-originated traffic before it returns. Pair it with user-
defined routes on the spoke that send `0.0.0.0/0` (or specific prefixes)
to the hub appliance.

`allow_gateway_transit` lets spokes use the hub's VPN or ExpressRoute
gateway for on-premises reachability. It only takes effect when the
matching spoke resource sets `use_remote_gateways: true` and the spoke
network has no gateway of its own.

## When to Use

- Connecting a new spoke VNet into an existing hub
- Hub-and-spoke designs with centralized firewall or VPN/ExpressRoute
- Inspection topologies where forwarded traffic must be admitted on the
  hub-to-spoke direction

## Key Configuration Choices

- **Name after the far side** (`hub-to-spoke1`) so the hub's peering list
  reads as a topology map
- **`allowForwardedTraffic: true`** -- required when spokes send traffic
  through hub NVAs or gateways and expect return traffic admitted
- **`allowGatewayTransit: true`** -- only when spokes should ride the hub
  gateway; omit both gateway flags for simple L3 peering with no transit

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<hub-vnet-resource-name>` | Planton metadata name of the hub `AzureVirtualNetwork` | Your hub network stack |
| `<spoke-vnet-resource-name>` | Planton metadata name of the spoke `AzureVirtualNetwork` | Your spoke network stack |
| `hub-to-spoke1` (`name`) | Peering name within the hub (name it after the far side) | Your naming convention |

## Pair With

Deploy `02-spoke-to-hub` on the same spoke with `virtualNetworkId` and
`remoteVirtualNetworkId` reversed. If using gateway transit, set
`useRemoteGateways: true` on the spoke-side resource only.
