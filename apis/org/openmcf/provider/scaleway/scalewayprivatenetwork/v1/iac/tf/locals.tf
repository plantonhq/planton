locals {
  # Private Network name from metadata
  private_network_name = var.metadata.name

  # VPC ID from spec (resolved from StringValueOrRef before Terraform runs)
  vpc_id = var.spec.vpc_id

  # Region from spec
  region = var.spec.region

  # Subnet configuration
  ipv4_subnet  = var.spec.ipv4_subnet != null ? var.spec.ipv4_subnet : ""
  ipv6_subnets = var.spec.ipv6_subnets != null ? var.spec.ipv6_subnets : []

  # Route propagation
  enable_default_route_propagation = var.spec.enable_default_route_propagation

  # Standard OpenMCF tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayPrivateNetwork",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
