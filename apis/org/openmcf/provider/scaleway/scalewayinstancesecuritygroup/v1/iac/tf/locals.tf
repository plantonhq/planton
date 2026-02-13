locals {
  # Security group name from metadata
  sg_name = var.metadata.name

  # Zone from spec
  zone = var.spec.zone

  # Description (empty string means Terraform will use the default)
  description = var.spec.description

  # Stateful flag
  stateful = var.spec.stateful

  # Default policies
  inbound_default_policy  = var.spec.inbound_default_policy
  outbound_default_policy = var.spec.outbound_default_policy

  # SMTP security
  enable_default_security = var.spec.enable_default_security

  # Rules from spec
  inbound_rules  = var.spec.inbound_rules
  outbound_rules = var.spec.outbound_rules

  # Standard OpenMCF tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayInstanceSecurityGroup",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
