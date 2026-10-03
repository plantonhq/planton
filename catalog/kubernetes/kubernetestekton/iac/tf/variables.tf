# Typed mirror of KubernetesTektonSpec (spec.proto). The spec arrives
# from the proto->tfvars converter in snake_case. The kind has no
# namespace field of its own — the TektonConfig's targetNamespace is
# where the operator installs the components.

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
  description = "KubernetesTekton specification"
  type = object({
    profile          = optional(string, "all")
    target_namespace = optional(string, "tekton-pipelines")
    target_namespace_metadata = optional(object({
      labels      = optional(map(string), {})
      annotations = optional(map(string), {})
    }))
    placement = optional(object({
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        key                = optional(string, "")
        operator           = optional(string, "")
        value              = optional(string, "")
        effect             = optional(string, "")
        toleration_seconds = optional(number)
      })), [])
      priority_class_name = optional(string, "")
    }))
    pipeline = optional(object({
      cloud_events_sink_url   = optional(string, "")
      enable_api_fields       = optional(string, "")
      default_timeout_minutes = optional(number)
      default_service_account = optional(string, "")
      features = optional(object({
        disable_creds_init                            = optional(bool)
        await_sidecar_readiness                       = optional(bool)
        running_in_environment_with_injected_sidecars = optional(bool)
        require_git_ssh_secret_known_hosts            = optional(bool)
        enable_custom_tasks                           = optional(bool)
        keep_pod_on_cancel                            = optional(bool)
        enable_provenance_in_status                   = optional(bool)
        set_security_context                          = optional(bool)
        enable_cel_in_whenexpression                  = optional(bool)
        enable_step_actions                           = optional(bool)
        enable_param_enum                             = optional(bool)
        results_from                                  = optional(string, "")
        max_result_size                               = optional(number)
        coschedule                                    = optional(string, "")
      }))
      resolvers = optional(object({
        enable_bundles_resolver = optional(bool)
        enable_hub_resolver     = optional(bool)
        enable_git_resolver     = optional(bool)
        enable_cluster_resolver = optional(bool)
      }))
      metrics = optional(object({
        taskrun_level             = optional(string, "")
        taskrun_duration_type     = optional(string, "")
        pipelinerun_level         = optional(string, "")
        pipelinerun_duration_type = optional(string, "")
        count_with_reason         = optional(bool)
      }))
      performance = optional(object({
        replicas               = optional(number)
        buckets                = optional(number)
        threads_per_controller = optional(number)
        kube_api_qps           = optional(number)
        kube_api_burst         = optional(number)
      }))
    }))
    trigger = optional(object({
      enable_api_fields       = optional(string, "")
      default_service_account = optional(string, "")
    }))
    dashboard = optional(object({
      readonly      = optional(bool, false)
      external_logs = optional(string, "")
    }))
    chain = optional(object({
      disabled                = optional(bool, false)
      generate_signing_secret = optional(bool, false)
    }))
    pruner = optional(object({
      schedule           = string
      resources          = list(string)
      keep               = optional(number)
      keep_since         = optional(number)
      prune_per_resource = optional(bool, false)
    }))
    additional_params = optional(list(object({
      name  = string
      value = string
    })), [])
  })
}
