# --- Managed Security Group ---

resource "aws_security_group" "cluster" {
  count = local.has_ingress_refs ? 1 : 0

  name        = var.metadata.id
  description = "Ingress for MSK cluster"
  vpc_id      = var.spec.vpc_id
  tags        = local.tags
}

# Kafka ports 9092-9098 from source security groups
resource "aws_security_group_rule" "ingress_kafka_sg" {
  for_each = local.has_ingress_refs ? toset(var.spec.security_group_ids) : toset([])

  type                     = "ingress"
  from_port                = 9092
  to_port                  = 9098
  protocol                 = "tcp"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.cluster[0].id
}

# ZooKeeper ports 2181-2182 from source security groups
resource "aws_security_group_rule" "ingress_zk_sg" {
  for_each = local.has_ingress_refs ? toset(var.spec.security_group_ids) : toset([])

  type                     = "ingress"
  from_port                = 2181
  to_port                  = 2182
  protocol                 = "tcp"
  source_security_group_id = each.value
  security_group_id        = aws_security_group.cluster[0].id
}

# Kafka ports from CIDR blocks
resource "aws_security_group_rule" "ingress_kafka_cidr" {
  count = local.has_ingress_refs && length(var.spec.allowed_cidr_blocks) > 0 ? 1 : 0

  type              = "ingress"
  from_port         = 9092
  to_port           = 9098
  protocol          = "tcp"
  cidr_blocks       = var.spec.allowed_cidr_blocks
  security_group_id = aws_security_group.cluster[0].id
}

# ZooKeeper ports from CIDR blocks
resource "aws_security_group_rule" "ingress_zk_cidr" {
  count = local.has_ingress_refs && length(var.spec.allowed_cidr_blocks) > 0 ? 1 : 0

  type              = "ingress"
  from_port         = 2181
  to_port           = 2182
  protocol          = "tcp"
  cidr_blocks       = var.spec.allowed_cidr_blocks
  security_group_id = aws_security_group.cluster[0].id
}

resource "aws_security_group_rule" "egress_all" {
  count = local.has_ingress_refs ? 1 : 0

  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.cluster[0].id
}

# --- Inline MSK Configuration ---

resource "aws_msk_configuration" "inline" {
  count = length(var.spec.server_properties) > 0 ? 1 : 0

  name              = "${var.metadata.id}-config"
  kafka_versions    = [var.spec.kafka_version]
  server_properties = local.server_properties
}

# --- MSK Cluster ---

resource "aws_msk_cluster" "this" {
  cluster_name           = var.metadata.id
  kafka_version          = var.spec.kafka_version
  number_of_broker_nodes = var.spec.number_of_broker_nodes
  enhanced_monitoring    = var.spec.enhanced_monitoring
  storage_mode           = var.spec.storage_mode != "" ? var.spec.storage_mode : null
  tags                   = local.tags

  broker_node_group_info {
    instance_type  = var.spec.instance_type
    client_subnets = var.spec.subnet_ids
    security_groups = concat(
      var.spec.associate_security_group_ids,
      local.has_ingress_refs ? [aws_security_group.cluster[0].id] : []
    )

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

    dynamic "connectivity_info" {
      for_each = var.spec.public_access_type != "" ? [1] : []
      content {
        public_access {
          type = var.spec.public_access_type
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
    for_each = length(var.spec.server_properties) > 0 ? [1] : (var.spec.configuration_arn != "" ? [1] : [])
    content {
      arn      = length(var.spec.server_properties) > 0 ? aws_msk_configuration.inline[0].arn : var.spec.configuration_arn
      revision = length(var.spec.server_properties) > 0 ? aws_msk_configuration.inline[0].latest_revision : var.spec.configuration_revision
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
            log_group = cloudwatch_logs.value.log_group
          }
        }

        dynamic "firehose" {
          for_each = logging_info.value.firehose != null ? [logging_info.value.firehose] : []
          content {
            enabled         = firehose.value.enabled
            delivery_stream = firehose.value.delivery_stream
          }
        }

        dynamic "s3_logs" {
          for_each = logging_info.value.s3 != null ? [logging_info.value.s3] : []
          content {
            enabled = s3_logs.value.enabled
            bucket  = s3_logs.value.bucket
            prefix  = s3_logs.value.prefix
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
}
