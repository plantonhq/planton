# External Static IP

This preset reserves a global external IPv4 address for use with HTTP(S) load balancers, Cloud CDN, or global forwarding rules. It is the simplest GcpGlobalAddress configuration — just a project, a name, and GCP assigns a public IP automatically.

## When to Use

- Production HTTP(S) load balancers that need a stable public IP for DNS A records
- Cloud CDN frontends or global forwarding rules
- Any workload where you need a static external IP that persists across resource recreation
- SSL certificate provisioning that requires a known IP before the load balancer is created

## Key Configuration Choices

- **EXTERNAL address type** (`addressType: EXTERNAL`) — reserves a public IPv4 address at global scope
- **IPV4** (`ipVersion: IPV4`) — standard IPv4; change to `IPV6` if your load balancer uses IPv6
- **No explicit `address` field** — GCP automatically assigns an available public IP; set `address` only if you need to reserve a specific IP you already own
- **No `network`, `purpose`, or `prefixLength`** — these fields are not applicable for external addresses

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the address will be reserved | GCP Console or `GcpProject` outputs |
| `<your-address-name>` | Name for this address resource (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `prod-lb-ip`) |

## Related Presets

- **02-internal-vpc-peering-range** — Reserve an internal IP range for VPC peering with managed services (Cloud SQL, Redis, AlloyDB)
- **03-private-service-connect** — Reserve an internal address for a Private Service Connect endpoint
