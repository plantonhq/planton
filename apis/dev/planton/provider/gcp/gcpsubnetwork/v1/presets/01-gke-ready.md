# GKE-Ready Subnet

This preset creates a subnet with secondary IP ranges for GKE pod and service CIDRs, plus Private Google Access enabled. This is the standard subnet configuration required before creating a VPC-native GKE cluster.

## When to Use

- Subnets that will host GKE clusters using VPC-native networking (alias IPs)
- Any GKE deployment requiring dedicated pod and service IP ranges
- Production environments where Private Google Access is needed for private nodes

## Key Configuration Choices

- **`/20` primary range** (`10.0.0.0/20`) -- 4,096 node IPs, sufficient for most GKE clusters
- **`/14` pods secondary range** (`10.4.0.0/14`) -- 262,144 pod IPs, supports large clusters with many pods per node
- **`/20` services secondary range** (`10.8.0.0/20`) -- 4,096 service IPs, sufficient for most Kubernetes services
- **Private Google Access** (`privateIpGoogleAccess: true`) -- nodes without external IPs can reach Google APIs (gcr.io, storage, logging)
- **Range names** -- `pods` and `services` are referenced by `GcpGkeCluster` presets

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<vpc-network-self-link>` | Self-link of the parent VPC network | `GcpVpc` status outputs |
| `<your-subnet-name>` | Name for this subnet (1-63 chars, lowercase) | Choose a descriptive name (e.g., `gke-us-central1`) |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |

## Related Presets

- **02-general-purpose** -- Use for subnets hosting Compute Engine VMs or Cloud Run (no secondary ranges needed)
