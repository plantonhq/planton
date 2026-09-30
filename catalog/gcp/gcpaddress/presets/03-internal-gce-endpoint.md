# Internal GCE Endpoint

This preset reserves a regional internal IP address with the `GCE_ENDPOINT` purpose within a specific subnetwork — the address type used for VM instances, alias IP ranges, and similar compute endpoints.

## When to Use

- VM instances that need a static internal IP in a specific subnetwork
- Alias IP ranges attached to VM network interfaces
- Any compute endpoint that requires a reserved private IP within a subnet's CIDR range

## Key Configuration Choices

- **INTERNAL address type** (`addressType: INTERNAL`) — reserves a private IP within your VPC
- **GCE_ENDPOINT purpose** — tells GCP this address is for a compute endpoint (validated: purpose requires INTERNAL)
- **Required subnetwork** (`subnetwork`) — GCE_ENDPOINT addresses must be scoped to a subnetwork; validated at the schema level
- **Required region** (`region`) — must match the subnetwork's region
- **Optional specific `address`** — set to reserve a particular IP within the subnetwork's range; omit to let GCP assign one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `vm-static-internal-ip` (`addressName`) | Name for this address resource | Choose a descriptive name |
| `us-central1` (`region`) | GCP region | Must match the subnetwork's region |
| `app-subnet` (`subnetwork`) | Subnetwork for the reservation (a name or self-link) | `GcpSubnetwork` outputs (`subnetwork_self_link`) or GCP Console |

## Related Presets

- **01-external-nat-ip** — External static IP for Cloud NAT
- **02-internal-lb-vip** — Internal shared load balancer VIP

## Related Components

- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — provides the subnetwork referenced by `subnetwork`
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — parent network for the subnetwork
