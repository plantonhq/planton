locals {
  # resource name and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-eks-cluster")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # safe settings with defaults
  safe_subnet_ids              = try(var.spec.subnet_ids, [])
  safe_cluster_role_arn        = try(var.spec.cluster_role_arn, {})
  safe_version                 = try(var.spec.version, null)
  safe_disable_public_endpoint = try(var.spec.disable_public_endpoint, false)
  safe_public_access_cidrs     = try(var.spec.public_access_cidrs, [])
  safe_enable_control_plane_logs = try(var.spec.enable_control_plane_logs, false)
  safe_kms_key_arn             = try(var.spec.kms_key_arn, {})

  # computed values
  subnet_ids              = [for subnet in local.safe_subnet_ids : subnet.value]
  cluster_role_arn        = local.safe_cluster_role_arn.value
  cluster_role_name       = split("/", local.cluster_role_arn)[length(split("/", local.cluster_role_arn)) - 1]
  cluster_version         = local.safe_version
  disable_public_endpoint = local.safe_disable_public_endpoint
  public_access_cidrs     = local.safe_public_access_cidrs
  enable_control_plane_logs = local.safe_enable_control_plane_logs
  kms_key_arn             = local.safe_kms_key_arn.value
  has_kms_key             = local.kms_key_arn != null && local.kms_key_arn != ""
}


