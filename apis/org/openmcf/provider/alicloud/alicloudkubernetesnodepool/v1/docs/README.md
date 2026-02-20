# AlicloudKubernetesNodePool Provider Research

## Provider Resources

- **Terraform**: `alicloud_cs_kubernetes_node_pool` (alicloud provider ~1.200+)
- **Pulumi**: `cs.NodePool` (pulumi-alicloud v3)

## Field Coverage

The Terraform resource exposes ~150+ fields. This v1 component covers the core
80% that satisfies production use cases:

**Included**: instance types, desired size, system/data disks, auto-scaling,
management (auto-repair/upgrade), spot instances, labels, taints, billing,
security groups, authentication, runtime configuration.

**Excluded for v2**: kubelet_configuration (30+ sub-fields), instance_patterns
(flexible matching), tee_config (confidential computing), eflo_node_group (HPC),
private_pool_options, upgrade_policy/rolling_policy (update-only operations),
and other niche features.

## Key Provider Behaviors

- `cluster_id` is ForceNew -- cannot move a node pool between clusters.
- `security_group_ids` is ForceNew -- cannot change security groups after creation.
- `desired_size` is a string in the TF schema (unusual); the proto uses int32.
- `name` is deprecated since v1.219.0; use `node_pool_name` instead.
- `node_count` is deprecated since v1.158.0; use `desired_size` instead.
- `security_group_id` (singular) is deprecated since v1.145.0; use `security_group_ids`.
- `platform` is deprecated since v1.145.0; use `image_type`.
- Auto-scaling (`scaling_config`) and `desired_size` are compatible -- desired_size
  sets the initial count, and the auto-scaler adjusts within min/max bounds.
- Labels use repeated key/value objects in the provider but map<string,string> in
  the proto for user convenience; the IaC modules handle conversion.
