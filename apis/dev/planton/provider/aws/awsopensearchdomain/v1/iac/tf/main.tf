# ---------------------------------------------------------------------------
# OpenSearch Service domain
# ---------------------------------------------------------------------------

resource "aws_opensearch_domain" "this" {
  domain_name    = local.resource_id
  engine_version = local.engine_version

  # ---------------------------------------------------------------------------
  # Cluster configuration
  # ---------------------------------------------------------------------------

  cluster_config {
    instance_type  = local.instance_type
    instance_count = local.instance_count

    # Dedicated master nodes
    dedicated_master_enabled = local.dedicated_master_enabled
    dedicated_master_type    = local.dedicated_master_enabled ? local.dedicated_master_type : null
    dedicated_master_count   = local.dedicated_master_enabled ? local.dedicated_master_count : null

    # Zone awareness
    zone_awareness_enabled = local.zone_awareness_enabled

    dynamic "zone_awareness_config" {
      for_each = local.zone_awareness_enabled ? [1] : []
      content {
        availability_zone_count = local.availability_zone_count
      }
    }

    # UltraWarm storage
    warm_enabled = local.warm_enabled
    warm_type    = local.warm_enabled ? local.warm_type : null
    warm_count   = local.warm_enabled ? local.warm_count : null

    # Cold storage
    dynamic "cold_storage_options" {
      for_each = local.cold_storage_enabled ? [1] : []
      content {
        enabled = true
      }
    }

    # Multi-AZ with standby
    multi_az_with_standby_enabled = local.multi_az_with_standby_enabled
  }

  # ---------------------------------------------------------------------------
  # EBS storage
  # ---------------------------------------------------------------------------

  ebs_options {
    ebs_enabled = local.ebs_enabled
    volume_type = local.ebs_enabled ? local.volume_type : null
    volume_size = local.ebs_enabled ? local.volume_size : null
    iops        = local.ebs_enabled && local.iops > 0 ? local.iops : null
    throughput  = local.ebs_enabled && local.throughput > 0 ? local.throughput : null
  }

  # ---------------------------------------------------------------------------
  # Encryption
  # ---------------------------------------------------------------------------

  encrypt_at_rest {
    enabled    = local.encrypt_at_rest_enabled
    kms_key_id = local.kms_key_id
  }

  node_to_node_encryption {
    enabled = local.node_to_node_encryption_enabled
  }

  # ---------------------------------------------------------------------------
  # VPC (conditional)
  # ---------------------------------------------------------------------------

  dynamic "vpc_options" {
    for_each = local.has_vpc_options ? [1] : []
    content {
      subnet_ids         = local.subnet_ids
      security_group_ids = length(local.sg_ids) > 0 ? local.sg_ids : null
    }
  }

  # ---------------------------------------------------------------------------
  # Domain endpoint options
  # ---------------------------------------------------------------------------

  domain_endpoint_options {
    enforce_https                   = local.enforce_https
    tls_security_policy             = local.tls_security_policy
    custom_endpoint_enabled         = local.custom_endpoint_enabled
    custom_endpoint                 = local.custom_endpoint_enabled ? local.custom_endpoint : null
    custom_endpoint_certificate_arn = local.custom_endpoint_enabled ? local.custom_endpoint_certificate_arn : null
  }

  # ---------------------------------------------------------------------------
  # Advanced security options / FGAC (conditional)
  # ---------------------------------------------------------------------------

  dynamic "advanced_security_options" {
    for_each = local.has_advanced_security ? [1] : []
    content {
      enabled                        = local.fgac_enabled
      internal_user_database_enabled = local.internal_user_database_enabled

      dynamic "master_user_options" {
        for_each = local.master_user_arn != null || local.master_user_name != null ? [1] : []
        content {
          master_user_arn      = local.master_user_arn
          master_user_name     = local.master_user_name
          master_user_password = local.master_user_password
        }
      }
    }
  }

  # ---------------------------------------------------------------------------
  # Log publishing (conditional, repeated)
  # ---------------------------------------------------------------------------

  dynamic "log_publishing_options" {
    for_each = local.log_publishing_options
    content {
      log_type                 = log_publishing_options.value.log_type
      cloudwatch_log_group_arn = try(log_publishing_options.value.cloudwatch_log_group_arn.value, "")
      enabled                  = try(log_publishing_options.value.enabled, true)
    }
  }

  # ---------------------------------------------------------------------------
  # Access policies
  # ---------------------------------------------------------------------------

  access_policies = local.access_policies_json

  # ---------------------------------------------------------------------------
  # Auto-Tune (conditional)
  # ---------------------------------------------------------------------------

  dynamic "auto_tune_options" {
    for_each = local.auto_tune_enabled ? [1] : []
    content {
      desired_state = "ENABLED"
    }
  }

  # ---------------------------------------------------------------------------
  # Software update options
  # ---------------------------------------------------------------------------

  software_update_options {
    auto_software_update_enabled = local.auto_software_update_enabled
  }

  # ---------------------------------------------------------------------------
  # IP address type
  # ---------------------------------------------------------------------------

  ip_address_type = local.ip_address_type

  # ---------------------------------------------------------------------------
  # Advanced options
  # ---------------------------------------------------------------------------

  advanced_options = length(local.advanced_options) > 0 ? local.advanced_options : null

  # ---------------------------------------------------------------------------
  # Tags
  # ---------------------------------------------------------------------------

  tags = local.tags
}
