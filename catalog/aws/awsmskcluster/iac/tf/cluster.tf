# --- Module-managed MSK Configuration (folded) ---
# Created only for inline server_properties. An update to the properties bumps
# the configuration's revision, and the cluster follows via configuration_info
# -- a rolling broker restart, never a replacement.

resource "aws_msk_configuration" "this" {
  count = local.manage_configuration ? 1 : 0

  name              = local.cluster_name
  kafka_versions    = [var.spec.kafka_version]
  server_properties = local.server_properties
}

# --- MSK Cluster ---
# Broker networking (subnets, security groups) and encryption_info are
# create-time in AWS: changing them replaces the cluster. Compute
# (instance_type), storage size, broker count (increase only), monitoring,
# configuration, and connectivity are all granular in-place update operations.

resource "aws_msk_cluster" "this" {
  cluster_name           = local.cluster_name
  kafka_version          = var.spec.kafka_version
  number_of_broker_nodes = var.spec.number_of_broker_nodes
  enhanced_monitoring    = var.spec.enhanced_monitoring != "" ? var.spec.enhanced_monitoring : null
  storage_mode           = var.spec.storage_mode != "" ? var.spec.storage_mode : null
  tags                   = local.aws_tags

  broker_node_group_info {
    instance_type  = var.spec.instance_type
    client_subnets = var.spec.subnet_ids
    # Attached directly -- ingress rules live on the referenced first-class
    # security-group nodes, never on a module-managed shadow group.
    security_groups = var.spec.security_group_ids

    dynamic "storage_info" {
      for_each = var.spec.ebs_volume_size_gib != null || var.spec.provisioned_throughput_enabled ? [1] : []
      content {
        ebs_storage_info {
          volume_size = var.spec.ebs_volume_size_gib

          dynamic "provisioned_throughput" {
            for_each = var.spec.provisioned_throughput_enabled ? [1] : []
            content {
              enabled           = true
              volume_throughput = var.spec.provisioned_throughput_mbs
            }
          }
        }
      }
    }

    # AWS activates public access, PrivateLink auth schemes, and dual-stack
    # addressing as follow-up connectivity updates after the cluster is
    # created; the provider drives that create-then-update flow from this one
    # declarative block.
    dynamic "connectivity_info" {
      for_each = local.manage_connectivity_info ? [1] : []
      content {
        network_type = var.spec.network_type != "" ? var.spec.network_type : null

        dynamic "public_access" {
          for_each = var.spec.public_access_type != "" ? [1] : []
          content {
            type = var.spec.public_access_type
          }
        }

        dynamic "vpc_connectivity" {
          for_each = local.vpc_connectivity_enabled ? [var.spec.vpc_connectivity] : []
          content {
            client_authentication {
              tls = vpc_connectivity.value.tls_enabled

              dynamic "sasl" {
                for_each = vpc_connectivity.value.sasl_iam_enabled || vpc_connectivity.value.sasl_scram_enabled ? [1] : []
                content {
                  iam   = vpc_connectivity.value.sasl_iam_enabled
                  scram = vpc_connectivity.value.sasl_scram_enabled
                }
              }
            }
          }
        }
      }
    }
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = var.spec.kms_key_arn != "" ? var.spec.kms_key_arn : null

    encryption_in_transit {
      client_broker = var.spec.client_broker_encryption
      in_cluster    = var.spec.in_cluster_encryption
    }
  }

  dynamic "client_authentication" {
    for_each = var.spec.authentication != null ? [var.spec.authentication] : []
    content {
      unauthenticated = client_authentication.value.unauthenticated

      dynamic "sasl" {
        for_each = client_authentication.value.sasl_iam_enabled || client_authentication.value.sasl_scram_enabled ? [1] : []
        content {
          iam   = client_authentication.value.sasl_iam_enabled
          scram = client_authentication.value.sasl_scram_enabled
        }
      }

      dynamic "tls" {
        for_each = client_authentication.value.tls_enabled ? [1] : []
        content {
          certificate_authority_arns = client_authentication.value.tls_certificate_authority_arns
        }
      }
    }
  }

  dynamic "configuration_info" {
    for_each = local.manage_configuration || var.spec.configuration_arn != "" ? [1] : []
    content {
      arn      = local.manage_configuration ? aws_msk_configuration.this[0].arn : var.spec.configuration_arn
      revision = local.manage_configuration ? aws_msk_configuration.this[0].latest_revision : var.spec.configuration_revision
    }
  }

  dynamic "logging_info" {
    for_each = var.spec.logging != null ? [var.spec.logging] : []
    content {
      broker_logs {
        dynamic "cloudwatch_logs" {
          for_each = logging_info.value.cloudwatch_logs != null ? [logging_info.value.cloudwatch_logs] : []
          content {
            enabled   = cloudwatch_logs.value.enabled
            log_group = cloudwatch_logs.value.log_group != "" ? cloudwatch_logs.value.log_group : null
          }
        }

        dynamic "firehose" {
          for_each = logging_info.value.firehose != null ? [logging_info.value.firehose] : []
          content {
            enabled         = firehose.value.enabled
            delivery_stream = firehose.value.delivery_stream != "" ? firehose.value.delivery_stream : null
          }
        }

        dynamic "s3" {
          for_each = logging_info.value.s3 != null ? [logging_info.value.s3] : []
          content {
            enabled = s3.value.enabled
            bucket  = s3.value.bucket != "" ? s3.value.bucket : null
            prefix  = s3.value.prefix != "" ? s3.value.prefix : null
          }
        }
      }
    }
  }

  dynamic "open_monitoring" {
    for_each = var.spec.jmx_exporter_enabled || var.spec.node_exporter_enabled ? [1] : []
    content {
      prometheus {
        jmx_exporter {
          enabled_in_broker = var.spec.jmx_exporter_enabled
        }
        node_exporter {
          enabled_in_broker = var.spec.node_exporter_enabled
        }
      }
    }
  }

  # Intelligent rebalancing is an Express-broker (express.* instance type)
  # capability; AWS rejects it on standard kafka.* clusters, so the block is
  # emitted only when the spec opts in.
  dynamic "rebalancing" {
    for_each = var.spec.rebalancing_status != "" ? [1] : []
    content {
      status = var.spec.rebalancing_status
    }
  }
}
