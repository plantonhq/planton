# DigitalOcean Kubernetes Cluster -- Pulumi Module

Deploys a `digitalocean:index/kubernetesCluster:KubernetesCluster` from a `DigitalOceanKubernetesCluster` stack input: version/region/VPC placement, the inline default node pool (labels, taints, tags, autoscaling), HA control plane, surge and auto upgrades, maintenance policy, control-plane firewall, pod/service subnets, cluster-autoscaler tuning, registry integration, kubeconfig expiry, destroy-time cleanup, and the routing-agent, AMD GPU device plugin, and AMD GPU metrics exporter addons. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.53.0`.

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

These spec fields are real and the Terraform module wires them. This module fails the apply with `PARITY-EXCEPTION` if they are meaningfully set (proto zero values pass), until the Pulumi DigitalOcean SDK grows the matching args:

- **`spec.sso`** -- no sso block on KubernetesCluster
- **`spec.isolated_workers`** -- no isolated_workers field
- **`spec.worker_subnet_uuid`** -- no worker_subnet_uuid field
- **`spec.default_node_pool.gpu_partition_mode`** -- no gpu_partition_mode on the pool
- **Six of the nine addon blocks** -- `p2p_oci_registry_plugin`, `amd_gpu_dra_driver`, `nvidia_gpu_device_plugin`, `nvidia_gpu_dra_driver`, `rdma_shared_device_plugin`, `coredns_autoscaler` -- do not exist in the SDK. `routing_agent`, `amd_gpu_device_plugin`, and `amd_gpu_device_metrics_exporter_plugin` do and are wired.

Every guard names the SDK pin it was verified against (v4.53.0) and is re-checked on every bump: a guard for a surface the SDK has since grown is deleted and the arm wired.

- `surge_upgrade` is sent only when present in the spec so the provider's default (true) applies when unset -- never coalesced to false.
- `version` carries `IgnoreChanges`: auto-upgrade moves the live version ahead of the pin, and the provider destroys and recreates the cluster when the configured version is lower than the live one. The pin is creation-only.
- Tags are the user's `spec.tags` plus the standard Planton labels rendered as `key:value` strings; the default pool's node labels are the same Planton labels under the user's labels -- both identical to the Terraform module.
- The SDK models the autoscaler configuration as an array; the provider reads only the first element, so the spec's single message is wrapped one-element here. See the kind [GUIDE](../../GUIDE.md).
