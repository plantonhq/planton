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
  description = "GcpSharedVpcHost specification"
  type = object({
    # The project that becomes the Shared VPC host: a reference to a
    # GcpProject (its project_id output) or the project ID as a literal.
    # Empty means the provider's default project -- the project the
    # credentials are configured for -- so enabling the project you are
    # deploying into needs no configuration at all. Immutable: to host from
    # a different project, destroy this and declare a new one.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # What destroying this resource does to the host role in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the project's host status is disabled; fails while any
    #                service project is still attached (destroy the
    #                GcpSharedVpcServiceProject resources first -- a chart's
    #                dependency order does this when they reference the host)
    #   "PREVENT" -- destroy FAILS; the guard for the host every service
    #                project in the organization depends on
    #   "ABANDON" -- the resource leaves management but the project stays a
    #                host with every attachment intact
    deletion_policy = optional(string, "")
  })
}
