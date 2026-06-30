variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  # IMPORTANT: Namespace Behavior
  # Unlike other Kubernetes components in Planton, the Tekton Operator uses fixed namespaces
  # that are managed by the operator itself:
  # - The Tekton Operator is installed in the 'tekton-operator' namespace
  # - Tekton components (Pipelines, Triggers, Dashboard) are installed in the 'tekton-pipelines' namespace
  # These namespaces are automatically created and managed by the Tekton Operator and cannot be customized.
  description = "Specification for KubernetesTektonOperator"
  type = object({

    # The Kubernetes cluster to install this operator on.
    target_cluster = optional(object({
      cluster_name = string
      cluster_kind = optional(number)
    }))

    # The container specifications for the Tekton operator.
    container = object({

      # The CPU and memory resources allocated to the Tekton operator container.
      resources = object({

        # The resource limits for the container.
        limits = object({
          cpu    = string
          memory = string
        })

        # The resource requests for the container.
        requests = object({
          cpu    = string
          memory = string
        })
      })
    })

    # Configuration for which Tekton components to install.
    components = object({
      # Enable Tekton Pipelines component.
      pipelines = bool

      # Enable Tekton Triggers component.
      triggers = bool

      # Enable Tekton Dashboard component.
      dashboard = bool
    })

    # The version of the Tekton Operator to deploy.
    # Default value (v0.78.0) is set in spec.proto via options.default
    # https://github.com/tektoncd/operator/releases
    # https://operatorhub.io/operator/tektoncd-operator
    operator_version = optional(string)
  })
}
