locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsSagemakerDomain"
    "planton.org/resource-id"   = var.metadata.id
  }

  has_domain_settings = (
    length(var.spec.domain_security_group_ids) > 0 ||
    var.spec.docker_settings != null
  )
}
