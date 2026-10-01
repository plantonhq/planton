# DigitalOcean Kubernetes Node Pool -- Pulumi Module

Deploys a `digitalocean:index/kubernetesNodePool:KubernetesNodePool` from a `DigitalOceanKubernetesNodePool` stack input: owning cluster, Droplet size, fixed or autoscaled node count, Kubernetes labels and taints, and DigitalOcean tags. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`.

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, node pool
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/node_pool.go` -- the node-pool resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `node_pool_id`, `cluster_id`.

The pool's nodes (`Nodes[*].Id`, `Nodes[*].DropletId`) are deliberately not exported: DOKS replaces nodes by design (autoscaling, upgrades, auto-repair), so an apply-time list is stale the next time the pool changes shape. Droplet-scoped wiring goes through the pool's tags.

## Behavior notes

- The full spec surface is wired; there is no Pulumi SDK gap on this resource at the pin. `gpu_partition_mode` arrives as null when unset (the provider rejects `""`) and is create-only -- changing it replaces the pool, exactly as on Terraform.

- `cluster` resolves to the owning DOKS cluster's UUID.
- Autoscaling bounds (`autoScale` / `minNodes` / `maxNodes`) are sent only when `autoScale` is true.
- Kubernetes node labels are user labels over the standard Planton identity labels — the exact map the Terraform module applies.
- Tags are `spec.tags` plus the standard Planton labels rendered as `key:value` — the exact set the Terraform module applies.
- See the kind [GUIDE](../../GUIDE.md).
