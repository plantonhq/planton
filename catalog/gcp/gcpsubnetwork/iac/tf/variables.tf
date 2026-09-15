variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP subnetwork"
  type = object({
    # The GCP project that owns the subnetwork. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Self-link of the parent VPC network (resolved from a GcpVpcNetwork reference
    # or given directly). Immutable.
    vpc_self_link = string

    # Name of the subnetwork in GCP (RFC1035). Immutable.
    subnetwork_name = string

    # Region the subnetwork lives in. Immutable.
    region = string

    # Primary IPv4 CIDR range. Expandable in place; shrinking recreates.
    # Empty only for IPV6_ONLY subnets.
    ip_cidr_range = optional(string, "")

    # What this subnet carries — for the operator doing IP planning later.
    description = optional(string)

    # PRIVATE (default) | REGIONAL_MANAGED_PROXY | GLOBAL_MANAGED_PROXY |
    # PRIVATE_SERVICE_CONNECT | PEER_MIGRATION | PRIVATE_NAT.
    purpose = optional(string)

    # ACTIVE | BACKUP — only for REGIONAL_MANAGED_PROXY subnets.
    role = optional(string, "")

    # Secondary IPv4 ranges for alias IPs (GKE pods/services). Each range's
    # CIDR comes from exactly one of ip_cidr_range or reserved_internal_range
    # (spec-enforced).
    secondary_ip_ranges = optional(list(object({
      range_name              = string
      ip_cidr_range           = optional(string, "")
      reserved_internal_range = optional(string, "")
    })))

    # Let VMs without external IPs reach Google APIs internally.
    private_ip_google_access = optional(bool)

    # IPv6 counterpart of private_ip_google_access.
    private_ipv6_google_access = optional(string, "")

    # IPV4_ONLY (default) | IPV4_IPV6 | IPV6_ONLY.
    stack_type = optional(string)

    # EXTERNAL | INTERNAL — required for IPv6-carrying stack types.
    ipv6_access_type = optional(string, "")

    # Pin a specific external /64 IPv6 prefix (EXTERNAL access type only).
    external_ipv6_prefix = optional(string, "")

    # Permit the subnet CIDR to overlap routes to destinations outside the
    # VPC (deliberate address-space reclaims only).
    allow_subnet_cidr_routes_overlap = optional(bool)

    # true: an empty secondary_ip_ranges list REMOVES existing secondary
    # ranges on update; false (default): an empty list leaves them untouched.
    send_secondary_ip_range_if_empty = optional(bool)

    # VPC Flow Logs; presence of the object enables logging. Every field has
    # a GCP default, so an empty object is valid.
    log_config = optional(object({
      enabled              = optional(bool)
      aggregation_interval = optional(string)
      flow_sampling        = optional(number)
      metadata             = optional(string)
      metadata_fields      = optional(list(string), [])
      filter_expr          = optional(string, "")
    }))

    # Source the primary CIDR from a Network Connectivity internal range
    # (alternative to ip_cidr_range). Immutable.
    reserved_internal_range = optional(string, "")

    # Pin a specific internal IPv6 prefix (INTERNAL access type only).
    # Immutable.
    internal_ipv6_prefix = optional(string, "")

    # BYOIP: PublicDelegatedPrefix the subnet's IPv6 space is drawn from
    # (dual-stack / IPv6-only subnets only).
    ip_collection = optional(string, "")

    # Resource Manager tags bound at create time (tagKeys/... ->
    # tagValues/...). Changing them replaces the subnetwork.
    resource_manager_tags = optional(map(string), {})

    # ARP subnet-mask resolution mode for appliance/NFV subnets. Immutable.
    resolve_subnet_mask = optional(string, "")

    # What happens to the subnetwork in GCP on destroy:
    # DELETE (provider default), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")
  })

  validation {
    condition     = can(regex("^[a-z]([-a-z0-9]*[a-z0-9])?$", var.spec.region))
    error_message = "region must be a valid GCP region name (lowercase letters, numbers, and hyphens)."
  }

  validation {
    condition     = var.spec.ip_cidr_range == "" || can(regex("^\\d+\\.\\d+\\.\\d+\\.\\d+/\\d+$", var.spec.ip_cidr_range))
    error_message = "ip_cidr_range must be an IPv4 CIDR like 10.10.0.0/20 (empty only for IPV6_ONLY subnets)."
  }

  # HCL's || does not short-circuit — nullable optionals are guarded with
  # coalesce/try so a null stack_type cannot crash validation.
  validation {
    condition     = coalesce(var.spec.stack_type, "IPV4_ONLY") == "IPV6_ONLY" ? var.spec.ip_cidr_range == "" : (var.spec.ip_cidr_range != "" || var.spec.reserved_internal_range != "")
    error_message = "set ip_cidr_range or reserved_internal_range (only IPV6_ONLY subnets omit both)."
  }

  validation {
    condition     = can(regex("^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$", var.spec.subnetwork_name))
    error_message = "subnetwork_name must be 1-63 characters, lowercase letters, numbers, or hyphens, starting with a letter and ending with a letter or number."
  }

  validation {
    condition     = contains(["PRIVATE", "REGIONAL_MANAGED_PROXY", "GLOBAL_MANAGED_PROXY", "PRIVATE_SERVICE_CONNECT", "PEER_MIGRATION", "PRIVATE_NAT"], coalesce(var.spec.purpose, "PRIVATE"))
    error_message = "purpose must be one of PRIVATE, REGIONAL_MANAGED_PROXY, GLOBAL_MANAGED_PROXY, PRIVATE_SERVICE_CONNECT, PEER_MIGRATION, or PRIVATE_NAT."
  }

  validation {
    condition     = contains(["", "ACTIVE", "BACKUP"], var.spec.role)
    error_message = "role must be ACTIVE or BACKUP."
  }

  validation {
    condition     = var.spec.role == "" || coalesce(var.spec.purpose, "PRIVATE") == "REGIONAL_MANAGED_PROXY"
    error_message = "role is only valid on REGIONAL_MANAGED_PROXY subnets."
  }

  validation {
    condition     = contains(["IPV4_ONLY", "IPV4_IPV6", "IPV6_ONLY"], coalesce(var.spec.stack_type, "IPV4_ONLY"))
    error_message = "stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY."
  }

  validation {
    condition     = coalesce(var.spec.stack_type, "IPV4_ONLY") == "IPV4_ONLY" ? var.spec.ipv6_access_type == "" : contains(["EXTERNAL", "INTERNAL"], var.spec.ipv6_access_type)
    error_message = "ipv6_access_type (EXTERNAL or INTERNAL) is required for IPV4_IPV6 and IPV6_ONLY subnets, and must be omitted for IPV4_ONLY."
  }

  validation {
    condition     = contains(["", "DISABLE_GOOGLE_ACCESS", "ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE", "ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE"], var.spec.private_ipv6_google_access)
    error_message = "private_ipv6_google_access must be DISABLE_GOOGLE_ACCESS, ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE, or ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE."
  }
}
