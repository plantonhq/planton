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
  description = "GcpCloudBuildRepository specification"
  type = object({
    # The connection the repository is linked through, by full name
    # (projects/{project}/locations/{location}/connections/{connection}): a
    # GcpCloudBuildConnection reference (its name output) or the literal
    # name. Google's provider parses the project and location from this full
    # form, so a short name is refused. Required. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    parent_connection = string

    # The repository's ID in Cloud Build, unique in the connection: letters,
    # digits, and any of -._~%!$&'()*+,;=@ (Google's rule). Usually the
    # repository's name on the host. Defaults to metadata.name. Immutable.
    repository_id = optional(string, "")

    # The repository's HTTPS clone URI on the code host, e.g.
    # "https://github.com/acme/orders.git". Required. Immutable.
    remote_uri = string

    # Annotations on the repository link (AIP-128 key/value metadata; a
    # repository has no labels). Immutable: a change replaces the link.
    annotations = optional(map(string), {})

    # What destroy does:
    #   "" / "DELETE" -- the link is removed from Cloud Build
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the link leaves management and stays in Cloud Build
    deletion_policy = optional(string, "")
  })
}
