locals {
  name = var.metadata.name

  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsTransitGateway"
    "planton.dev/resource-id"   = var.metadata.id
  }

  enable_disable = {
    true  = "enable"
    false = "disable"
  }

  vpc_attachments_map = {
    for att in var.spec.vpc_attachments : att.name => att
  }
}
