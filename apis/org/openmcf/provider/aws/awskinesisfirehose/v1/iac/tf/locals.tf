locals {
  # ---------------------------------------------------------------------------
  # Name & tags
  # ---------------------------------------------------------------------------

  delivery_stream_name = coalesce(try(var.metadata.name, null), "awskinesisfirehose")

  tags = merge({
    "Name" = local.delivery_stream_name
  }, try(var.metadata.labels, {}))

  # ---------------------------------------------------------------------------
  # Destination type (derived from which oneof field is populated)
  # ---------------------------------------------------------------------------
  # The proto uses a oneof for destination_config. In the JSON-decoded spec,
  # exactly one of the four destination fields is non-null. The destination
  # string is used by the aws_kinesis_firehose_delivery_stream resource.

  destination_type = (
    try(var.spec.extended_s3, null) != null ? "extended_s3" :
    try(var.spec.opensearch, null) != null ? "opensearch" :
    try(var.spec.http_endpoint, null) != null ? "http_endpoint" :
    try(var.spec.redshift, null) != null ? "redshift" :
    "extended_s3" # fallback — should never hit due to proto validation
  )

  # ---------------------------------------------------------------------------
  # Kinesis stream source (optional — Direct PUT when absent)
  # ---------------------------------------------------------------------------

  has_kinesis_source = try(var.spec.kinesis_stream_source, null) != null

  # ---------------------------------------------------------------------------
  # Server-side encryption (Direct PUT only)
  # ---------------------------------------------------------------------------
  # When a Kinesis source is configured, SSE must NOT be set — the source
  # stream handles its own encryption. Proto-level CEL validates this.

  sse_enabled     = try(var.spec.sse_enabled, false)
  sse_kms_key_arn = try(var.spec.sse_kms_key_arn.value, null)
  sse_key_type    = local.sse_kms_key_arn != null ? "CUSTOMER_MANAGED_CMK" : "AWS_OWNED_CMK"
}
