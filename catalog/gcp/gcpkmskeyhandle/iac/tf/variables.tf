variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpKmsKeyHandle specification"
  type = object({
    # The project of the resource the key will protect (where the handle
    # lives): a literal project ID or a GcpProject reference. Empty means the
    # provider's default project. Autokey must be on for this project or a
    # folder above it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The location of the resource the key will protect, e.g. "us-central1",
    # "europe-west4", or a multi-region such as "us". A CMEK key must sit in
    # the same location as its resource, so this must equal the resource's
    # location, and Autokey needs Cloud HSM there.
    location = string

    # The resource type the key is for, as Google names it:
    #   "storage.googleapis.com/Bucket", "compute.googleapis.com/Disk",
    #   "bigquery.googleapis.com/Dataset", "pubsub.googleapis.com/Topic",
    #   "sqladmin.googleapis.com/Instance", "secretmanager.googleapis.com/Secret",
    #   "artifactregistry.googleapis.com/Repository", "run.googleapis.com/Service",
    #   "spanner.googleapis.com/Database", "redis.googleapis.com/Instance", ...
    # Google's Autokey page lists every compatible type and the key
    # granularity (one key per resource, or per location for some types).
    resource_type_selector = string

    # The handle's ID, unique per project and location. Defaults to
    # metadata.name. A destroyed handle keeps its ID in Google, so a
    # recreated handle needs a new one.
    key_handle_name = optional(string, "")
  })
}
