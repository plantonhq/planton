# Global Internet Origin

## Use Case

Front an origin that lives outside Google Cloud -- a SaaS API, a server on another cloud, a legacy data center reachable over the internet -- with a global external Application Load Balancer, so it gains Google's anycast address, Cloud Armor, and Cloud CDN. A global `INTERNET_FQDN_PORT` group names the origin by hostname; Google resolves it at connection time.

## When to Use

- Putting Cloud Armor or Cloud CDN in front of an external origin
- Serving one hostname from Google Cloud while the backend still lives elsewhere
- A global load balancer whose backend is reachable over the public internet (for private on-premises addresses, use the hybrid preset)

## What This Creates

- A global internet group `origin-neg` (no zone, no network)
- Type `INTERNET_FQDN_PORT`, default port 443
- One endpoint, `origin.example.com:443`, as its own endpoint resource

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `networkEndpointType` | `INTERNET_FQDN_PORT` | `INTERNET_IP_PORT` to name the origin by IP address (`endpoints[].ipAddress`) instead of hostname. Immutable. |
| `endpoints[].fqdn` | `origin.example.com` | Your origin's hostname; every FQDN endpoint names one. |
| `defaultPort` / `port` | 443 | The port the origin listens on. |

The consuming global `GcpBackendService` names this group's `self_link` output in `backends[].group`; the load balancer reaches the origin over the internet, so the origin must accept traffic from Google's proxy ranges.
