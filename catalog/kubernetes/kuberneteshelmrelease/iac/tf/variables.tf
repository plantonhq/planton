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
  description = "KubernetesHelmRelease specification"
  type = object({
    namespace                  = string
    create_namespace           = optional(bool, false)
    repo                       = string
    chart                      = string
    version                    = string
    release_name               = optional(string, "")
    values_yaml                = optional(string, "")
    set                        = optional(map(string), {})
    set_string                 = optional(map(string), {})
    set_sensitive              = optional(map(string), {})
    repository_username        = optional(string, "")
    repository_password        = optional(string, "")
    atomic                     = optional(bool, false)
    cleanup_on_fail            = optional(bool, false)
    skip_await                 = optional(bool, false)
    wait_for_jobs              = optional(bool, false)
    timeout_seconds            = optional(number)
    dependency_update          = optional(bool, false)
    max_history                = optional(number)
    replace                    = optional(bool, false)
    force_update               = optional(bool, false)
    reuse_values               = optional(bool, false)
    reset_values               = optional(bool, false)
    disable_webhooks           = optional(bool, false)
    disable_openapi_validation = optional(bool, false)
    take_ownership             = optional(bool, false)
    description                = optional(string, "")
    crds = optional(object({
      install            = optional(bool)
      keep_on_uninstall  = optional(bool)
      allow_helm_managed = optional(bool)
    }))
  })
}