# Peer Two VPCs Side B

The spoke's half of a hub-spoke peering: the spoke imports the hub's
custom routes so its workloads reach on-premises through the hub's VPN.
The pair goes `ACTIVE` once both sides exist.

## What it configures

- `network` — the spoke `GcpVpcNetwork`, by reference.
- `peerNetwork` — the hub `GcpVpcNetwork`, by reference.
- `importCustomRoutes: true` — accepts the routes the hub exports (side A
  sets `exportCustomRoutes`); without both flags only subnet routes cross.
- `deletionPolicy: PREVENT` — removing either side breaks the pair.

## Adjust before deploying

- **Both `valueFrom.name`s** — your spoke and hub manifest names.
- **Add `exportCustomRoutes: true` here too** if the hub should learn routes
  the spoke carries (rare for a spoke).

## When to choose something else

A spoke that must never let the hub's route changes alter its traffic
unilaterally adds `updateStrategy: CONSENSUS` on both sides.
