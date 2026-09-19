# DigitalOcean Kubernetes Cluster

Managed Kubernetes on DigitalOcean: one Planton component models the full `digitalocean_kubernetes_cluster` resource — version and region placement, the inline default node pool (labels, taints, tags, autoscaling, GPU partitioning), a highly available control plane, surge and automatic upgrades, maintenance policy, control-plane firewall, pod/service/worker subnets, isolated workers, SSO, cluster-autoscaler tuning, container-registry integration, kubeconfig expiry, destroy-time cleanup, and every managed addon toggle.

## What this component models

The spec maps one-to-one onto DigitalOcean's managed Kubernetes cluster:

| Spec field | What it controls |
|---|---|
| `clusterName` | The cluster's name in DigitalOcean |
| `region` | Data-center region; create-only |
| `kubernetesVersion` | The creation version pin, as a minor prefix (`"1.35"`, preferred) or a full slug DigitalOcean offers today; patch upgrades ride `autoUpgrade` — see "Behavior worth knowing" |
| `vpc` | Required VPC placement — a literal UUID or a reference to a `DigitalOceanVpc`; create-only |
| `highlyAvailable` | Multi-replica control plane (extra cost); one-way — cannot be turned off |
| `autoUpgrade` | Automatic patch upgrades inside the maintenance window |
| `surgeUpgrade` | Surge nodes during upgrades; unset defers to DigitalOcean's default (on) |
| `maintenancePolicy` | Weekly window (`day`, `startTime` UTC) for maintenance and auto-upgrades |
| `controlPlaneFirewall` | Restricts API server access (`enabled` + allowed IPs/CIDRs) |
| `clusterSubnet` / `serviceSubnet` | Pod and service CIDR blocks; create-only |
| `workerSubnetUuid` | DigitalOcean-managed worker subnet placement; create-only |
| `isolatedWorkers` | Blocks worker nodes from public internet access; create-only |
| `registryIntegration` | Cluster-wide pulls from the account's DigitalOcean Container Registry |
| `kubeconfigExpireSeconds` | Validity of fetched kubeconfig credentials (0 = 7-day default) |
| `destroyAllAssociatedResources` | Destroy-time: also deletes the cluster's load balancers, volumes, and snapshots |
| `clusterAutoscalerConfiguration` | Scale-down threshold, unneeded time, expander strategies |
| `sso` | OpenID Connect single sign-on for the Kubernetes API |
| `routingAgent`, `corednsAutoscaler`, GPU/RDMA addon toggles | Managed addons; unset defers to DigitalOcean's default per addon |
| `tags` | Your tags, applied alongside the standard Planton labels |
| `defaultNodePool` | The inline pool: size, a fixed count OR autoscaling bounds (never both), node labels, taints, pool tags, GPU partition mode |

Additional node pools beyond the inline default are separate `DigitalOceanKubernetesNodePool` resources with their own lifecycles.

## Quick start

A production cluster with an HA control plane, a firewall in front of the API server, and an autoscaling pool:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanKubernetesCluster
metadata:
  name: app-cluster
  org: acme-corp
  env: prod
spec:
  clusterName: app-cluster
  region: nyc3
  kubernetesVersion: "1.35"
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  highlyAvailable: true
  autoUpgrade: true
  maintenancePolicy:
    day: sunday
    startTime: "03:00"
  controlPlaneFirewall:
    enabled: true
    allowedAddresses:
      - "203.0.113.0/24"
  defaultNodePool:
    size: s-4vcpu-8gb
    autoScale: true
    minNodes: 2
    maxNodes: 5
```

```shell
planton apply -f app-cluster.yaml
```

## Outputs

Both provisioners export the identical output set:

| Output | Description |
|---|---|
| `cluster_id` | The cluster's UUID |
| `kubeconfig` | Raw kubeconfig YAML (not base64) — sensitive; write it to a file and point `KUBECONFIG` at it |
| `api_server_endpoint` | The Kubernetes API server URL |
| `urn` | `do:kubernetes:<cluster_id>`, for project attachment |
| `ipv4_address` | The control plane's public IPv4 when DigitalOcean reports one — empty on clusters created today (single-replica included); reach the API server through `api_server_endpoint` |
| `default_node_pool_id` | The inline default pool's UUID |
| `cluster_subnet` / `service_subnet` | The pod and service CIDR blocks in effect |

## Behavior worth knowing

- **`kubernetesVersion` is the creation pin.** Both provisioners deliberately ignore later drift on it: DigitalOcean's auto-upgrade moves the live version forward, and the provider destroys and recreates the whole cluster when the configured version is lower than the live one. Patch upgrades ride `autoUpgrade`; this component does not drive in-place upgrades through a spec edit.
- **The default pool's `size` replaces the entire cluster.** So does `gpuPartitionMode`. Size the inline pool deliberately and grow with separate `DigitalOceanKubernetesNodePool` resources.
- **HA is one-way.** Once `highlyAvailable` is true, it cannot be turned back off.
- **`destroyAllAssociatedResources` is dangerous.** On destroy it also deletes the load balancers, volumes, and volume snapshots the cluster created. It never affects the running cluster.
- **`surgeUpgrade` unset means ON** — DigitalOcean's default. Set it to `false` explicitly to disable surge upgrades.
- **Every spec field deploys on both provisioners** — `sso`, `isolatedWorkers`, `workerSubnetUuid`, `gpuPartitionMode`, and all nine addon toggles included. A few carry prerequisites DigitalOcean enforces: `isolatedWorkers` needs a NAT gateway attached to the cluster's VPC, `workerSubnetUuid` needs a subnet inside that VPC, the GPU addons and `gpuPartitionMode` only mean anything on GPU node sizes, and `p2pOciRegistryPlugin` needs `kubernetesVersion` 1.36.0-do.2 or later (an older version fails the whole create with a validation 422 and creates nothing).
- **`corednsAutoscaler` unset follows the version** — DigitalOcean's default is off through 1.35 and on from 1.36; set it explicitly to pin the behavior across upgrades.

See `GUIDE.md` for operational judgment (upgrade practice, pool sizing, firewall posture) and `catalog.md` for the deployment-store page.

## Module layout

- `v1alpha1/` — the versioned contract: `spec.proto`, `outputs.proto`, validation tests, generated reference
- `iac/tf/` and `iac/pulumi/` — the two provisioner modules implementing the same contract with identical outputs
- `iac/provider-parity.yaml` — the recorded mapping judgment against the pinned provider
- `iac/import-map.yaml` — how an existing cluster's identity derives for import
- `presets/` — ready-to-deploy starting points (production HA, development)
- `e2e/` — test profile, canonical manifests, and live-lane scenarios

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
