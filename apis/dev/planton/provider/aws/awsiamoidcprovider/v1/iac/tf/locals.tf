locals {
  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsIamOidcProvider"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Pass thumbprints only when provided. An empty list is normalized to null so the
  # provider treats thumbprint_list as Computed and lets AWS derive it from its trusted
  # CA store -- this is the single explicit Pulumi/Terraform parity point for this module.
  thumbprint_list = length(var.spec.thumbprint_list) > 0 ? var.spec.thumbprint_list : null
}
