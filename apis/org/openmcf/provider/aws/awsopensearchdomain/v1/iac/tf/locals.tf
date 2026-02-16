locals {
  # ---------------------------------------------------------------------------
  # Metadata
  # ---------------------------------------------------------------------------

  # Stable resource ID from metadata
  resource_id = coalesce(try(var.metadata.id, null), var.metadata.name)

  tags = merge({
    "Name" = local.resource_id
  }, try(var.metadata.labels, {}))

  # ---------------------------------------------------------------------------
  # Engine
  # ---------------------------------------------------------------------------

  engine_version = var.spec.engine_version

  # ---------------------------------------------------------------------------
  # Cluster configuration
  # ---------------------------------------------------------------------------

  instance_type  = var.spec.cluster_config.instance_type
  instance_count = coalesce(try(var.spec.cluster_config.instance_count, null), 1)

  # Dedicated master nodes
  dedicated_master_enabled = coalesce(try(var.spec.cluster_config.dedicated_master_enabled, null), false)
  dedicated_master_type    = try(var.spec.cluster_config.dedicated_master_type, null)
  dedicated_master_count   = coalesce(try(var.spec.cluster_config.dedicated_master_count, null), 3)

  # Zone awareness
  zone_awareness_enabled   = coalesce(try(var.spec.cluster_config.zone_awareness_enabled, null), false)
  availability_zone_count  = coalesce(try(var.spec.cluster_config.availability_zone_count, null), 2)

  # UltraWarm storage
  warm_enabled = coalesce(try(var.spec.cluster_config.warm_enabled, null), false)
  warm_type    = try(var.spec.cluster_config.warm_type, null)
  warm_count   = coalesce(try(var.spec.cluster_config.warm_count, null), 0)

  # Cold storage
  cold_storage_enabled = coalesce(try(var.spec.cluster_config.cold_storage_enabled, null), false)

  # Multi-AZ with standby
  multi_az_with_standby_enabled = coalesce(try(var.spec.cluster_config.multi_az_with_standby_enabled, null), false)

  # ---------------------------------------------------------------------------
  # EBS options
  # ---------------------------------------------------------------------------

  ebs_enabled = coalesce(try(var.spec.ebs_options.ebs_enabled, null), true)
  volume_type = try(var.spec.ebs_options.volume_type, "gp3")
  volume_size = coalesce(try(var.spec.ebs_options.volume_size, null), 10)
  iops        = coalesce(try(var.spec.ebs_options.iops, null), 0)
  throughput  = coalesce(try(var.spec.ebs_options.throughput, null), 0)

  # ---------------------------------------------------------------------------
  # Encryption
  # ---------------------------------------------------------------------------

  encrypt_at_rest_enabled         = coalesce(try(var.spec.encrypt_at_rest_enabled, null), true)
  kms_key_id                      = try(var.spec.kms_key_id.value, null)
  node_to_node_encryption_enabled = coalesce(try(var.spec.node_to_node_encryption_enabled, null), true)

  # ---------------------------------------------------------------------------
  # VPC options
  # ---------------------------------------------------------------------------

  subnet_ids      = [for s in coalesce(try(var.spec.vpc_options.subnet_ids, []), []) : s.value]
  sg_ids          = [for s in coalesce(try(var.spec.vpc_options.security_group_ids, []), []) : s.value]
  has_vpc_options = length(local.subnet_ids) > 0

  # ---------------------------------------------------------------------------
  # Domain endpoint options
  # ---------------------------------------------------------------------------

  enforce_https       = coalesce(try(var.spec.domain_endpoint_options.enforce_https, null), true)
  tls_security_policy = try(var.spec.domain_endpoint_options.tls_security_policy, null)

  custom_endpoint_enabled         = coalesce(try(var.spec.domain_endpoint_options.custom_endpoint_enabled, null), false)
  custom_endpoint                 = try(var.spec.domain_endpoint_options.custom_endpoint, null)
  custom_endpoint_certificate_arn = try(var.spec.domain_endpoint_options.custom_endpoint_certificate_arn.value, null)

  # ---------------------------------------------------------------------------
  # Fine-grained access control (FGAC)
  # ---------------------------------------------------------------------------

  fgac_enabled                    = coalesce(try(var.spec.advanced_security_options.enabled, null), false)
  internal_user_database_enabled  = coalesce(try(var.spec.advanced_security_options.internal_user_database_enabled, null), false)
  master_user_arn                 = try(var.spec.advanced_security_options.master_user_arn.value, null)
  master_user_name                = try(var.spec.advanced_security_options.master_user_name, null)
  master_user_password            = try(var.spec.advanced_security_options.master_user_password.value, null)
  has_advanced_security           = local.fgac_enabled

  # ---------------------------------------------------------------------------
  # Log publishing options
  # ---------------------------------------------------------------------------

  log_publishing_options = coalesce(try(var.spec.log_publishing_options, []), [])

  # ---------------------------------------------------------------------------
  # Access policies
  # ---------------------------------------------------------------------------

  # access_policies is a google.protobuf.Struct — arrives as a map from the manifest.
  # Encode it as JSON for the aws_opensearch_domain resource.
  access_policies_json = try(var.spec.access_policies, null) != null ? jsonencode(var.spec.access_policies) : null

  # ---------------------------------------------------------------------------
  # Auto-Tune and maintenance
  # ---------------------------------------------------------------------------

  auto_tune_enabled              = coalesce(try(var.spec.auto_tune_enabled, null), false)
  auto_software_update_enabled   = coalesce(try(var.spec.auto_software_update_enabled, null), false)

  # ---------------------------------------------------------------------------
  # IP and advanced options
  # ---------------------------------------------------------------------------

  ip_address_type  = try(var.spec.ip_address_type, null) != "" ? try(var.spec.ip_address_type, null) : null
  advanced_options = coalesce(try(var.spec.advanced_options, null), {})
}
