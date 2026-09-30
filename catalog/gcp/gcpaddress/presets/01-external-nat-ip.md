# External NAT IP

This preset reserves a regional external IPv4 address for use with Cloud NAT, regional load balancers, or VM instances. It is the simplest GcpAddress configuration — a project, a name, a region, and GCP assigns a public IP automatically.

## When to Use

- Cloud NAT gateways that need a stable public IP for outbound traffic
- Regional external load balancers that need a fixed frontend IP
- VM instances that require a static external IP in a specific region
- Any workload where you need a regional static external IP that persists across resource recreation

## Key Configuration Choices

- **EXTERNAL address type** (`addressType: EXTERNAL`) — reserves a public IPv4 address at regional scope
- **Required region** (`region`) — regional addresses are scoped to a single GCP region (unlike GcpGlobalAddress)
- **PREMIUM network tier** (`networkTier: PREMIUM`) — the default for external addresses; use STANDARD only when cost optimization outweighs global routing performance
- **No explicit `address` field** — GCP automatically assigns an available public IP; set `address` only if you need to reserve a specific IP you already own
- **No `network`, `subnetwork`, or `purpose`** — these fields are not applicable for external addresses

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID where the address will be reserved | GCP Console or `GcpProject` outputs |
| `nat-external-ip` (`addressName`) | Name for this address resource (1-63 chars, lowercase, hyphens) | Choose a descriptive name |
| `us-central1` (`region`) | GCP region for the reservation | Match the region of the NAT gateway, LB, or VM |

## Related Presets

- **02-internal-lb-vip** — Reserve an internal shared load balancer VIP
- **03-internal-gce-endpoint** — Reserve an internal IP within a subnetwork for a GCE endpoint

## Related Components

- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — for global-scope static IPs (HTTP(S) load balancers, CDN)
- [GcpNatGateway](/docs/catalog/gcp/gcpnatgateway) — consumes this address for Cloud NAT
