# GcpVpcPeering Guide

The judgment this guide protects: a peering is two half-entries, and
this block is one of them. Declare each side where its network lives,
and decide route exchange per side, on purpose.

## Two resources for one peering

Google models a peering as an entry in each network. When you own both
networks, declare two `GcpVpcPeering` blocks -- one with
`network: hub, peerNetwork: spoke`, one with the reverse -- and the pair
goes `ACTIVE` when both exist. When the other network belongs to
someone else, declare your side and hand them your network's self link;
your `state` output reads `INACTIVE` with a `state_details` that says
the peer has no matching entry until they create theirs. The provider
serializes the two sides when one chart creates both, so declaring them
together is safe.

## Create versus routes-config

Set `peerNetwork` and this block CREATES the peering. Leave it empty and
the block manages only the route exchange of a peering that already
exists under `peeringName` -- the shape for
`servicenetworking-googleapis-com`, the peering Google creates when a
`GcpServiceNetworkingConnection` reserves a private range for Cloud SQL
or Memorystore. You cannot create that peering yourself, but you can
turn `exportCustomRoutes` on so an on-premises network reached over VPN
learns the route to the database. Destroying the routes-config form
does nothing in GCP; the peering keeps its last flags.

## Which routes cross

Subnet routes always cross. Custom routes -- static routes and the
dynamic routes a Cloud Router learns over VPN or Interconnect -- cross
only when the side that has them sets `exportCustomRoutes` AND the side
that wants them sets `importCustomRoutes`. Two flags, two sides, both
deliberate. Peering is not transitive: routes learned across one
peering are never re-exported across another, so a hub-and-spoke needs
one peering per spoke and the hub's VPN routes reach a spoke only
through that spoke's own peering.

## Nearly everything is immutable

The name, both networks, and the two public-IP subnet-route flags
recreate the peering, which drops traffic until the new pair is active.
Change the custom-route flags, `stackType`, and `updateStrategy` freely;
plan anything else as a new peering.

## Consensus when two teams own the sides

`updateStrategy: CONSENSUS` makes a changed route-exchange setting take
effect only once both sides carry the same value. Use it when the other
side is another team's chart, so nobody's one-sided flip can change what
traffic flows.

## Limits Google enforces

Subnet ranges must not overlap between the two networks (Google rejects
the peering). A network can hold 25 active peerings by default (a quota
increase raises it). Dual-stack peering (`stackType: IPV4_IPV6`) needs
an internal IPv6 range on both `GcpVpcNetwork`s.
