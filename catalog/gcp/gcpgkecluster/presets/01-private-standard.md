# Private Production Cluster — Standard

This preset creates a regional Standard cluster with private nodes, Dataplane V2, planned secondary ranges, a control-plane access allowlist, a daily maintenance window, and the Security Posture dashboard — the GCP-recommended production shape.

## When to Use

- Production workloads that need control over node pools (machine types, GPUs, spot mixes) via `GcpGkeNodePool` resources
- Clusters where nodes must not have public IP addresses
- Organizations that plan pod/service address space explicitly on the subnetwork

## Key Configuration Choices

- **Regional location** (`us-central1`) — control-plane replicas across three zones; zonal clusters save money but take a brief API outage during upgrades
- **Private nodes** (`privateCluster.enablePrivateNodes`) — no public node IPs; compose a `GcpRouterNat` on the network for image pulls and other egress
- **Peering-based control plane** (`masterIpv4CidrBlock`) — a dedicated /28 that must not overlap any VPC range
- **Dataplane V2** (`datapathProvider: ADVANCED_DATAPATH`) — native NetworkPolicy enforcement without the Calico addon, plus dataplane observability
- **Named secondary ranges** (`ipAllocation`) — pod/service space comes from planned ranges on the `GcpSubnetwork`, not ad-hoc GKE carving
- **Master authorized networks** — the API endpoint only accepts your listed CIDRs
- **Cost allocation on** (`enableCostManagement`) — per-namespace cost attribution in the billing export
- **Deletion protection on** (the default) — a destroy fails until it is explicitly set to `false`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-app-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `my-gke-subnet` | Your `GcpSubnetwork` resource name (with `pods`/`services` secondary ranges) | Your subnetwork manifest |
| `172.16.0.16/28` | Control-plane /28 block | Your network plan; must not overlap any VPC range |
| `203.0.113.0/24` | CIDR allowed to reach the API server | Your office/VPN egress range |

## Related Presets

- **02-autopilot** — let GKE manage nodes entirely and bill per pod
- **03-dev-zonal** — the smallest, cheapest cluster for development

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the cluster lives in
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — carries the pod/service secondary ranges
- [GcpGkeNodePool](/docs/catalog/gcp/gcpgkenodepool) — compute for Standard clusters
- [GcpRouterNat](/docs/catalog/gcp/gcprouternat) — outbound internet for private nodes
