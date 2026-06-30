locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "click_house_kubernetes"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && try(var.metadata.env, "") != ""
    ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Namespace is always the value from spec - the create_namespace flag only controls whether we create it
  namespace = var.spec.namespace

  # Determine cluster name - use spec.cluster_name if provided, otherwise use metadata.name
  cluster_name = coalesce(try(var.spec.cluster_name, ""), var.metadata.name)

  # Determine ClickHouse version - use spec.version if provided, otherwise use default
  clickhouse_version = coalesce(try(var.spec.version, ""), "24.8")

  # The service name for ClickHouse
  kube_service_name = var.metadata.name

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "clickhouse-my-db")
  password_secret_name     = "${var.metadata.name}-password"
  external_lb_service_name = "${var.metadata.name}-external-lb"
  keeper_installation_name = "${var.metadata.name}-keeper"
  # Altinity operator creates keeper service with pattern: keeper-<chk-name>
  keeper_service_name = "keeper-${local.keeper_installation_name}"

  # Altinity operator uses these labels for pod selection
  # These labels are automatically applied by the operator to ClickHouse pods
  clickhouse_pod_selector_labels = {
    "clickhouse.altinity.com/chi"     = local.cluster_name
    "clickhouse.altinity.com/cluster" = local.cluster_name
  }

  # Safely handle optional ingress values
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Cluster configuration
  cluster_is_enabled = try(var.spec.cluster.is_enabled, false)
  shard_count        = local.cluster_is_enabled ? try(var.spec.cluster.shard_count, 1) : 1
  replica_count      = local.cluster_is_enabled ? try(var.spec.cluster.replica_count, 1) : var.spec.container.replicas

  # ClickHouse authentication defaults
  default_username    = "default"
  password_secret_key = "admin-password"

  # ZooKeeper configuration
  zookeeper_use_external = try(var.spec.zookeeper.use_external, false)
  zookeeper_nodes        = try(var.spec.zookeeper.nodes, [])

  # Build ZooKeeper configuration
  zookeeper_config = local.zookeeper_use_external && length(local.zookeeper_nodes) > 0 ? {
    nodes = [for node in local.zookeeper_nodes : {
      host = node
    }]
    } : {
    # Operator-managed ZooKeeper
    nodes = [{
      host = "zookeeper"
      port = 2181
    }]
  }
}
