variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for KubernetesPeerAuthentication"
  type = object({
    # Namespace the PeerAuthentication is created in (resolved foreign key).
    namespace = string

    # Workload selector. When omitted, the policy applies to all workloads in the
    # namespace (or, in the mesh root namespace, the whole mesh). Matched by
    # label at runtime by istiod; not an OpenMCF foreign key.
    selector = optional(object({
      match_labels = optional(map(string))
    }))

    # Mesh-TLS mode applied to the selected workloads. When omitted, the mode is
    # inherited from the parent (namespace, then mesh) policy.
    mtls = optional(object({
      mode = string
    }))

    # Per-port mTLS overrides, keyed by the workload's port number (as a string,
    # since CRD/JSON map keys are strings). Only honored when a selector is set.
    port_level_mtls = optional(map(object({
      mode = string
    })))
  })
}
