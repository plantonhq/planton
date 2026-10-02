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
  description = "GcpSharedVpcServiceProject specification"
  type = object({
    # The Shared VPC host to attach to: a reference to a GcpSharedVpcHost
    # (its host_project_id output) or the host's project ID as a literal.
    # Referencing the HOST KIND, not the project, is what orders a chart
    # correctly -- the project is enabled as a host before anything attaches
    # to it. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    host_project_id = string

    # The project being attached as a service project: a reference to a
    # GcpProject (its project_id output) or the project ID as a literal.
    # Required -- Google's API takes both projects explicitly and the
    # provider has no ambient default for this one. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_project_id = string

    # What destroying this resource does to the attachment in GCP. This
    # resource's own provider argument accepts ONE value, not the usual
    # DELETE/PREVENT/ABANDON trio:
    #   ""        -- the attachment is removed: the service project is
    #                detached from the host (the provider default). Fails
    #                while any resource in the service project still uses a
    #                host subnetwork.
    #   "ABANDON" -- the resource leaves management but the project stays
    #                attached with every workload intact -- for a service
    #                project whose VMs must outlive the chart that attached
    #                it (a shared platform project handed to another team).
    deletion_policy = optional(string, "")
  })
}
