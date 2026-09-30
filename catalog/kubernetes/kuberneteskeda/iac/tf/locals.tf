# Computed values for the KubernetesKeda module. Every resolution here has
# an exact twin in the Pulumi module's locals.go / values.go — keep them in
# lockstep.
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
# short-circuit, so chained null checks still dereference the null — and
# var.spec is typed 'any', so an absent attribute is an error, not a null.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's vars:
  # cross-engine chart-name drift deploys two different products from one
  # manifest.
  helm_chart_name = "keda"
  helm_chart_repo = "https://kedacore.github.io/charts"

  # Release name FIXED to the chart name: KEDA registers the cluster-wide
  # v1beta1.external.metrics.k8s.io APIService, a singleton — one
  # installation per cluster is an upstream constraint.
  release_name = local.helm_chart_name

  # Chart version resolved to the pinned default when unset, so both engines
  # install the same chart whether or not the platform's defaulting
  # middleware ran — mirror of the Pulumi module's DefaultChartVersion
  # (chart 2.20.1 = KEDA 2.20.1 — chart and app versions move together).
  chart_version = coalesce(try(var.spec.chart_version, null), "2.20.1")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the namespace this module creates
  # (never injected into the chart's own resources — Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesKeda"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- CRD lifecycle -------------------------------------------------------
  # crds.install matches the chart's own default (true) — rendered only on
  # explicit opt-out. keep_on_uninstall has NO chart knob: the chart
  # templates its CRDs and Helm would DELETE them on uninstall, cascading to
  # every ScaledObject/ScaledJob/TriggerAuthentication in the cluster.
  # Planton default keeps them via the standard Helm resource-policy
  # annotation, which the chart forwards onto the CRDs
  # (crds.additionalAnnotations) — the ESO-family precedent. The keep
  # annotation only makes sense when this release owns the CRDs, so it rides
  # along only when install && keep.
  crds_install = try(var.spec.crds.install, null) != null ? var.spec.crds.install : true
  crds_keep    = try(var.spec.crds.keep_on_uninstall, null) != null ? var.spec.crds.keep_on_uninstall : true

  crds_values = {
    for k, v in {
      install               = local.crds_install ? null : false
      additionalAnnotations = local.crds_install && local.crds_keep ? { "helm.sh/resource-policy" = "keep" } : null
    } : k => v if v != null
  }

  # ---- shared ContainerResources shape (per component) ----------------------
  # The chart groups resources under ONE top-level block keyed per component
  # — and the metrics server's key there is "metricServer" (SINGULAR),
  # unlike the "metricsServer" component block. Trap; twin of the Pulumi
  # module's resourcesMap call sites.
  operator_resources = try(var.spec.operator.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.operator.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.operator.resources.limits.cpu, "") != "" ? var.spec.operator.resources.limits.cpu : null
          memory = try(var.spec.operator.resources.limits.memory, "") != "" ? var.spec.operator.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.operator.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.operator.resources.requests.cpu, "") != "" ? var.spec.operator.resources.requests.cpu : null
          memory = try(var.spec.operator.resources.requests.memory, "") != "" ? var.spec.operator.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  metrics_server_resources = try(var.spec.metrics_server.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.metrics_server.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.metrics_server.resources.limits.cpu, "") != "" ? var.spec.metrics_server.resources.limits.cpu : null
          memory = try(var.spec.metrics_server.resources.limits.memory, "") != "" ? var.spec.metrics_server.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.metrics_server.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.metrics_server.resources.requests.cpu, "") != "" ? var.spec.metrics_server.resources.requests.cpu : null
          memory = try(var.spec.metrics_server.resources.requests.memory, "") != "" ? var.spec.metrics_server.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  webhooks_resources = try(var.spec.webhooks.resources, null) == null ? null : {
    for k, v in {
      limits = try(var.spec.webhooks.resources.limits, null) == null ? null : {
        for lk, lv in {
          cpu    = try(var.spec.webhooks.resources.limits.cpu, "") != "" ? var.spec.webhooks.resources.limits.cpu : null
          memory = try(var.spec.webhooks.resources.limits.memory, "") != "" ? var.spec.webhooks.resources.limits.memory : null
        } : lk => lv if lv != null
      }
      requests = try(var.spec.webhooks.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = try(var.spec.webhooks.resources.requests.cpu, "") != "" ? var.spec.webhooks.resources.requests.cpu : null
          memory = try(var.spec.webhooks.resources.requests.memory, "") != "" ? var.spec.webhooks.resources.requests.memory : null
        } : rk => rv if rv != null
      }
    } : k => v if v != null && length(v) > 0
  }

  resources_values = {
    for k, v in {
      operator     = local.operator_resources
      metricServer = local.metrics_server_resources
      webhooks     = local.webhooks_resources
    } : k => v if v != null && length(v) > 0
  }

  # ---- admission webhooks ----------------------------------------------------
  # enabled matches the chart's own default (true) — rendered only on
  # explicit opt-out. Resources live in the shared block above, replicas
  # here — the chart's asymmetric layout.
  webhooks_values = {
    for k, v in {
      enabled       = try(var.spec.webhooks.enabled, null) == false ? false : null
      failurePolicy = try(var.spec.webhooks.failure_policy, null) != null && try(var.spec.webhooks.failure_policy, "") != "" ? var.spec.webhooks.failure_policy : null
      replicaCount  = try(var.spec.webhooks.replicas, null)
    } : k => v if v != null
  }

  # ---- pod identity ------------------------------------------------------------
  # The arms configure independent chart blocks — enabled cross-cloud
  # combinations render side by side.
  pod_identity_values = {
    for k, v in {
      aws = try(var.spec.pod_identity.aws_irsa.enabled, false) ? {
        irsa = {
          enabled = true
          roleArn = var.spec.pod_identity.aws_irsa.role_arn
        }
      } : null
      azureWorkload = try(var.spec.pod_identity.azure_workload_identity.enabled, false) ? {
        enabled  = true
        clientId = var.spec.pod_identity.azure_workload_identity.client_id
        tenantId = var.spec.pod_identity.azure_workload_identity.tenant_id
      } : null
      gcp = try(var.spec.pod_identity.gcp_workload_identity.enabled, false) ? {
        enabled              = true
        gcpIAMServiceAccount = var.spec.pod_identity.gcp_workload_identity.service_account_email
      } : null
    } : k => v if v != null
  }

  # ---- internal TLS certificates --------------------------------------------------
  # "operator" (or unset) is the chart default — the KEDA operator
  # self-generates certificates and patches the APIService caBundle; nothing
  # to render. With cert_manager and no issuer reference the chart generates
  # its own self-signed CA + Issuer chain — the issuer block stays absent.
  certificates_type = try(var.spec.certificates.type, "operator")
  cert_manager_values = {
    for k, v in {
      enabled = true
      # The arm is the grain: cluster_issuer renders a ClusterIssuer, issuer
      # a namespaced Issuer.
      issuer = try(var.spec.certificates.cert_manager_issuer, null) != null ? {
        generate = false
        name     = try(var.spec.certificates.cert_manager_issuer.cluster_issuer.name, var.spec.certificates.cert_manager_issuer.issuer.name)
        kind     = try(var.spec.certificates.cert_manager_issuer.cluster_issuer.name, null) != null ? "ClusterIssuer" : "Issuer"
        group    = "cert-manager.io"
      } : null
    } : k => v if v != null
  }
  certificates_values = local.certificates_type == "cert_manager" ? {
    certManager = local.cert_manager_values
  } : null

  # ---- own telemetry ---------------------------------------------------------------
  # KEDA exposes its own /metrics per component — the chart mirrors the
  # per-component layout, so one spec flag fans out to the operator and
  # metrics-server blocks identically.
  prometheus_component_values = {
    for k, v in {
      enabled        = true
      serviceMonitor = try(var.spec.prometheus.service_monitor, false) ? { enabled = true } : null
    } : k => v if v != null
  }
  prometheus_values = try(var.spec.prometheus.enabled, false) ? {
    operator     = local.prometheus_component_values
    metricServer = local.prometheus_component_values
  } : null

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) ------
  # No fullnameOverride: the chart names its components keda-operator /
  # keda-operator-metrics-apiserver / keda-admission-webhooks independent of
  # the release name — there is nothing for an override to pin.
  typed_values = {
    for k, v in {
      crds = length(local.crds_values) > 0 ? local.crds_values : null

      watchNamespace = try(var.spec.watch_namespace, "") != "" ? var.spec.watch_namespace : null

      # Replica counts live under each component while resources are
      # grouped in the shared top-level block — the chart's asymmetric key
      # layout (operator.replicaCount vs resources.operator).
      operator = try(var.spec.operator.replicas, null) != null ? {
        replicaCount = var.spec.operator.replicas
      } : null
      metricsServer = try(var.spec.metrics_server.replicas, null) != null ? {
        replicaCount = var.spec.metrics_server.replicas
      } : null
      webhooks  = length(local.webhooks_values) > 0 ? local.webhooks_values : null
      resources = length(local.resources_values) > 0 ? local.resources_values : null

      podIdentity  = length(local.pod_identity_values) > 0 ? local.pod_identity_values : null
      certificates = local.certificates_values

      http = try(var.spec.http_timeout_ms, null) != null ? {
        timeout = var.spec.http_timeout_ms
      } : null

      priorityClassName = try(var.spec.priority_class_name, "") != "" ? var.spec.priority_class_name : null
      nodeSelector      = length(try(var.spec.node_selector, {})) > 0 ? var.spec.node_selector : null
      tolerations = length(try(var.spec.tolerations, [])) > 0 ? [
        for t in var.spec.tolerations : {
          for tk, tv in {
            key               = try(t.key, "") != "" ? t.key : null
            operator          = try(t.operator, "") != "" ? t.operator : null
            value             = try(t.value, "") != "" ? t.value : null
            effect            = try(t.effect, "") != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null

      prometheus = local.prometheus_values
    } : k => v if v != null
  }
}
