# DigitalOcean Kubernetes Cluster -- Pulumi Module

Deploys a `digitalocean:index/kubernetesCluster:KubernetesCluster` from a `DigitalOceanKubernetesCluster` stack input: version/region/VPC placement, the inline default node pool (labels, taints, tags, autoscaling), HA control plane, surge and auto upgrades, maintenance policy, control-plane firewall, pod/service subnets, cluster-autoscaler tuning, registry integration, kubeconfig expiry, destroy-time cleanup, single sign-on, isolated workers and worker subnet placement, GPU partitioning on the default pool, and all nine managed addon toggles. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.79.1`.

Additional node pools are separate `KubernetesNodePool` resources, not part of this module.

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, cluster
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/cluster.go` -- the cluster resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `cluster_id`, `kubeconfig`, `api_server_endpoint`, `urn`, `ipv4_address`, `default_node_pool_id`, `cluster_subnet`, `service_subnet`. The kubeconfig is a Pulumi secret output.

## Behavior notes

- The full spec surface is wired; there is no Pulumi SDK gap on this resource at the pin. `sso` is a single message in the spec and a one-element `Ssos` array on the SDK (the provider reads only the first element -- the same wrapping the cluster-autoscaler configuration uses). Empty `issuer_url`/`client_id` strings arrive as null. `worker_subnet_uuid` and the pool's `gpu_partition_mode` arrive as null when unset (the provider rejects `""`); `isolated_workers` is always sent as the spec's bool. Each addon block is sent only when its spec message is present, asserting ON or OFF -- the Terraform module's dynamic-block contract. The module lets DigitalOcean's own validation enforce the addon prerequisites (GPU node sizes for the GPU family; Kubernetes 1.36.0-do.2 or later for the P2P OCI registry plugin, which otherwise fails the create with a 422).
- `surge_upgrade` is sent only when present in the spec so the provider's default (true) applies when unset -- never coalesced to false.
- `version` carries `IgnoreChanges`: auto-upgrade moves the live version ahead of the pin, and the provider destroys and recreates the cluster when the configured version is lower than the live one. The pin is creation-only.
- Tags are the user's `spec.tags` plus the standard Planton labels rendered as `key:value` strings; the default pool's node labels are the same Planton labels under the user's labels -- both identical to the Terraform module.
- The SDK models the autoscaler configuration as an array; the provider reads only the first element, so the spec's single message is wrapped one-element here. See the kind [GUIDE](../../GUIDE.md).
