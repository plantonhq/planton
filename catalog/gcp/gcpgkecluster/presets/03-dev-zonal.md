# Development Zonal Cluster

This preset creates the smallest, cheapest GKE control plane: a zonal Standard cluster with GKE-managed IP ranges, public nodes, the RAPID channel, and deletion protection off.

## When to Use

- Personal or team development sandboxes that are created and destroyed freely
- Trying new Kubernetes versions early (RAPID channel)
- Cost-sensitive non-critical environments where a brief control-plane outage during upgrades is acceptable

## Key Configuration Choices

- **Zonal location** (`us-central1-a`) — a single control-plane instance; upgrades briefly interrupt the API server
- **GKE-managed IP ranges** — no `ipAllocation` block, so GKE creates and manages the pod/service secondary ranges itself
- **Public nodes** — no `privateCluster` block; nodes get public IPs and need no Cloud NAT (simplest egress for a sandbox)
- **RAPID channel** — new Kubernetes minors weeks after upstream release
- **Deletion protection off** — sandboxes are rebuilt freely; production keeps the default (`true`)

Add a small `GcpGkeNodePool` (e.g. two `e2-medium` nodes) to run workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-app-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `my-dev-subnet` | Your `GcpSubnetwork` resource name | Your subnetwork manifest |

## Related Presets

- **01-private-standard** — the production shape (regional, private nodes, planned ranges)
- **02-autopilot** — no node management at all, per-pod billing

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the cluster lives in
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the subnetwork nodes attach to
- [GcpGkeNodePool](/docs/catalog/gcp/gcpgkenodepool) — compute for the sandbox
