locals {
  # Resource-identity tags match the Pulumi module key-for-key. CloudFront
  # distributions have no AWS name -- metadata.name drives the Name tag and
  # consumers address the distribution through its ID/ARN/domain outputs.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsCloudFront"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Origins keyed by their user-chosen ID. CEL guarantees uniqueness across
  # origins and origin groups, so these maps never collide.
  origins_by_id = { for o in var.spec.origins : o.origin_id => o }

  # Origins that asked the module to create an Origin Access Control -- the
  # modern private-S3 shape. Each gets its own OAC resource (SigV4 request
  # signing); the bucket policy must allow the distribution's ARN.
  oac_origins = {
    for id, o in local.origins_by_id : id => o
    if try(o.s3_origin.create_origin_access_control, false)
  }

  # The behavior field set is identical for the default and ordered
  # behaviors, so both render through the same expressions. Empty method
  # lists keep CloudFront's static-content defaults.
  default_behavior = var.spec.default_cache_behavior

  # The generator flattens StringValueOrRef to its resolved string (the
  # orchestrator resolves any value_from before the module runs).
  web_acl_arn = var.spec.web_acl_arn

  # Viewer certificate arms. Absent block (or neither arm set) serves the
  # default *.cloudfront.net certificate; CEL blocks aliases in that shape.
  viewer_cert = var.spec.viewer_certificate
  has_custom_viewer_cert = (
    try(var.spec.viewer_certificate.acm_certificate_arn, "") != "" ||
    try(var.spec.viewer_certificate.iam_certificate_id, "") != ""
  )
}
