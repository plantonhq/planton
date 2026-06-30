locals {
  # VPC name from metadata
  vpc_name = var.metadata.name

  # Region from spec
  region = var.spec.region

  # Routing flags
  enable_routing                    = var.spec.enable_routing
  enable_custom_routes_propagation  = var.spec.enable_custom_routes_propagation

  # Standard Planton tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayVpc",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
