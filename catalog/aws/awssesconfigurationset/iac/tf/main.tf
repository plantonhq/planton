# Amazon SES (SESv2) configuration set.
#
# The set itself carries the sending posture (TLS, IP pool, tracking,
# suppression, VDM, reputation and sending switches); each event
# destination is a separate AWS sub-resource keyed by (set, name) and is
# materialized per-name below so destinations can be added and removed
# without touching the set.
resource "aws_sesv2_configuration_set" "this" {
  # The cloud name comes from metadata.name (the catalog naming basis) --
  # set explicitly so both engines create the same configuration set.
  configuration_set_name = var.metadata.name

  # Transport controls. The block is emitted only when the manifest sets
  # it -- an absent block keeps AWS's own defaults (OPTIONAL TLS, 14-hour
  # retry, shared IP space) without materializing them into state.
  dynamic "delivery_options" {
    for_each = var.spec.delivery_options != null ? [var.spec.delivery_options] : []
    content {
      # tls_policy is contract-defaulted to OPTIONAL (AWS's own default),
      # so a set manifest never flips TLS posture implicitly.
      tls_policy           = coalesce(delivery_options.value.tls_policy, "OPTIONAL")
      max_delivery_seconds = delivery_options.value.max_delivery_seconds
      sending_pool_name    = delivery_options.value.sending_pool_name != "" ? delivery_options.value.sending_pool_name : null
    }
  }

  # Reputation metrics: a plain switch AWS defaults to off. Emitted
  # unconditionally so an explicit false in the manifest converges a set
  # that was hand-enabled in the console.
  reputation_options {
    reputation_metrics_enabled = var.spec.reputation_metrics_enabled
  }

  # The per-set sending kill switch (default true).
  sending_options {
    sending_enabled = local.sending_enabled
  }

  # Suppression override: an ABSENT block inherits the account-level
  # suppression configuration, which is different from an explicit empty
  # list (suppress nothing) -- so the block is emitted only when the
  # manifest lists reasons.
  dynamic "suppression_options" {
    for_each = length(var.spec.suppressed_reasons) > 0 ? [1] : []
    content {
      suppressed_reasons = var.spec.suppressed_reasons
    }
  }

  # Custom open/click tracking domain.
  dynamic "tracking_options" {
    for_each = var.spec.tracking_options != null ? [var.spec.tracking_options] : []
    content {
      custom_redirect_domain = tracking_options.value.custom_redirect_domain
      https_policy           = coalesce(tracking_options.value.https_policy, "OPTIONAL")
    }
  }

  # Virtual Deliverability Manager overrides. The provider models the two
  # dials as nested single-field blocks with ENABLED/DISABLED strings; the
  # spec's booleans map onto them here.
  dynamic "vdm_options" {
    for_each = var.spec.vdm_options != null ? [var.spec.vdm_options] : []
    content {
      dashboard_options {
        engagement_metrics = coalesce(vdm_options.value.engagement_metrics_enabled, false) ? "ENABLED" : "DISABLED"
      }
      guardian_options {
        optimized_shared_delivery = coalesce(vdm_options.value.optimized_shared_delivery_enabled, false) ? "ENABLED" : "DISABLED"
      }
    }
  }

  tags = local.aws_tags
}

# Event destinations -- one AWS sub-resource per named entry. Exactly one
# destination arm is set per entry (spec-level CEL); each arm maps to the
# provider's corresponding nested block.
resource "aws_sesv2_configuration_set_event_destination" "this" {
  for_each = local.event_destinations

  configuration_set_name = aws_sesv2_configuration_set.this.configuration_set_name
  event_destination_name = each.key

  event_destination {
    # AWS defaults `enabled` to FALSE, which silently publishes nothing --
    # the catalog defaults it to true and always sends the value
    # explicitly so a created destination actually delivers events.
    enabled              = coalesce(each.value.enabled, true)
    matching_event_types = each.value.matching_event_types

    dynamic "cloud_watch_destination" {
      for_each = each.value.cloud_watch != null ? [each.value.cloud_watch] : []
      content {
        dynamic "dimension_configuration" {
          for_each = cloud_watch_destination.value.dimensions
          content {
            dimension_name          = dimension_configuration.value.name
            dimension_value_source  = dimension_configuration.value.value_source
            default_dimension_value = dimension_configuration.value.default_value
          }
        }
      }
    }

    # The ref-flattened string arms carry the contract default "" when
    # absent -- plain != "" comparisons, never coalesce(x, ""), which
    # errors in HCL when every argument is empty.
    dynamic "event_bridge_destination" {
      for_each = each.value.event_bus != "" ? [each.value.event_bus] : []
      content {
        event_bus_arn = event_bridge_destination.value
      }
    }

    dynamic "kinesis_firehose_destination" {
      for_each = each.value.firehose != null ? [each.value.firehose] : []
      content {
        delivery_stream_arn = kinesis_firehose_destination.value.delivery_stream
        iam_role_arn        = kinesis_firehose_destination.value.iam_role
      }
    }

    dynamic "sns_destination" {
      for_each = each.value.sns_topic != "" ? [each.value.sns_topic] : []
      content {
        topic_arn = sns_destination.value
      }
    }

    dynamic "pinpoint_destination" {
      for_each = each.value.pinpoint_application_arn != "" ? [each.value.pinpoint_application_arn] : []
      content {
        application_arn = pinpoint_destination.value
      }
    }
  }
}
