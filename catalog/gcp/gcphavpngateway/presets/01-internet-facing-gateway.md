# Internet Facing Gateway

The everyday HA VPN gateway: two public IPv4 interfaces on the hub VPC in
one region, a Cloud Router speaking a 16-bit private ASN and advertising
every subnet, protected against destroy because every site's device is
configured against its addresses.

## What it configures

- `region` and `network` — where the gateway lives; every connection's
  tunnels are in this region. Both immutable.
- `router.bgp.asn: 64514` — the ASN Google speaks as toward every peer;
  private and fixed for the router's life.
- Advertisement left at `DEFAULT` — every VPC subnet is advertised to every
  peer; a connection's session can narrow it per peer.
- `deletionPolicy: PREVENT` — a recreated gateway has new public IPs.

## Adjust before deploying

- **`network.valueFrom.name`** and **`region`** — your VPC and region.
- **`asn`** — any private ASN (64512-65534, or 4200000000 and up) that
  differs from every peer's.

## When to choose something else

For IPv6 traffic inside the tunnels or a curated advertisement, start from
**Dual Stack Gateway**; for HA VPN over Cloud Interconnect, from
**Interconnect Backed Gateway**.
