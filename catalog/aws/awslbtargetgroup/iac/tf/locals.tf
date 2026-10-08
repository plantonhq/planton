locals {
  # The AWS name: spec.target_group_name when set, else metadata.name. AWS
  # limits target group names to 32 characters; metadata.name is truncated
  # deterministically so the same manifest always yields the same name.
  target_group_name = var.spec.target_group_name != "" ? var.spec.target_group_name : substr(var.metadata.name, 0, 32)

  # A Lambda target group has no network identity: no port, no protocol, and
  # no VPC -- the load balancer invokes the function directly.
  is_lambda = var.spec.target_type == "lambda"

  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = local.target_group_name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsLbTargetGroup"
    "planton.ai/resource-id"   = var.metadata.id
  }
}
