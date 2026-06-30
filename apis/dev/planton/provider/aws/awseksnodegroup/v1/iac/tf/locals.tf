locals {
  metadata = var.metadata
  spec     = var.spec

  # Flatten foreign key references to primitive values
  cluster_name = coalesce(try(var.spec.cluster_name.value, null), try(var.spec.cluster_name.value_from.name, null))
  node_role_arn = coalesce(try(var.spec.node_role_arn.value, null), try(var.spec.node_role_arn.value_from.name, null))
  subnet_ids    = [for s in try(var.spec.subnet_ids, []) : coalesce(try(s.value, null), try(s.value_from.name, null))]

  # Capacity type mapping to AWS provider expected strings
  is_spot       = lower(try(var.spec.capacity_type, "on_demand")) == "spot"
  capacity_type = local.is_spot ? "SPOT" : "ON_DEMAND"

  # Safety helpers
  disk_size_gb = try(var.spec.disk_size_gb, null)
  ssh_key_name = try(var.spec.ssh_key_name, null)

  # Labels map may be empty; pass through directly
  labels = try(var.spec.labels, {})
}



