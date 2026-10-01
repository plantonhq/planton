# DigitalOcean Kubernetes Node Pool -- Terraform Module

Deploys a `digitalocean_kubernetes_node_pool` from a `DigitalOceanKubernetesNodePool` spec: owning cluster, Droplet size, fixed or autoscaled node count, Kubernetes labels and taints, DigitalOcean tags, and AMD GPU partitioning. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanKubernetesNodePool`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)
- An existing DOKS cluster (the `cluster` argument is that cluster's UUID)

## Usage

```hcl
module "node_pool" {
  source = "./path/to/module"

  metadata = {
    name = "app-pool"
  }

  spec = {
    node_pool_name = "app-pool"
    cluster        = "fb7d9b81-fe06-4ee5-87f1-b9efc5af46fd"
    size           = "s-2vcpu-4gb"
    node_count     = 3
    auto_scale     = true
    min_nodes      = 2
    max_nodes      = 6
    labels = {
      workload = "app"
    }
    taints = [{
      key    = "dedicated"
      value  = "app"
      effect = "PreferNoSchedule"
    }]
    tags = ["team:platform"]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Outputs

Exactly the kind's stack-output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `node_pool_id` | The pool's UUID (import id for `digitalocean_kubernetes_node_pool`) |
| `cluster_id` | The owning cluster's UUID |

The resource's `nodes[*].id` and `nodes[*].droplet_id` are deliberately not exported: DOKS replaces nodes by design, so an apply-time list is stale the next time the pool changes shape. Droplet-scoped wiring goes through the pool's tags.

## Behavior notes

- `cluster` arrives flattened as a plain string (the owning cluster's UUID). Changing it replaces the pool.
- `min_nodes` / `max_nodes` are sent only when `auto_scale` is true; otherwise they are null so a fixed pool never carries stale bounds.
- `gpu_partition_mode` is omitted when empty. Changing it replaces the pool.
- Kubernetes node labels are user labels over the standard Planton identity labels — the exact map the Pulumi module applies.
- Tags are `spec.tags` plus the standard Planton labels rendered as `key:value` — the exact set the Pulumi module applies.
- Taint `value` is always sent, possibly empty (Kubernetes allows valueless taints; the provider requires the leaf be present).
- See the kind [GUIDE](../../GUIDE.md).
