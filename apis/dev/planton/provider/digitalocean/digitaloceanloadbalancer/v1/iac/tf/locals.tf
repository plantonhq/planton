# Local variables for computed values and helper functions

locals {
  # Extract VPC UUID from value or ref (assuming value is provided directly for simplicity)
  vpc_uuid = var.spec.vpc.value != null ? var.spec.vpc.value : ""

  # Extract droplet IDs from the list of value/ref objects
  droplet_ids = var.spec.droplet_ids != null ? [
    for d in var.spec.droplet_ids : d.value != null ? tonumber(d.value) : 0
  ] : []

  # Filter out any zero values (invalid droplet IDs)
  valid_droplet_ids = [for id in local.droplet_ids : id if id != 0]

  # Protocol mapping from protobuf enum string to DigitalOcean API format
  # Expected input: "http", "https", "tcp" (lowercase enum names)
  # Output: Same format (DigitalOcean uses lowercase)
  
  # Convert forwarding rules from spec format to Terraform resource format
  forwarding_rules = [
    for rule in var.spec.forwarding_rules : {
      entry_port       = rule.entry_port
      entry_protocol   = lower(rule.entry_protocol)
      target_port      = rule.target_port
      target_protocol  = lower(rule.target_protocol)
      certificate_name = rule.certificate_name
    }
  ]

  # Health check configuration
  healthcheck = var.spec.health_check != null ? {
    port                   = var.spec.health_check.port
    protocol               = lower(var.spec.health_check.protocol)
    path                   = var.spec.health_check.path
    check_interval_seconds = var.spec.health_check.check_interval_sec != null ? var.spec.health_check.check_interval_sec : 10
  } : null

  # Sticky sessions configuration
  sticky_sessions = var.spec.enable_sticky_sessions ? {
    type = "cookies"
  } : null

  # Common tags/labels for the load balancer
  # DigitalOcean doesn't support tags on load balancers directly,
  # but we can use them in terraform for organization
  resource_labels = merge(
    {
      "managed-by" = "terraform"
      "resource"   = "digitalocean-load-balancer"
    },
    var.metadata.labels != null ? var.metadata.labels : {}
  )
}

