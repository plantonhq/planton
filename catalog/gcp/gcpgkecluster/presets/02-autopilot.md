# Autopilot Cluster

This preset creates a private Autopilot cluster: GKE provisions and manages the nodes, bills per running pod, and enforces a hardened security posture out of the box.

## When to Use

- Teams that want Kubernetes without owning node management, capacity planning, or node upgrades
- Workloads with spiky or unpredictable resource demand (you pay for pod requests, not idle nodes)
- Organizations standardizing on GKE's opinionated hardened defaults (shielded nodes, Workload Identity, Dataplane V2 — all always on)

## Key Configuration Choices

- **Autopilot mode** (`enableAutopilot: true`) — no `GcpGkeNodePool` resources attach; node-management fields (node auto-provisioning, the shielded-nodes flag, Calico network policy, the dns-cache/stateful-ha addons) are rejected before deploy by validation rules
- **Private nodes** — Autopilot composes with the same private-cluster topology as Standard clusters
- **Recurring weekend maintenance window** — Autopilot upgrades nodes continuously; the window scopes when disruptive maintenance may run
- **REGULAR release channel** — Autopilot clusters must be on a release channel
- **NET_ADMIN** — if a networking agent or service mesh needs it, set `allowNetAdmin: true` (Autopilot-only field)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-app-vpc` | Your `GcpVpcNetwork` resource name | Your VPC manifest |
| `my-gke-subnet` | Your `GcpSubnetwork` resource name | Your subnetwork manifest |
| `172.16.0.32/28` | Control-plane /28 block | Your network plan; must not overlap any VPC range |
| `203.0.113.0/24` | CIDR allowed to reach the API server | Your office/VPN egress range |

## Related Presets

- **01-private-standard** — when you need control over node pools (machine types, GPUs, spot)
- **03-dev-zonal** — the smallest, cheapest cluster for development

## Related Kinds

- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network the cluster lives in
- [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the subnetwork nodes and pods draw addresses from
