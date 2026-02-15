# Internal VPC Peering Range

This preset reserves a `/20` internal IP range for VPC peering with Google managed services. Cloud SQL, Memorystore (Redis), AlloyDB, and Filestore all use this mechanism to assign private IPs from your VPC, keeping traffic off the public internet. The reserved range is consumed by a `google_service_networking_connection` (typically created by the GcpVpc component's Private Services Access feature or manually).

## When to Use

- Any project that runs Cloud SQL, AlloyDB, Memorystore, or Filestore with private IP connectivity
- When enabling Private Services Access on a VPC for the first time
- Multi-service environments that need a dedicated CIDR block for Google-managed service instances
- Compliance environments that prohibit public IPs on database or cache instances

## Key Configuration Choices

- **INTERNAL address type** (`addressType: INTERNAL`) — reserves a private IP range inside the VPC, not a public address
- **VPC_PEERING purpose** (`purpose: VPC_PEERING`) — tells GCP this range is for the service networking peering connection
- **`/20` prefix length** (`prefixLength: 20`) — reserves 4,096 IPs; sufficient for dozens of Cloud SQL and Redis instances. Use a smaller prefix (e.g., `/16`) for large-scale deployments or a larger prefix (e.g., `/24`) when IP space is scarce
- **Network reference required** — the VPC network whose IP space this range is allocated from; the network cannot be deleted while this range exists
- **No explicit `address` field** — GCP picks an available RFC1918 range; set `address` only to pin the start IP

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the range will be reserved | GCP Console or `GcpProject` outputs |
| `<your-address-name>` | Name for this address resource (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `managed-services-range`) |
| `<vpc-network-name-or-self-link>` | VPC network name (e.g., `prod-vpc`) or full self-link | `GcpVpc` status outputs |

## Related Presets

- **01-external-static-ip** — Reserve a public IP for HTTP(S) load balancers
- **03-private-service-connect** — Reserve an internal address for a Private Service Connect endpoint
