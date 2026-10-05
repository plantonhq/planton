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
  description = "GcpGkeFleetMembership specification"
  type = object({
    # The fleet host project the membership lives in: a literal project ID
    # or a GcpGkeFleet reference (the fleet this cluster joins, which orders
    # the membership after the fleet in a chart). Empty means the provider's
    # default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The membership's ID, unique in the fleet; by convention the cluster's
    # name. 1-63 lowercase letters, digits, and hyphens, starting and ending
    # with a letter or digit. Defaults to metadata.name. Immutable.
    membership_id = optional(string, "")

    # Where the membership lives: "global" (the default, and what explicit
    # registration normally uses) or a region. Immutable.
    location = optional(string, "")

    # The GKE cluster to register: a GcpGkeCluster reference (its
    # cluster_id, "projects/{p}/locations/{l}/clusters/{name}") or that
    # literal path. Both modules send it as Google's resource link,
    # "//container.googleapis.com/projects/{p}/locations/{l}/clusters/{name}".
    # Omit it only for a cluster outside Google Cloud that registers through
    # its own Connect agent. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    gke_cluster = optional(string, "")

    # The cluster's OIDC issuer, which turns on fleet Workload Identity for
    # this membership: Google then trusts tokens from this issuer within the
    # fleet's workload identity pool ("{project}.hub.id.goog"). For a GKE
    # cluster it is "https://container.googleapis.com/v1/" followed by the
    # cluster's cluster_id output -- Google requires the locations/ form, so
    # the cluster's self_link (zones/ for zonal clusters) does not fit.
    # Empty leaves fleet Workload Identity off. Immutable: Google refuses an
    # issuer change in place.
    issuer = optional(string, "")

    # Labels on the membership. The platform attribution labels are added on
    # top and win on key conflicts.
    labels = optional(map(string), {})

    # What destroy does:
    #   "" / "DELETE" -- the cluster is unregistered from the fleet
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the membership leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
