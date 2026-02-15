locals {
  # Consumer name derived from metadata.name.
  consumer_name = coalesce(try(var.metadata.name, null), "awskinesisstreamconsumer")

  # Parent stream ARN — required input.
  stream_arn = try(var.spec.stream_arn.value, var.spec.stream_arn)

  # Tags
  tags = merge({
    "Name" = local.consumer_name
  }, try(var.metadata.labels, {}))
}
