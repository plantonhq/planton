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
  description = "spec"
  type = object({

    # Kubernetes namespace for the Helm release
    namespace = optional(object({
      value = optional(string)
      ref = optional(object({
        kind = optional(string)
        name = optional(string)
      }))
    }))

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The repository URL where the Helm chart is hosted.
    # For example, "https://charts.helm.sh/stable".
    # an example for chart-repo (redis chart) can be found in https://artifacthub.io/packages/helm/bitnami/redis?modal=install
    repo = optional(string)

    # The name of the Helm chart to deploy.
    # For example, "nginx-ingress".
    name = string

    # The version of the Helm chart to deploy.
    # For example, "1.41.3".
    version = optional(string)

    # A map of key-value pairs representing custom values for the Helm chart.
    # These values override the default settings in the chart's values.yaml file.
    # values = object({
    #
    #   # Description for key
    #   key = string
    #
    #   # Description for value
    #   value = string
    # })
  })
}
