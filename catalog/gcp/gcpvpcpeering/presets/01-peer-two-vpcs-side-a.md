# Peer Two VPCs Side A

The hub's half of a hub-spoke peering: the hub exports its custom routes
(the VPN and Interconnect routes its Cloud Router learns) so the spoke
can reach on-premises through the hub. Pair it with **Peer Two VPCs Side
B** on the spoke.

## What it configures

- `network` — the hub `GcpVpcNetwork`, by reference.
- `peerNetwork` — the spoke `GcpVpcNetwork`, by reference; setting it makes
  this the CREATE form.
- `exportCustomRoutes: true` — the hub's static and dynamic routes are
  offered to the spoke (which must import them on its side).
- `deletionPolicy: PREVENT` — production traffic depends on the pair.

## Adjust before deploying

- **Both `valueFrom.name`s** — your hub and spoke manifest names, or the
  networks' self links as literals.
- **Subnet ranges** of the two networks must not overlap.

## When to choose something else

Use **Service Networking Routes** to manage the routes of a peering Google
created for Cloud SQL or Memorystore instead of creating one.
