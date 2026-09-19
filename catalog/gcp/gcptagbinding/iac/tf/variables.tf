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
  description = "GcpTagBinding specification"
  type = object({
    # The tag value to bind: a reference to a GcpTagValue resource (its
    # `name` output, `tagValues/{id}`) or, as a literal, that name or the
    # namespaced form `{org_id}/{key_short_name}/{value_short_name}`.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    tag_value = string

    # The resource that carries the tag. At most one arm; all empty means
    # the provider's default project (the project the deploying credentials
    # are configured for) -- the common "tag the project I am deploying
    # into" needs no parent at all. Immutable.
    parent = optional(object({
      # A project: a reference to a GcpProject resource (its project_number
      # output) or, as a literal, the project's NUMBER or ID. Google requires
      # the number in a binding's parent; when a literal is not numeric the
      # module looks the number up once at apply time (a read of the project,
      # which the deploying identity can already see).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # A folder: the folder's numeric ID -- a literal, or a reference to a
      # GcpFolder resource (its folder_id output). Every project and folder
      # beneath it inherits the tag.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")

      # The organization: the numeric organization ID, without the
      # `organizations/` prefix. Everything in the estate inherits the tag.
      organization_id = optional(string, "")

      # Any other taggable resource, by its full resource name -- the
      # `//{service}.googleapis.com/...` form Google's "full resource name"
      # convention defines. Common shapes:
      #   //compute.googleapis.com/projects/{project}/zones/{zone}/instances/{name}
      #     (set location to the zone)
      #   //compute.googleapis.com/projects/{project}/global/networks/{name}
      #   //sqladmin.googleapis.com/projects/{project}/instances/{name}
      #     (set location to the region)
      #   //container.googleapis.com/projects/{project}/locations/{location}/clusters/{name}
      #     (set location to the cluster's location)
      #   //storage.googleapis.com/projects/_/buckets/{name}
      #   //bigquery.googleapis.com/projects/{project}/datasets/{name}
      # Set `location` for every regional or zonal resource.
      resource_name = optional(string, "")
    }))

    # For a REGIONAL or ZONAL resource named in parent.resource_name: its
    # region (`us-central1`) or zone (`us-central1-a`). Required for such
    # resources -- Google serves their bindings from a regional endpoint --
    # and must stay empty for organizations, folders, projects, and other
    # global resources. Immutable.
    location = optional(string, "")

    # What destroying this resource does to the binding in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the binding is removed; the resource no longer carries
    #                the tag (immediate, no recovery window)
    #   "PREVENT" -- destroy FAILS; the guard for a tag a policy depends on
    #   "ABANDON" -- the binding is removed from management but the resource
    #                keeps the tag
    deletion_policy = optional(string, "")
  })
}
