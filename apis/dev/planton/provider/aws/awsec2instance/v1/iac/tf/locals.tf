locals {
  # Safe dereferences and convenience flags
  safe_tags = coalesce(var.spec.tags, {})

  use_ssm             = try(var.spec.connection_method, "SSM") == "SSM"
  use_instance_connect = try(var.spec.connection_method, "SSM") == "INSTANCE_CONNECT"
  use_bastion         = try(var.spec.connection_method, "SSM") == "BASTION"

  needs_key_name = local.use_instance_connect || local.use_bastion

  # Convert instance profile ARN to name when provided
  iam_instance_profile_name = (
    local.use_ssm && try(var.spec.iam_instance_profile_arn.value, null) != null
  ) ? element(
    split("/", var.spec.iam_instance_profile_arn.value),
    length(split("/", var.spec.iam_instance_profile_arn.value)) - 1
  ) : null

  # Extract concrete values from StringValueOrRef lists and filter nulls
  security_group_ids_values = [
    for sg in try(var.spec.security_group_ids, []) : sg.value
    if try(sg.value, null) != null
  ]
}


