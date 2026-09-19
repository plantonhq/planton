# Dual Stack Gateway

A gateway whose tunnels carry IPv4 and IPv6, with a 32-bit private ASN and
a curated advertisement: every subnet plus one extra range the hub routes
for.

## What it configures

- `stackType: IPV4_IPV6` — dual-stack tunnels; the VPC must have an
  internal IPv6 range and each connection's session enables IPv6.
- `router.bgp.advertiseMode: CUSTOM` with `advertisedGroups: [ALL_SUBNETS]`
  and one `advertisedIpRanges` entry — the subnets plus a range the hub
  reaches through another path (a peered network, a second VPN).
- `keepaliveInterval: 30` — a slower BGP keepalive for a link with jitter;
  hold time is three times this.
- `asn: 4200000001` — a 32-bit private ASN, for estates that have used up
  the 16-bit range.

## Adjust before deploying

- **`network.valueFrom.name`**, **`region`**, **`asn`** — yours.
- **`advertisedIpRanges`** — the ranges your peers must learn beyond the
  VPC's own subnets.

## When to choose something else

Most estates start from **Internet Facing Gateway** and add IPv6 later;
the stack type is immutable, so decide before the peers are configured.
