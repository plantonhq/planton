# Computed values for the KubernetesMetricsServer module. Every resolution
# here has an exact twin in the Pulumi module's locals.go / values.go — keep
# them in lockstep.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# The two tempting alternatives are both broken: `cond ? {...} : {}`
# ternaries fail plan-time type unification when branches carry different
# attributes, and `merge(concat(cond ? [{...}] : [], ...)...)` silently
# UNIFIES primitive-only sibling objects into map(string) — numbers and
# booleans arrive in the chart values as strings. The null-prune form
# preserves every value's type.
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's vars:
  # cross-engine chart-name drift deploys two different products from one
  # manifest.
  helm_chart_name = "metrics-server"
  helm_chart_repo = "https://kubernetes-sigs.github.io/metrics-server/"

  # Release name FIXED to the chart name: metrics-server registers the
  # cluster-wide v1beta1.metrics.k8s.io APIService, a singleton — one
  # installation per cluster is an upstream constraint.
  release_name = local.helm_chart_name

  # Chart version resolved to the pinned default when unset, so both engines
  # install the same chart whether or not the platform's defaulting
  # middleware ran — mirror of the Pulumi module's DefaultChartVersion.
  chart_version = coalesce(var.spec.chart_version, "3.13.1")

  namespace = var.spec.namespace

  # The APIService name is fixed by the API contract, not the chart (the
  # resource-metrics API has exactly this one group/version). Empty only
  # when the spec explicitly opts out of creating it (default: create).
  api_service_name = try(var.spec.api_service.create, null) == false ? "" : "v1beta1.metrics.k8s.io"

  # Resource-identity labels stamped on the namespace this module creates
  # (never injected into the chart's own resources — Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesMetricsServer"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- kubelet scrape flags ---------------------------------------------
  # The chart concatenates defaultArgs + args into the container command
  # line, and pflag's last-occurrence-wins would make duplicates confusing
  # rather than broken — so the module OWNS defaultArgs: it re-renders the
  # chart's default list with the typed substitutions applied, keeping the
  # pod spec canonical. Twin of the Pulumi module's kubeletAddressTypes /
  # metricResolution helpers.
  kubelet_address_types = length(var.spec.kubelet_preferred_address_types) > 0 ? join(",", var.spec.kubelet_preferred_address_types) : "InternalIP,ExternalIP,Hostname"
  metric_resolution     = coalesce(var.spec.metric_resolution, "15s")

  default_args = [
    "--cert-dir=/tmp",
    "--kubelet-preferred-address-types=${local.kubelet_address_types}",
    "--kubelet-use-node-status-port",
    "--metric-resolution=${local.metric_resolution}",
  ]

  # ---- serving certificate (tls values) -----------------------------------
  # ONE object shape with null-pruned conditional keys — a per-type ternary
  # chain would trip HCL's "inconsistent conditional result types" because
  # each arm carries a different attribute set (the exact class this file's
  # header warns about; caught by the offline plan proof).
  tls_type = try(var.spec.tls.type, "self_signed")
  tls_values_raw = {
    type = (
      local.tls_type == "helm" ? "helm" :
      local.tls_type == "cert_manager" ? "cert-manager" :
      local.tls_type == "existing_secret" ? "existingSecret" :
      null
    )
    certManager = local.tls_type == "cert_manager" && try(var.spec.tls.cert_manager_issuer, null) != null ? {
      # The arm is the grain: cluster_issuer renders a ClusterIssuer, issuer
      # a namespaced Issuer.
      existingIssuer = {
        enabled = true
        kind    = try(var.spec.tls.cert_manager_issuer.cluster_issuer.name, null) != null ? "ClusterIssuer" : "Issuer"
        name    = try(var.spec.tls.cert_manager_issuer.cluster_issuer.name, var.spec.tls.cert_manager_issuer.issuer.name)
      }
    } : null
    existingSecret = local.tls_type == "existing_secret" ? {
      name = try(var.spec.tls.existing_secret_name, "")
    } : null
  }
  # self_signed is the chart default ("metrics-server"); nothing to render.
  tls_values = local.tls_type == "self_signed" ? null : {
    for k, v in local.tls_values_raw : k => v if v != null
  }

  # ---- shared ContainerResources shape --------------------------------------
  server_resources = try(var.spec.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.resources.limits.cpu, "") != "" ? var.spec.resources.limits.cpu : null
          memory = try(var.spec.resources.limits.memory, "") != "" ? var.spec.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.resources.requests.cpu, "") != "" ? var.spec.resources.requests.cpu : null
          memory = try(var.spec.resources.requests.memory, "") != "" ? var.spec.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) ------
  typed_values = {
    for k, v in {
      # Pin the chart's fullname to the (fixed) release name: chart objects
      # get deterministic names ("metrics-server", the Service the
      # APIService routes to) — what verification and imports key off.
      fullnameOverride = local.release_name

      replicas = try(var.spec.replicas, null)

      defaultArgs = local.default_args
      args        = var.spec.kubelet_insecure_tls ? ["--kubelet-insecure-tls"] : null

      hostNetwork = var.spec.host_network ? { enabled = true } : null

      apiService = try(var.spec.api_service, null) == null ? null : {
        for ak, av in {
          create                = try(var.spec.api_service.create, null) == false ? false : null
          insecureSkipTLSVerify = try(var.spec.api_service.insecure_skip_tls_verify, null) == false ? false : null
          caBundle              = try(var.spec.api_service.ca_bundle, "") != "" ? var.spec.api_service.ca_bundle : null
        } : ak => av if av != null
      }

      tls = local.tls_values

      resources    = local.server_resources
      nodeSelector = length(var.spec.node_selector) > 0 ? var.spec.node_selector : null
      tolerations = length(var.spec.tolerations) > 0 ? [
        for t in var.spec.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.priority_class_name, null) != null && try(var.spec.priority_class_name, "") != "" ? var.spec.priority_class_name : null

      podDisruptionBudget = var.spec.pod_disruption_budget ? {
        enabled      = true
        minAvailable = 1
      } : null

      metrics = try(var.spec.prometheus.enabled, false) ? { enabled = true } : null
      serviceMonitor = try(var.spec.prometheus.enabled, false) && try(var.spec.prometheus.service_monitor, false) ? {
        for sk, sv in {
          enabled          = true
          interval         = try(var.spec.prometheus.service_monitor_interval, null)
          additionalLabels = length(try(var.spec.prometheus.service_monitor_labels, {})) > 0 ? var.spec.prometheus.service_monitor_labels : null
        } : sk => sv if sv != null
      } : null

      image = (
        try(var.spec.image.repository, "") != "" || try(var.spec.image.tag, "") != ""
        ) ? {
        for ik, iv in {
          repository = try(var.spec.image.repository, "") != "" ? var.spec.image.repository : null
          tag        = try(var.spec.image.tag, "") != "" ? var.spec.image.tag : null
        } : ik => iv if iv != null
      } : null
    } : k => v if v != null
  }
}
