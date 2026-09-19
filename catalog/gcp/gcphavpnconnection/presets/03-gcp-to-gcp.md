# Gcp To Gcp

The hub's half of an HA VPN between two Google Cloud VPCs: two tunnels to
the spoke's gateway, interfaces paired by Google. The spoke declares the
mirror connection pointing back at `hub-vpn`, with the same two secrets
and `peerAsn: 64514` (the hub's ASN).

## What it configures

- `gateway`, `router`, `region` — references to the hub `GcpHaVpnGateway`.
- `peer.gcpGateway` — a reference to the spoke `GcpHaVpnGateway`'s self
  link; no external gateway is created.
- `tunnels` — one from each gateway interface, no
  `peerExternalGatewayInterface` (Google pairs interface 0 to 0 and 1 to
  1); each with its own /30 and a session to the spoke router's ASN.

## Adjust before deploying

- **The two gateway `valueFrom.name`s** — your hub and spoke gateways.
- **`peerAsn`** — the spoke gateway's `router_asn` output.
- **The two secrets** — the same values on both sides' connections.
- **The spoke's mirror** — `169.254.30.2/30` and `169.254.31.2/30` as its
  interface ranges so both ends of each /30 agree.

## When to choose something else

Two VPCs that can peer should use `GcpVpcPeering` instead: same private
reachability, no tunnel-hours, intra-VPC egress rates. VPN between VPCs is
for networks that cannot peer -- overlapping ranges, transitive routing
needs, or an encryption requirement on the wire.
