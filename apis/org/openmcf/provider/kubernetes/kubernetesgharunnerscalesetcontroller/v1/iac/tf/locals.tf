locals {
  release_name = var.metadata.name
  chart_oci    = "oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller"

  resource_id = coalesce(var.metadata.id, var.metadata.name)

  labels = merge(
    {
      "resource"      = "true"
      "resource_id"   = local.resource_id
      "resource_kind" = "gha_runner_scale_set_controller_kubernetes"
      "resource_name" = var.metadata.name
    },
    var.metadata.org != null && var.metadata.org != "" ? { "organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "environment" = var.metadata.env } : {}
  )

  metrics_enabled = var.spec.metrics != null && try(var.spec.metrics.controller_manager_addr, "") != ""

  helm_values = {
    replicaCount = var.spec.replica_count
    labels       = local.labels

    resources = {
      requests = {
        cpu    = try(var.spec.container.resources.requests.cpu, "100m")
        memory = try(var.spec.container.resources.requests.memory, "128Mi")
      }
      limits = {
        cpu    = try(var.spec.container.resources.limits.cpu, "500m")
        memory = try(var.spec.container.resources.limits.memory, "512Mi")
      }
    }

    flags = merge(
      {
        logLevel                      = try(var.spec.flags.log_level, "debug")
        logFormat                     = try(var.spec.flags.log_format, "text")
        runnerMaxConcurrentReconciles = try(var.spec.flags.runner_max_concurrent_reconciles, 2)
        updateStrategy                = try(var.spec.flags.update_strategy, "immediate")
      },
      try(var.spec.flags.watch_single_namespace, "") != "" ? { watchSingleNamespace = var.spec.flags.watch_single_namespace } : {},
      try(length(var.spec.flags.exclude_label_propagation_prefixes), 0) > 0 ? { excludeLabelPropagationPrefixes = var.spec.flags.exclude_label_propagation_prefixes } : {},
      try(var.spec.flags.k8s_client_rate_limiter_qps, 0) > 0 ? { k8sClientRateLimiterQPS = var.spec.flags.k8s_client_rate_limiter_qps } : {},
      try(var.spec.flags.k8s_client_rate_limiter_burst, 0) > 0 ? { k8sClientRateLimiterBurst = var.spec.flags.k8s_client_rate_limiter_burst } : {}
    )

    priorityClassName = try(var.spec.priority_class_name, "")

    imagePullSecrets = [for s in try(var.spec.image_pull_secrets, []) : { name = s }]
  }

  # Helm values split into separate yamlencode entries to avoid HCL
  # "Inconsistent conditional result types" errors. Helm deep-merges
  # values list entries in order (like multiple -f flags).
  helm_values_list = concat(
    [yamlencode(local.helm_values)],
    try(var.spec.container.image.repository, "") != "" ? [yamlencode({
      image = {
        repository = var.spec.container.image.repository
        tag        = var.spec.container.image.tag
        pullPolicy = var.spec.container.image.pull_policy
      }
    })] : [],
    local.metrics_enabled ? [yamlencode({
      metrics = {
        controllerManagerAddr = var.spec.metrics.controller_manager_addr
        listenerAddr          = try(var.spec.metrics.listener_addr, "")
        listenerEndpoint      = try(var.spec.metrics.listener_endpoint, "")
      }
    })] : []
  )
}
