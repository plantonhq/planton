locals {
  name = var.metadata.name

  tags = {
    "openmcf.org/resource"      = "true"
    "openmcf.org/organization"  = var.metadata.org
    "openmcf.org/environment"   = var.metadata.env
    "openmcf.org/resource-kind" = "AwsTransitGateway"
    "openmcf.org/resource-id"   = var.metadata.id
  }

  enable_disable = {
    true  = "enable"
    false = "disable"
  }

  vpc_attachments_map = {
    for att in var.spec.vpc_attachments : att.name => att
  }
}
