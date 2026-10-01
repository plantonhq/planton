# DigitalOcean Kubernetes Cluster

Deploys a managed Kubernetes cluster (DOKS) on DigitalOcean with the full provider surface: configurable node sizing with labels, taints, and autoscaling, optional high availability, surge and automatic upgrades with a maintenance policy, a control-plane firewall, custom pod/service subnets, SSO, cluster-autoscaler tuning, container registry integration, and managed addon toggles. The decisions that lock in at creation deserve the most care: VPC and subnet placement can never be changed, HA can never be turned off, and changing the default node pool's size later replaces the entire cluster rather than resizing the pool.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DigitalOcean Kubernetes Cluster** -- a managed DOKS cluster in the specified region and VPC, at the configured Kubernetes version, with optional HA control plane, surge and auto upgrades, custom subnets, and SSO
- **Default Node Pool** -- worker nodes with the configured instance size, node count, optional autoscaling between `minNodes` and `maxNodes`, Kubernetes node labels and taints, and pool tags
- **Control Plane Firewall** -- created only when `controlPlaneFirewall` is provided; restricts API server access to the listed IPs and CIDR blocks
- **Maintenance Policy** -- created only when `maintenancePolicy` is provided; pins maintenance and auto-upgrades to a weekly slot
- **Cluster Autoscaler Configuration** -- created only when `clusterAutoscalerConfiguration` is provided; tunes scale-down behavior and expander strategies
- **Managed Addons** -- routing agent, CoreDNS autoscaler, GPU device plugins/DRA drivers, RDMA, and P2P OCI registry toggles, each asserted only when set
- **Container Registry Integration** -- configured only when `registryIntegration` is enabled; wires the account's DigitalOcean Container Registry into the cluster
- **DigitalOcean Tags** -- your `tags` plus resource metadata tags (organization, environment, resource kind) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A VPC network** in the target region (required). Provide the VPC UUID directly or reference a DigitalOceanVpc Cloud Resource via ValueFromRef.
- **A Kubernetes version DigitalOcean offers today** -- check with `doctl kubernetes options versions` (or `GET /v2/kubernetes/options`). Prefer a minor prefix (`"1.35"`): DigitalOcean resolves it to the current patch, and it stays creatable for the minor's whole support window. A full slug (`"1.35.7-do.5"`) pins the exact starting point but is retired within weeks, after which a create naming it fails with 422. Measured 2026-09-16: 1.34, 1.35, 1.36 offered.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Kubernetes Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production HA Kubernetes Cluster** preset in the [Presets](#presets) tab for a resilient configuration with autoscaling and an API firewall.

### CLI

Create a manifest and apply it:

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
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  defaultNodePool:
    size: s-4vcpu-8gb
    nodeCount: 3
```

```shell
planton apply -f do-k8s-cluster.yaml
```

This creates a 3-node Kubernetes cluster in the NYC3 region with everything else at DigitalOcean defaults. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to a VPC deployed in the same InfraPipeline:

```yaml
spec:
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then provisions the Kubernetes cluster on it.

## Key Configuration

These are the most important decisions when configuring a DOKS cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**High availability** -- Set `highlyAvailable: true` to provision multiple control-plane replicas for fault tolerance. HA increases cost, is required for production SLAs, and is one-way: it cannot be turned off once enabled.

**Node pool sizing and autoscaling** -- `defaultNodePool.size` sets the instance type (`"s-2vcpu-4gb"` for development, `"s-4vcpu-8gb"` for production). Changing it later replaces the entire cluster, so size deliberately and grow with separate `DigitalOceanKubernetesNodePool` resources. Enable `autoScale` with `minNodes` and `maxNodes` to let DOKS adjust node count with scheduling pressure.

**Automatic upgrades** -- Set `autoUpgrade: true` to let DigitalOcean apply Kubernetes patch updates, and pin them to a weekly slot with `maintenancePolicy` (`day` plus `startTime` in UTC). `kubernetesVersion` is the creation pin only; both provisioners deliberately ignore later drift on it.

**Control-plane firewall** -- Provide `controlPlaneFirewall` with `enabled: true` and the IPs/CIDRs allowed to reach the Kubernetes API server. When omitted, the API server is publicly accessible. Restrict to VPN or office CIDRs for production -- and make sure the list includes wherever the provisioner runs.

**Addons and advanced placement** -- all nine addon toggles (`routingAgent`, `p2pOciRegistryPlugin`, `corednsAutoscaler`, the AMD and NVIDIA device plugins and DRA drivers, `amdGpuDeviceMetricsExporterPlugin`, `rdmaSharedDevicePlugin`), `sso`, `isolatedWorkers`, `workerSubnetUuid`, and `gpuPartitionMode` deploy on both provisioners. DigitalOcean enforces the prerequisites: a NAT gateway on the VPC for `isolatedWorkers`, a subnet in that VPC for `workerSubnetUuid`, GPU node sizes for the GPU addons and partitioning, and Kubernetes 1.36.0-do.2 or later for `p2pOciRegistryPlugin` (an older version fails the create with a validation 422). `corednsAutoscaler` defaults off through 1.35 and on from 1.36.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanVpc** | `vpc` | `status.outputs.vpc_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Unique cluster identifier (UUID) on DigitalOcean | Node pool attachment, firewall source/destination rules, database firewall trusted sources, DigitalOcean API operations |
| `kubeconfig` | Raw kubeconfig YAML for cluster access (sensitive) | Kubernetes Provider Connections, CI/CD pipeline configuration |
| `api_server_endpoint` | Kubernetes API server endpoint URL | kubectl configuration, application health checks |
| `urn` | `do:kubernetes:<cluster_id>` | DigitalOcean project attachment |
| `ipv4_address` | Control plane public IPv4 when DigitalOcean reports one; clusters created today report none (single-replica included), so allowlist by `api_server_endpoint` | Network allowlists (legacy clusters only) |
| `default_node_pool_id` | The inline default pool's UUID | DigitalOcean API operations on the pool |
| `cluster_subnet` / `service_subnet` | Pod and service CIDR blocks in effect | VPC peering and routing plans |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production HA cluster** -- HA control plane, autoscaling node pool (2-5 nodes), automatic patch upgrades in a Sunday maintenance window, container registry integration, and a control-plane firewall. Start from the **Production HA Kubernetes Cluster** preset.

**Development cluster** -- Non-HA control plane, 2 fixed-size nodes with smaller instances, no auto-upgrade or API firewall. Minimal cost for dev/test and feature-branch clusters. Start from the **Development Kubernetes Cluster** preset.

## Works With

- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- provides the VPC network for cluster placement (required)
- [**DigitalOcean Kubernetes Node Pool**](/cloud-catalog/digital-ocean-kubernetes-node-pool) -- adds independently sized worker pools to the cluster
- [**DigitalOcean Cloud Firewall**](/cloud-catalog/digital-ocean-firewall) -- Droplet firewall rules that allow traffic from or to the cluster by its `cluster_id`
- [**DigitalOcean Database Firewall**](/cloud-catalog/digital-ocean-database-firewall) -- trusts the cluster as a source so in-cluster workloads can reach a managed database
