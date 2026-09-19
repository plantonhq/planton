# Four Tunnel To Onprem

A data center with four public addresses (two devices, two uplinks each):
two tunnels from each gateway interface, four BGP sessions, with route
priorities that make the peer prefer one tunnel per interface and hold the
other as standby.

## What it configures

- `gateway`, `router`, `region` — literals, for a gateway created outside
  this chart.
- `peer.externalGateway` — `FOUR_IPS_REDUNDANCY` with interfaces 0-3.
- `tunnels` — tunnels 0 and 1 from gateway interface 0 to device interfaces
  0 and 1; tunnels 2 and 3 from gateway interface 1 to device interfaces 2
  and 3; each on its own /30.
- `advertisedRoutePriority` 100 / 200 — the MED the peer sees; the lower
  value wins, so the device prefers tunnels 0 and 2 and fails over to 1 and
  3. Equal values would split traffic across all four.

## Adjust before deploying

- **The four `ipAddress`es**, **`peerAsn`**, and the four secrets — from the
  data-center team.
- **The gateway literals** — your gateway's self link, router name, and
  region (or switch to `valueFrom` references).

## When to choose something else

Two addresses is the common case: **Two Tunnel To Onprem** earns the same
99.99% SLA with half the tunnel-hours.
