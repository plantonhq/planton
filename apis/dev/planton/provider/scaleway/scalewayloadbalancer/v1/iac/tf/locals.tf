locals {
  # Resource identity
  lb_name = var.metadata.name
  zone    = var.spec.zone

  # Spec fields
  lb_type                 = var.spec.type
  description             = var.spec.description
  ssl_compatibility_level = var.spec.ssl_compatibility_level
  private_network_id      = var.spec.private_network_id

  # Indexed maps for for_each iteration
  backends_map     = { for b in var.spec.backends : b.name => b }
  frontends_map    = { for f in var.spec.frontends : f.name => f }
  certificates_map = { for c in var.spec.certificates : c.name => c }

  # Standard Planton tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayLoadBalancer",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
