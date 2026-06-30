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
  description = <<-EOT
    Specification for KubernetesTekton manifest-based deployment.
    
    IMPORTANT: Namespace Behavior
    Tekton components are installed in the 'tekton-pipelines' namespace,
    which is created automatically by the official Tekton manifests.
    This namespace cannot be customized.
  EOT
  type = object({

    # The Kubernetes cluster to install Tekton on.
    target_cluster = optional(object({
      cluster_name = string
      cluster_kind = optional(number)
    }))

    # The version of Tekton Pipelines to deploy.
    # Maps to release versions from https://github.com/tektoncd/pipeline/releases
    # Examples: "latest", "v0.65.2", "v0.64.0"
    pipeline_version = optional(string, "latest")

    # Dashboard configuration for Tekton.
    dashboard = optional(object({
      # Flag to enable or disable dashboard deployment.
      enabled = optional(bool, false)

      # The version of Tekton Dashboard to deploy.
      # Maps to release versions from https://github.com/tektoncd/dashboard/releases
      # Examples: "latest", "v0.53.0", "v0.52.0"
      version = optional(string, "latest")

      # Ingress configuration for external access to the dashboard.
      ingress = optional(object({
        # Flag to enable or disable dashboard ingress.
        enabled = optional(bool, false)

        # The full hostname for external access to the dashboard.
        hostname = optional(string)
      }))
    }))

    # Cloud Events configuration for pipeline notifications.
    cloud_events = optional(object({
      # The URL where CloudEvents will be sent.
      sink_url = string
    }))
  })
}
