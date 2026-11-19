# Enable the Compute Engine API so a fresh project can host the subnetwork.
# disable_on_destroy is false: tearing down one subnet must never disable the
# API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A subnetwork in a custom-mode VPC — the regional address space workloads
# live in: primary IPv4 range for VM interfaces, secondary ranges for alias
# IPs (GKE pods/services), optional IPv6, and VPC Flow Logs.
#
# name, project, region, network, and description are immutable (ForceNew):
# changing any of them destroys and recreates the subnet — an outage for
# everything addressed in it. The primary range is the one asymmetric knob:
# EXPANDING ip_cidr_range updates in place, shrinking recreates (the provider
# guards this with a shrink-only ForceNew diff).
#
# purpose creates the special-role subnets other features depend on:
# REGIONAL_MANAGED_PROXY reserves Envoy address space for the region's
# regional load balancers (required before a regional ALB exists in the VPC);
# PRIVATE_SERVICE_CONNECT backs published PSC services.
resource "google_compute_subnetwork" "main" {
  name          = var.spec.subnetwork_name
  project       = local.project_id
  region        = var.spec.region
  network       = var.spec.vpc_self_link
  description   = var.spec.description
  ip_cidr_range = local.ip_cidr_range

  # Alternative primary-CIDR source: a centrally allocated Network
  # Connectivity internal range (exactly one of the two is set,
  # spec-enforced).
  reserved_internal_range = var.spec.reserved_internal_range != "" ? var.spec.reserved_internal_range : null

  purpose = local.purpose
  role    = local.role

  private_ip_google_access   = local.private_ip_google_access
  private_ipv6_google_access = local.private_ipv6_google_access

  # Dual-stack wiring: stack_type opts into IPv6; ipv6_access_type decides
  # whether the assigned prefix is internet-routable (EXTERNAL GUAs) or
  # VPC-internal (INTERNAL ULAs — requires the VPC's ULA range enabled).
  stack_type           = local.stack_type
  ipv6_access_type     = local.ipv6_access_type
  external_ipv6_prefix = local.external_ipv6_prefix

  # INTERNAL counterpart of external_ipv6_prefix: pin a ULA prefix instead
  # of letting Google allocate one from the VPC's internal IPv6 range.
  internal_ipv6_prefix = var.spec.internal_ipv6_prefix != "" ? var.spec.internal_ipv6_prefix : null

  # BYOIP: draw the subnet's IPv6 space from a PublicDelegatedPrefix.
  ip_collection = var.spec.ip_collection != "" ? var.spec.ip_collection : null

  # ARP subnet-mask resolution for appliance/NFV subnets.
  resolve_subnet_mask = var.spec.resolve_subnet_mask != "" ? var.spec.resolve_subnet_mask : null

  # Deliberate address-space reclaims only: subnet routes still win over the
  # overlapping peer/on-prem routes this permits.
  # Google refuses the field on a special-purpose subnet (proxy-only, PSC
  # NAT, ...): "Invalid field set in subnetwork with purpose
  # REGIONAL_MANAGED_PROXY". It is sent only on an ordinary PRIVATE subnet.
  allow_subnet_cidr_routes_overlap = local.purpose == "PRIVATE" ? coalesce(var.spec.allow_subnet_cidr_routes_overlap, false) : null

  # Safety latch: by default an empty secondary-range list is NOT sent on
  # update, so a partial manifest cannot silently wipe GKE pod ranges.
  send_secondary_ip_range_if_empty = coalesce(var.spec.send_secondary_ip_range_if_empty, false)

  # Secondary (alias) ranges — the mechanism GKE uses for pod/service IPs.
  # Consumers select a range by its name (e.g. ip_allocation_policy). Each
  # range's CIDR comes from exactly one of a literal CIDR or a reserved
  # internal range (spec-enforced).
  dynamic "secondary_ip_range" {
    for_each = local.secondary_ip_ranges
    content {
      range_name              = secondary_ip_range.value.range_name
      ip_cidr_range           = secondary_ip_range.value.ip_cidr_range != "" ? secondary_ip_range.value.ip_cidr_range : null
      reserved_internal_range = secondary_ip_range.value.reserved_internal_range != "" ? secondary_ip_range.value.reserved_internal_range : null
    }
  }

  # Create-time Resource Manager tag bindings; the params block is
  # create-only, so tag changes replace the subnetwork.
  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.resource_manager_tags
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # VPC Flow Logs: presence of the block enables logging. Defaults mirror
  # the API's own (5s aggregation, 50% sampling, all metadata) so an empty
  # spec object behaves sanely; metadata_fields only accompanies
  # CUSTOM_METADATA (spec-enforced).
  dynamic "log_config" {
    for_each = local.log_config != null ? [local.log_config] : []
    content {
      aggregation_interval = log_config.value.aggregation_interval
      flow_sampling        = log_config.value.flow_sampling
      metadata             = log_config.value.metadata
      metadata_fields      = log_config.value.metadata == "CUSTOM_METADATA" ? log_config.value.metadata_fields : null
      filter_expr          = log_config.value.filter_expr
    }
  }

  # Depend on API enablement so a fresh project works first try.
  depends_on = [google_project_service.compute_api]
}
