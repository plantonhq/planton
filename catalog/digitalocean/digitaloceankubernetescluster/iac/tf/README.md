# DigitalOcean Kubernetes Cluster -- Terraform Module

Deploys a `digitalocean_kubernetes_cluster` from a `DigitalOceanKubernetesCluster` spec: version/region/VPC placement, the inline default node pool (labels, taints, tags, autoscaling, GPU partitioning), HA control plane, surge and auto upgrades, maintenance policy, control-plane firewall, pod/service/worker subnets, isolated workers, SSO, cluster-autoscaler tuning, registry integration, kubeconfig expiry, destroy-time cleanup, and every managed addon toggle. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanKubernetesCluster`). Do not hand-edit it. The API token lives in `credentials.tf`.

Additional node pools are separate `digitalocean_kubernetes_node_pool` resources, not part of this module.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "kubernetes_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "app-cluster"
  }

  spec = {
    cluster_name       = "app-cluster"
    region             = "nyc3"
    kubernetes_version = "1.35"
    vpc                = "b5648f9e-a28a-4760-bb87-b2fad07ae295"
    highly_available   = true
    auto_upgrade       = true
    maintenance_policy = {
      day        = "sunday"
      start_time = "03:00"
    }
    control_plane_firewall = {
      enabled           = true
      allowed_addresses = ["203.0.113.0/24"]
    }
    default_node_pool = {
      size       = "s-4vcpu-8gb"
      node_count = 3
      auto_scale = true
      min_nodes  = 2
      max_nodes  = 5
    }
  }

  digitalocean_token = var.digitalocean_token
}

output "kubeconfig" {
  value     = module.kubernetes_cluster.kubeconfig
  sensitive = true
}
```

## Outputs

Exactly the kind's stack-output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `cluster_id` | The cluster UUID (import id for `digitalocean_kubernetes_cluster`) |
| `kubeconfig` | Raw kubeconfig YAML (sensitive; not base64) |
| `api_server_endpoint` | Kubernetes API server URL |
| `urn` | `do:kubernetes:<cluster_id>` |
| `ipv4_address` | Control plane public IPv4 when DigitalOcean reports one -- empty on clusters created today, single-replica included; use `api_server_endpoint` |
| `default_node_pool_id` | The inline default pool's UUID |
| `cluster_subnet` / `service_subnet` | Pod and service CIDR blocks in effect |

## Behavior notes

- `lifecycle.ignore_changes = [version]` is a safety rail, not a convenience: auto-upgrade moves the live version ahead of the pin, and the provider destroys and recreates the cluster when the configured version is LOWER than the live one. The version is creation-only; upgrades ride `auto_upgrade`.
- `surge_upgrade` passes through null when unset so the provider's default (true) applies -- it is never coalesced to false.
- Changing the default pool's `size` or `gpu_partition_mode` replaces the entire cluster (provider ForceNew inside the `node_pool` block).
- The provider itself appends the `terraform:default-node-pool` marker tag to the inline pool; the module never sets it.
- Addon blocks (`routing_agent`, GPU plugins/DRA drivers, `rdma_shared_device_plugin`, `coredns_autoscaler`, `p2p_oci_registry_plugin`) are emitted only when the spec sets them, so unset defers to DigitalOcean's own default per addon. See the kind [GUIDE](../../GUIDE.md).
