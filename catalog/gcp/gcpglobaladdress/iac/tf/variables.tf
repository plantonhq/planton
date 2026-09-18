variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpGlobalAddress specification"
  type = object({
    # The GCP project in which to create this global address reservation.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Example: "my-prod-project-123"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the global address resource in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a letter or number.
    # Example: "lb-external-ip", "vpc-peering-range"
    address_name = string

    # The static IP address to reserve. If omitted, GCP assigns an address automatically.
    # For EXTERNAL addresses this is a single IP. For INTERNAL VPC_PEERING addresses
    # this is the start of the reserved CIDR range.
    address = optional(string, "")

    # The type of address to reserve.
    # EXTERNAL reserves a public IP address (default). INTERNAL reserves a private IP range
    # within a VPC network for purposes like VPC peering or Private Service Connect.
    address_type = optional(string)

    # Human-readable description of this address reservation.
    # Example: "Static IP for production HTTPS load balancer"
    description = optional(string, "")

    # The IP version for this address. Defaults to IPV4.
    ip_version = optional(string)

    # The VPC network for internal address reservations.
    # Required when address_type is INTERNAL. Accepts a network name or full self-link URL.
    # The reserved IP range must be in RFC1918 space and the network cannot be deleted
    # while reserved IP ranges refer to it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The prefix length of the IP range to reserve.
    # Required for VPC_PEERING purpose (e.g., 20 reserves a /20 range).
    # Not applicable for single IP reservations or PRIVATE_SERVICE_CONNECT addresses.
    # Valid range: 8 to 29.
    prefix_length = optional(number)

    # The purpose of this address reservation. Only applicable for INTERNAL addresses,
    # and REQUIRED for them — the GCP API rejects an internal global reservation
    # without a purpose ("The field must be specified for reserving internal IP
    # Addresses").
    # VPC_PEERING — reserves a CIDR range for VPC network peering. Used by managed services
    # like Cloud SQL, Redis, AlloyDB, and Filestore for private networking.
    # PRIVATE_SERVICE_CONNECT — reserves an address for a Private Service Connect endpoint.
    # Leave empty for standard external address reservations.
    purpose = optional(string, "")

    # User labels attached to the reserved global address, merged with
    # Planton's platform labels (which win on key conflicts). The one mutable
    # surface on this resource — every other change destroys and re-reserves.
    labels = optional(map(string), {})

    # Deletion policy for the reserved global address — what happens on
    # destroy:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the reservation is released (GCP refuses while a global
    #                forwarding rule or PSA range still uses it); a released
    #                external anycast IP is gone for good
    #   "PREVENT" -- destroy FAILS; protects an IP that external DNS and
    #                client allow-lists may still point at
    #   "ABANDON" -- the reservation is removed from management but stays
    #                reserved in GCP
    deletion_policy = optional(string, "")
  })
}
