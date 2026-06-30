# NOTE: dead_letter_config and log_config are supported in the Pulumi module
# but require AWS provider >= 6.x for Terraform. These blocks are omitted here
# to maintain provider version consistency (5.82.0) across all Planton TF modules.
# The Pulumi module provides full feature support.

resource "aws_cloudwatch_event_bus" "this" {
  name = local.resource_name

  event_source_name = local.event_source_name

  tags = local.tags
}
