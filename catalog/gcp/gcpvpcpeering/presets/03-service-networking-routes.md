# Service Networking Routes

Turns on custom-route export for the peering Google created when a
`GcpServiceNetworkingConnection` reserved a private range — the
`servicenetworking-googleapis-com` peering behind Cloud SQL private IP and
Memorystore — so an on-premises network reached over VPN or Interconnect
learns the route to the database.

## What it configures

- `peerNetwork` left empty — this is the ROUTES-CONFIG form: the peering
  already exists and only its route exchange is managed.
- `peeringName: servicenetworking-googleapis-com` — the name Google gives
  the private-services-access peering.
- `network` — the VPC that holds the private range, by reference.
- `exportCustomRoutes: true` — the VPC's VPN/Interconnect routes are
  offered to Google's service network, which is what lets on-premises
  reach the private instances.

## Adjust before deploying

- **`network.valueFrom.name`** — your VPC's manifest name.
- **`importCustomRoutes`** — leave false; Google's service network has no
  custom routes you want.

## When to choose something else

To create a peering between two of your own networks, use the **Peer Two
VPCs** presets. Destroying this preset changes nothing in GCP; the peering
keeps its current flags.
