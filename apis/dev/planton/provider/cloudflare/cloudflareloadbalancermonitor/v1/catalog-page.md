# Cloudflare Load Balancer Monitor

Define a reusable health check that probes load-balancer origins and drives
failover decisions.

## What Gets Created

- A `cloudflare_load_balancer_monitor` (account-scoped) of the chosen protocol
  (HTTP/HTTPS/TCP/UDP-ICMP/ICMP-PING/SMTP).

## Prerequisites

- A Cloudflare account ID.
- The **Load Balancing add-on** enabled on the account (paid add-on) — otherwise the
  Load Balancing API returns `403`.
- An API token with **Account → Load Balancing: Monitors and Pools → Edit** (monitors
  are account-scoped).

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.

**Optional**

- `type`, `description`, `path`, `expectedCodes`, `expectedBody`, `method`,
  `headers`, `port`, `interval`, `timeout`, `retries`, `consecutiveUp`,
  `consecutiveDown`, `followRedirects`, `allowInsecure`, `probeZone`.

## Stack Outputs

| Output | Description |
|---|---|
| `monitor_id` | The monitor ID |
| `monitor_type` | The health-check protocol |

## Related Components

- CloudflareLoadBalancerPool
- CloudflareLoadBalancer
