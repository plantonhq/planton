# Hybrid Connectivity Group

## Use Case

Load-balance to servers that are not in Google Cloud -- an on-premises data center or another cloud reached over Cloud VPN or Cloud Interconnect. A `NON_GCP_PRIVATE_IP_PORT` group lists those private addresses so a Google Cloud load balancer can front them, during a migration or permanently.

## When to Use

- Migrating a service to Google Cloud gradually, splitting traffic between cloud and on-premises backends
- Fronting on-premises services with Cloud Armor, Cloud CDN, or Google's global anycast address
- Any load-balancer backend whose addresses are reachable only over VPN or Interconnect

## What This Creates

- A zonal group `onprem-api-neg` in `us-central1-b` on `hybrid-vpc`
- Type `NON_GCP_PRIVATE_IP_PORT`, default port 443
- Three on-premises endpoints, one on port 8443

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `zone` | `us-central1-b` | A zone in the region whose VPN or Interconnect attachment reaches the addresses. Immutable. |
| `network` | `hybrid-vpc` | The VPC that holds the VPN or Interconnect routes to the addresses. Immutable. |
| `endpoints[].ipAddress` | three RFC 1918 addresses | Your on-premises servers; every hybrid endpoint names an address, never an instance. |
| `defaultPort` / `port` | 443 / 8443 | The ports the servers listen on. |

Google accepts hybrid groups only on backend services with the `EXTERNAL`, `EXTERNAL_MANAGED`, `INTERNAL_MANAGED`, or `INTERNAL_SELF_MANAGED` scheme in RATE or CONNECTION balancing mode; health checks reach the addresses over the same VPN or Interconnect, so allow the health-check ranges on-premises.
