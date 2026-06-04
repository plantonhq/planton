# DigitalOcean Kubernetes Cluster

Deploys a managed Kubernetes cluster on DigitalOcean with a configurable default node pool, optional high-availability control plane, automatic patch upgrades, and control plane firewall restrictions. The component supports VPC placement, container registry integration, and surge upgrades out of the box.

## What Gets Created

When you deploy a DigitalOceanKubernetesCluster resource, OpenMCF provisions:

- **Kubernetes Cluster** — a `digitalocean_kubernetes_cluster` resource with the specified Kubernetes version, region, and VPC, including a default node pool with the configured Droplet size and node count
- **Default Node Pool** — embedded in the cluster resource, supports fixed sizing or auto-scaling between configurable min/max node counts
- **Maintenance Policy** — created only when `maintenanceWindow` is specified, sets the preferred time window for cluster updates
- **Control Plane Firewall** — created only when `controlPlaneFirewallAllowedIps` is specified, restricts API server access to the listed CIDR ranges

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **A DigitalOcean VPC** in the target region (can reference a DigitalOceanVpc resource via `valueFrom`)
- **A supported Kubernetes version** available in the chosen region (e.g., `1.31.1-do.5`)

## Quick Start

Create a file `doks.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanKubernetesCluster.my-cluster
spec:
  clusterName: my-cluster
  region: nyc3
  kubernetesVersion: "1.31.1-do.5"
  vpc:
    value: "vpc-uuid-here"
  defaultNodePool:
    size: s-4vcpu-8gb
    nodeCount: 3
```

Deploy:

```shell
openmcf apply -f doks.yaml
```

This creates a three-node Kubernetes cluster in the NYC3 region with `s-4vcpu-8gb` Droplets and surge upgrades enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterName` | `string` | Name of the Kubernetes cluster in DigitalOcean. Must be unique per account. | Required |
| `region` | `enum` | DigitalOcean region for the cluster. Valid values: `nyc3`, `sfo3`, `fra1`, `sgp1`, `lon1`, `tor1`, `blr1`, `ams3`. | Required |
| `kubernetesVersion` | `string` | Kubernetes version to deploy (e.g., `1.31.1-do.5`). Must be a version supported by DigitalOcean. | Required |
| `vpc` | `StringValueOrRef` | VPC UUID where the cluster resides. Can reference a DigitalOceanVpc resource via `valueFrom`. | Required |
| `defaultNodePool.size` | `string` | Droplet size slug for nodes (e.g., `s-4vcpu-8gb`). Determines CPU and memory per node. | Required |
| `defaultNodePool.nodeCount` | `uint32` | Number of nodes in the default pool. Acts as initial desired count when auto-scaling is enabled. | Required, must be > 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `highlyAvailable` | `bool` | `false` | Enables a high-availability control plane with multiple masters. Incurs additional cost. |
| `autoUpgrade` | `bool` | `false` | Automatically upgrades the cluster to new Kubernetes patch releases when available. |
| `disableSurgeUpgrade` | `bool` | `false` | When `true`, disables surge upgrades. By default, upgrades temporarily provision extra nodes to minimize downtime. |
| `maintenanceWindow` | `string` | — | Preferred maintenance window for cluster updates. Format: `day=HH:MM` or `any=HH:MM` (e.g., `sunday=02:00`). |
| `registryIntegration` | `bool` | `false` | Enables DigitalOcean Container Registry (DOCR) integration, automatically creating imagePullSecrets for private images. |
| `controlPlaneFirewallAllowedIps` | `string[]` | `[]` | CIDR ranges allowed to access the Kubernetes API server. If empty, the API server is publicly accessible. |
| `tags` | `string[]` | `[]` | Tags to apply to the cluster for organization within DigitalOcean. |
| `defaultNodePool.autoScale` | `bool` | `false` | Enables auto-scaling for the default node pool. |
| `defaultNodePool.minNodes` | `uint32` | `0` | Minimum node count when auto-scaling is enabled. Required if `defaultNodePool.autoScale` is `true`. |
| `defaultNodePool.maxNodes` | `uint32` | `0` | Maximum node count when auto-scaling is enabled. Required if `defaultNodePool.autoScale` is `true`. |

## Examples

### HA Cluster with Auto-Upgrade

A production cluster with a highly available control plane and automatic patch upgrades:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanKubernetesCluster.prod-cluster
spec:
  clusterName: prod-cluster
  region: fra1
  kubernetesVersion: "1.31.1-do.5"
  vpc:
    value: "vpc-prod-uuid"
  highlyAvailable: true
  autoUpgrade: true
  maintenanceWindow: "sunday=03:00"
  defaultNodePool:
    size: s-4vcpu-8gb
    nodeCount: 3
```

### Auto-Scaling Node Pool with API Firewall

A cluster with auto-scaling nodes and restricted API server access for security:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesCluster
metadata:
  name: secure-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanKubernetesCluster.secure-cluster
spec:
  clusterName: secure-cluster
  region: sfo3
  kubernetesVersion: "1.31.1-do.5"
  vpc:
    value: "vpc-secure-uuid"
  highlyAvailable: true
  controlPlaneFirewallAllowedIps:
    - "203.0.113.0/24"
    - "198.51.100.5/32"
  tags:
    - production
    - team-platform
  defaultNodePool:
    size: s-8vcpu-16gb
    nodeCount: 3
    autoScale: true
    minNodes: 2
    maxNodes: 10
```

### Full-Featured Cluster with Registry and VPC Reference

Production configuration using a VPC foreign key reference, DOCR integration, and all optional features:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesCluster
metadata:
  name: full-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanKubernetesCluster.full-cluster
spec:
  clusterName: full-cluster
  region: ams3
  kubernetesVersion: "1.31.1-do.5"
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: prod-vpc
      field: status.outputs.vpc_id
  highlyAvailable: true
  autoUpgrade: true
  disableSurgeUpgrade: false
  maintenanceWindow: "saturday=04:00"
  registryIntegration: true
  controlPlaneFirewallAllowedIps:
    - "10.0.0.0/8"
  tags:
    - production
    - managed
  defaultNodePool:
    size: s-8vcpu-16gb
    nodeCount: 5
    autoScale: true
    minNodes: 3
    maxNodes: 15
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | UUID of the created Kubernetes cluster |
| `kubeconfig` | `string` | Base64-encoded kubeconfig for accessing the cluster |
| `api_server_endpoint` | `string` | Endpoint URL of the Kubernetes API server |

## Related Components

- [DigitalOceanVpc](/docs/catalog/digitalocean/digitaloceanvpc) — provides the VPC for cluster placement
- [DigitalOceanKubernetesNodePool](/docs/catalog/digitalocean/digitaloceankubernetesnodepool) — adds additional node pools to the cluster
- [DigitalOceanContainerRegistry](/docs/catalog/digitalocean/digitaloceancontainerregistry) — hosts private container images for registry integration
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/digitaloceanloadbalancer) — provisions load balancers for exposing cluster services
- [DigitalOceanFirewall](/docs/catalog/digitalocean/digitaloceanfirewall) — controls network access to cluster nodes
