# Internal Load Balancer VIP

This preset reserves a regional internal IP address with the `SHARED_LOADBALANCER_VIP` purpose — the address type used by internal Application Load Balancers when multiple backends share a single VIP.

## When to Use

- Regional internal Application Load Balancers (ILB) that need a stable internal frontend IP
- Shared VIP configurations where multiple backend services reference the same internal address
- Internal traffic routing where the VIP must persist independently of backend changes

## Key Configuration Choices

- **INTERNAL address type** (`addressType: INTERNAL`) — reserves a private IP within your VPC
- **SHARED_LOADBALANCER_VIP purpose** — tells GCP this address is a shared internal LB VIP (validated: purpose requires INTERNAL)
- **Required region** (`region`) — must match the region of the internal load balancer and its backends
- **No `subnetwork`** — SHARED_LOADBALANCER_VIP does not require a subnetwork (unlike GCE_ENDPOINT)
- **No `network_tier`** — internal traffic always uses Premium tier; setting network_tier on INTERNAL addresses is rejected by validation

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `ilb-vip` (`addressName`) | Name for this address resource | Choose a descriptive name |
| `us-central1` (`region`) | GCP region | Must match the ILB region |

## Related Presets

- **01-external-nat-ip** — External static IP for Cloud NAT or regional external LBs
- **03-internal-gce-endpoint** — Internal IP within a subnetwork for a VM or alias IP

## Related Kinds

- [GcpRegionalBackendService](/docs/catalog/gcp/gcpregionalbackendservice) — backend service that the ILB routes to
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — provides the VPC network for internal addresses
