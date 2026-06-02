locals {
  # Name and tags
  resource_name = coalesce(try(var.metadata.name, null), "awssnstopic")
  is_fifo       = coalesce(try(var.spec.fifo_topic, null), false)

  # FIFO topics must have names ending with ".fifo".
  topic_name = local.is_fifo && !endswith(local.resource_name, ".fifo") ? "${local.resource_name}.fifo" : local.resource_name

  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Encryption
  kms_master_key_id = try(var.spec.kms_key_id.value, null)

  # Access policy is a free-form JSON object (google.protobuf.Struct); aws_sns_topic wants
  # policy as a JSON string, so encode the object here.
  policy = try(jsonencode(var.spec.policy), null)

  # Delivery policy
  delivery_policy = try(var.spec.delivery_policy, null) != "" ? try(var.spec.delivery_policy, null) : null

  # Observability
  tracing_config    = try(var.spec.tracing_config, null) != "" ? try(var.spec.tracing_config, null) : null
  signature_version = try(var.spec.signature_version, null) != 0 ? try(var.spec.signature_version, null) : null

  # Subscriptions — build a map keyed by subscription name for for_each.
  subscriptions_list = coalesce(try(var.spec.subscriptions, null), [])
  subscriptions_map = {
    for sub in local.subscriptions_list : sub.name => sub
  }

  # Subscription redrive policies — build per-subscription redrive JSON or null.
  subscription_redrive_policies = {
    for sub in local.subscriptions_list : sub.name => (
      try(sub.redrive_config, null) != null ? jsonencode({
        deadLetterTargetArn = try(sub.redrive_config.dead_letter_target_arn.value, "")
      }) : null
    )
  }
}
