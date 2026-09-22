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
  description = "GcpRedisClusterEndpointSet specification"
  type = object({
    # The GCP project the cluster lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The cluster the connections belong to. A GcpRedisCluster reference
    # resolves to its full resource path (name output); a literal takes
    # either the full path or the bare cluster name. The modules derive the
    # bare name Google's resource expects.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster = string

    # The cluster's region (e.g. "us-central1").
    region = string

    # One entry per consumer VPC, each carrying a connection per cluster
    # service attachment. This list IS the cluster's user-created endpoint
    # set: anything not listed is removed on apply.
    endpoints = list(object({
      # The connections that make up this endpoint, one per cluster service
      # attachment, all in one consumer VPC.
      connections = list(object({
        # The consumer-side forwarding rule, as its URI. A GcpGlobalForwardingRule
        # reference (the regional PSC form: empty scheme, the cluster's service
        # attachment as target) resolves to its self_link output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        forwarding_rule = string

        # The PSC connection ID Google assigned to that forwarding rule. The SAME
        # GcpGlobalForwardingRule as forwarding_rule, referenced on its
        # psc_connection_id output -- a reference names one output path, so the
        # rule is named twice.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        psc_connection_id = string

        # The IP address the forwarding rule serves on the consumer network. A
        # GcpAddress reference (the reserved internal address the rule was given)
        # resolves to its address output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        address = string

        # The consumer VPC network the address lives in, as
        # projects/{project}/global/networks/{name}. A GcpVpcNetwork reference
        # resolves to its network_id output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = string

        # The cluster service attachment this connection targets, as
        # projects/{project}/regions/{region}/serviceAttachments/{id}. A
        # GcpRedisCluster reference resolves to its discovery_service_attachment
        # output by default; the connection for the primary or reader endpoint
        # names that handle through an explicit fieldPath
        # (status.outputs.primary_service_attachment,
        # status.outputs.reader_service_attachment).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_attachment = string

        # The consumer project the forwarding rule was created in. If omitted,
        # Google records the project it finds on the rule -- set it only when
        # the rule lives in a different project than the cluster.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_id = optional(string, "")
      }))
    }))

    # What happens to the registration when this resource is destroyed:
    #   "" / "DELETE" -- the connections are deregistered (the forwarding
    #                    rules themselves belong to their own blocks)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the registration leaves management but stays on
    #                    the cluster
    deletion_policy = optional(string, "")
  })
}
