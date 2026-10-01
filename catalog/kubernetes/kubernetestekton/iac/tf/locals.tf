# Computed values for the KubernetesTekton module. Every resolution here
# has an exact twin in the Pulumi module (tekton_config.go) — keep them
# in byte lockstep: same TektonConfig body, same outputs.
#
# HCL DISCIPLINE: conditional keys are contributed via merge() of
# `cond ? { key = value } : {}` singleton maps, or one object literal
# pruned with `{ for k, v in {...} : k => v if v != null }` — a ternary
# whose branches are differently-shaped objects fails plan-time type
# unification. Optional nested blocks are read with try(): HCL's && does
# NOT short-circuit.

locals {
  # THE CR NAME IS FIXED: the operator's admission webhook allows exactly
  # one TektonConfig per cluster and requires the name `config` —
  # metadata.name of the Planton resource keys the state identity only.
  tekton_config_name = "config"
  api_version        = "operator.tekton.dev/v1alpha1"

  # The resolved profile / target namespace (spec defaults mirror the
  # operator's own).
  profile          = try(var.spec.profile, "") != "" ? var.spec.profile : "all"
  target_namespace = try(var.spec.target_namespace, "") != "" ? var.spec.target_namespace : "tekton-pipelines"

  # Planton governance labels on the rendered CR.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesTekton"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- targetNamespaceMetadata ---------------------------------------------------
  target_namespace_metadata = {
    for k, v in {
      labels      = length(try(var.spec.target_namespace_metadata.labels, {})) > 0 ? var.spec.target_namespace_metadata.labels : null
      annotations = length(try(var.spec.target_namespace_metadata.annotations, {})) > 0 ? var.spec.target_namespace_metadata.annotations : null
    } : k => v if v != null
  }

  # ---- spec.config (placement for every Tekton component pod) --------------------
  placement_tolerations = [
    for t in try(var.spec.placement.tolerations, []) : { for k, v in {
      key               = t.key != "" ? t.key : null
      operator          = t.operator != "" ? t.operator : null
      value             = t.value != "" ? t.value : null
      effect            = t.effect != "" ? t.effect : null
      tolerationSeconds = t.toleration_seconds
    } : k => v if v != null }
  ]

  placement_config = {
    for k, v in {
      nodeSelector      = length(try(var.spec.placement.node_selector, {})) > 0 ? var.spec.placement.node_selector : null
      tolerations       = length(local.placement_tolerations) > 0 ? local.placement_tolerations : null
      priorityClassName = try(var.spec.placement.priority_class_name, "") != "" ? var.spec.placement.priority_class_name : null
    } : k => v if v != null
  }

  # ---- spec.pipeline --------------------------------------------------------------
  # Field names are the operator API's own JSON keys; values render ONLY
  # when declared so the operator's defaulting stays authoritative
  # (tri-state booleans: the proto optionals arrive as null when unset).
  # Every block is ONE object literal of try(..., null) values pruned
  # with the for-comprehension — NO outer ternary: `cond ? {} : {for…}`
  # fails HCL type unification (the empty object cannot gain the typed
  # branch's attributes; this exact expression failed the plan gate).
  pipeline_features = {
    for k, v in {
      "disable-creds-init"                            = try(var.spec.pipeline.features.disable_creds_init, null)
      "await-sidecar-readiness"                       = try(var.spec.pipeline.features.await_sidecar_readiness, null)
      "running-in-environment-with-injected-sidecars" = try(var.spec.pipeline.features.running_in_environment_with_injected_sidecars, null)
      "require-git-ssh-secret-known-hosts"            = try(var.spec.pipeline.features.require_git_ssh_secret_known_hosts, null)
      "enable-custom-tasks"                           = try(var.spec.pipeline.features.enable_custom_tasks, null)
      "keep-pod-on-cancel"                            = try(var.spec.pipeline.features.keep_pod_on_cancel, null)
      "enable-provenance-in-status"                   = try(var.spec.pipeline.features.enable_provenance_in_status, null)
      "set-security-context"                          = try(var.spec.pipeline.features.set_security_context, null)
      "enable-cel-in-whenexpression"                  = try(var.spec.pipeline.features.enable_cel_in_whenexpression, null)
      "enable-step-actions"                           = try(var.spec.pipeline.features.enable_step_actions, null)
      "enable-param-enum"                             = try(var.spec.pipeline.features.enable_param_enum, null)
      "results-from"                                  = try(var.spec.pipeline.features.results_from, "") != "" ? var.spec.pipeline.features.results_from : null
      "max-result-size"                               = try(var.spec.pipeline.features.max_result_size, null)
      "coschedule"                                    = try(var.spec.pipeline.features.coschedule, "") != "" ? var.spec.pipeline.features.coschedule : null
    } : k => v if v != null
  }

  pipeline_resolvers = {
    for k, v in {
      "enable-bundles-resolver" = try(var.spec.pipeline.resolvers.enable_bundles_resolver, null)
      "enable-hub-resolver"     = try(var.spec.pipeline.resolvers.enable_hub_resolver, null)
      "enable-git-resolver"     = try(var.spec.pipeline.resolvers.enable_git_resolver, null)
      "enable-cluster-resolver" = try(var.spec.pipeline.resolvers.enable_cluster_resolver, null)
    } : k => v if v != null
  }

  pipeline_metrics = {
    for k, v in {
      "metrics.taskrun.level"             = try(var.spec.pipeline.metrics.taskrun_level, "") != "" ? var.spec.pipeline.metrics.taskrun_level : null
      "metrics.taskrun.duration-type"     = try(var.spec.pipeline.metrics.taskrun_duration_type, "") != "" ? var.spec.pipeline.metrics.taskrun_duration_type : null
      "metrics.pipelinerun.level"         = try(var.spec.pipeline.metrics.pipelinerun_level, "") != "" ? var.spec.pipeline.metrics.pipelinerun_level : null
      "metrics.pipelinerun.duration-type" = try(var.spec.pipeline.metrics.pipelinerun_duration_type, "") != "" ? var.spec.pipeline.metrics.pipelinerun_duration_type : null
      "metrics.count.enable-reason"       = try(var.spec.pipeline.metrics.count_with_reason, null)
    } : k => v if v != null
  }

  pipeline_performance = {
    for k, v in {
      replicas                 = try(var.spec.pipeline.performance.replicas, null)
      buckets                  = try(var.spec.pipeline.performance.buckets, null)
      "threads-per-controller" = try(var.spec.pipeline.performance.threads_per_controller, null)
      "kube-api-qps"           = try(var.spec.pipeline.performance.kube_api_qps, null)
      "kube-api-burst"         = try(var.spec.pipeline.performance.kube_api_burst, null)
    } : k => v if v != null
  }

  pipeline_block_full = merge(
    {
      for k, v in {
        "enable-api-fields"       = try(var.spec.pipeline.enable_api_fields, "") != "" ? var.spec.pipeline.enable_api_fields : null
        "default-timeout-minutes" = try(var.spec.pipeline.default_timeout_minutes, null)
        "default-service-account" = try(var.spec.pipeline.default_service_account, "") != "" ? var.spec.pipeline.default_service_account : null
      } : k => v if v != null
    },
    local.pipeline_features,
    local.pipeline_resolvers,
    local.pipeline_metrics,
    length(local.pipeline_performance) > 0 ? { performance = local.pipeline_performance } : {},
    # The sink lives in the config-events ConfigMap, whose `sink` supersedes
    # config-defaults' deprecated default-cloud-events-sink. The TektonConfig
    # reaches that ConfigMap only through options.configMaps, which the
    # operator merges into the one it installs; tektonv1 is the one event
    # format Tekton emits. Pulumi twin: pipelineBody.
    try(var.spec.pipeline.cloud_events_sink_url, "") != "" ? {
      options = {
        configMaps = {
          "config-events" = {
            data = {
              sink    = var.spec.pipeline.cloud_events_sink_url
              formats = "tektonv1"
            }
          }
        }
      }
    } : {}
  )

  # ---- spec.trigger / spec.dashboard / spec.chain ---------------------------------
  trigger_block = {
    for k, v in {
      "enable-api-fields"       = try(var.spec.trigger.enable_api_fields, "") != "" ? var.spec.trigger.enable_api_fields : null
      "default-service-account" = try(var.spec.trigger.default_service_account, "") != "" ? var.spec.trigger.default_service_account : null
    } : k => v if v != null
  }

  # `readonly` is a required (non-pointer) upstream field — render it
  # whenever the block is declared (false must render, so its presence
  # gate is the block's own, not value pruning).
  dashboard_block = {
    for k, v in {
      readonly        = try(var.spec.dashboard, null) != null ? try(var.spec.dashboard.readonly, false) : null
      "external-logs" = try(var.spec.dashboard.external_logs, "") != "" ? var.spec.dashboard.external_logs : null
    } : k => v if v != null
  }

  chain_block = {
    for k, v in {
      disabled              = try(var.spec.chain, null) != null ? try(var.spec.chain.disabled, false) : null
      generateSigningSecret = try(var.spec.chain.generate_signing_secret, false) ? true : null
    } : k => v if v != null
  }

  # ---- spec.pruner ------------------------------------------------------------------
  # An absent block renders nothing — no pruner cron is scheduled
  # (Tekton's own default). `disabled: false` renders whenever the block
  # is declared (its presence gate, like the dashboard's readonly).
  pruner_block = {
    for k, v in {
      disabled             = try(var.spec.pruner, null) != null ? false : null
      schedule             = try(var.spec.pruner.schedule, null)
      resources            = try(var.spec.pruner.resources, null)
      keep                 = try(var.spec.pruner.keep, null)
      "keep-since"         = try(var.spec.pruner.keep_since, null)
      "prune-per-resource" = try(var.spec.pruner.prune_per_resource, false) ? true : null
    } : k => v if v != null
  }

  # ---- the TektonConfig spec body ---------------------------------------------------
  # Pulumi twin: tektonConfigSpecBody.
  tekton_config_spec = merge(
    {
      profile         = local.profile
      targetNamespace = local.target_namespace
    },
    length(local.target_namespace_metadata) > 0 ? { targetNamespaceMetadata = local.target_namespace_metadata } : {},
    length(local.placement_config) > 0 ? { config = local.placement_config } : {},
    length(local.pipeline_block_full) > 0 ? { pipeline = local.pipeline_block_full } : {},
    length(local.trigger_block) > 0 ? { trigger = local.trigger_block } : {},
    length(local.dashboard_block) > 0 ? { dashboard = local.dashboard_block } : {},
    length(local.chain_block) > 0 ? { chain = local.chain_block } : {},
    length(local.pruner_block) > 0 ? { pruner = local.pruner_block } : {},
    length(try(var.spec.additional_params, [])) > 0 ? {
      params = [for p in var.spec.additional_params : { name = p.name, value = p.value }]
    } : {}
  )

  # ---- outputs ----------------------------------------------------------------------
  # Dashboard handles are empty unless the profile installs the
  # dashboard (`all`).
  dashboard_service_name  = "tekton-dashboard"
  dashboard_port          = 9097
  dashboard_service       = local.profile == "all" ? local.dashboard_service_name : ""
  dashboard_kube_endpoint = local.profile == "all" ? "http://${local.dashboard_service_name}.${local.target_namespace}.svc.cluster.local:${local.dashboard_port}" : ""
  port_forward_command    = local.profile == "all" ? "kubectl port-forward -n ${local.target_namespace} service/${local.dashboard_service_name} ${local.dashboard_port}:${local.dashboard_port}" : ""
}
