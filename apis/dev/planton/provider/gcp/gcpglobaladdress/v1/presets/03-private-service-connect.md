# Private Service Connect Endpoint Address

This preset reserves a global internal IP address for a Private Service Connect (PSC) endpoint. PSC allows private, secure connectivity to Google APIs, Google managed services, or third-party services published via service attachments — all without traffic leaving the VPC. The reserved address becomes the consumer-side IP that routes to the target service.

## When to Use

- Accessing Google APIs (e.g., `storage.googleapis.com`, `bigquery.googleapis.com`) over a private IP instead of public endpoints
- Connecting to third-party SaaS or partner services published via PSC service attachments
- Zero-trust or compliance environments that require all traffic to stay within the VPC (no public internet egress to Google APIs)
- Replacing Private Google Access with a more granular, per-service approach

## Key Configuration Choices

- **INTERNAL address type** (`addressType: INTERNAL`) — reserves a private IP within the VPC, not a public address
- **PRIVATE_SERVICE_CONNECT purpose** (`purpose: PRIVATE_SERVICE_CONNECT`) — marks this address for use with a PSC endpoint
- **No `prefixLength`** — PSC endpoints use a single IP address, not a CIDR range
- **Network reference required** — the VPC network from which the private IP is allocated
- **No explicit `address` field** — GCP assigns an available RFC1918 IP; set `address` only to pin a specific IP

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the address will be reserved | GCP Console or `GcpProject` outputs |
| `<your-address-name>` | Name for this address resource (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `psc-endpoint-ip`) |
| `<vpc-network-name-or-self-link>` | VPC network name (e.g., `prod-vpc`) or full self-link | `GcpVpc` status outputs |

## Related Presets

- **01-external-static-ip** — Reserve a public IP for HTTP(S) load balancers
- **02-internal-vpc-peering-range** — Reserve an internal IP range for VPC peering with managed services (Cloud SQL, Redis, AlloyDB)
