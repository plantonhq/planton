locals {
  # Resource identity
  gateway_name = var.metadata.name

  # Spec fields
  zone               = var.spec.zone
  gateway_type       = var.spec.type
  private_network_id = var.spec.private_network_id
  enable_masquerade  = var.spec.enable_masquerade
  enable_smtp        = var.spec.enable_smtp
  reverse_dns        = var.spec.reverse_dns
  pat_rules          = var.spec.pat_rules

  # Bastion configuration (safely unwrap optional nested object)
  bastion_enabled     = var.spec.bastion != null ? var.spec.bastion.enabled : false
  bastion_port        = var.spec.bastion != null ? var.spec.bastion.port : 22
  allowed_ip_ranges   = var.spec.bastion != null ? var.spec.bastion.allowed_ip_ranges : []

  # Standard OpenMCF tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayPublicGateway",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
