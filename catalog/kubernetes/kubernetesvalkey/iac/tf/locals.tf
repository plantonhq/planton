# Computed values for the KubernetesValkey module. Every resolution here
# has an exact twin in the Pulumi module's locals.go / values.go /
# secrets.go — keep them in lockstep.
#
# HCL DISCIPLINE (applies to every conditional object in this file):
# conditional entries are written as `key = cond ? value : null` inside ONE
# object literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# The two tempting alternatives are both broken: `cond ? {...} : {}`
# ternaries fail plan-time type unification when branches carry different
# attributes, and `merge(concat(cond ? [{...}] : [], ...)...)` silently
# UNIFIES primitive-only sibling objects into map(string) — numbers and
# booleans arrive in the chart values as strings. The null-prune form
# preserves every value's type. Lists of same-shaped elements assemble with
# concat().
#
# Optional nested blocks are read with try(): HCL's && does NOT
# short-circuit, so chained null checks still dereference the null.

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's vars:
  # cross-engine chart-name drift deploys two different products from one
  # manifest.
  helm_chart_name = "valkey"
  helm_chart_repo = "https://valkey.io/valkey-helm/"

  # Release name — metadata.name, NOT a fixed chart name: several Valkey
  # instances coexist in one cluster, so each manifest gets its own
  # release. The chart's fullname is pinned to the same value below, which
  # is what makes the rendered Service names (`<name>`, `<name>-headless`,
  # `<name>-read`) deterministic.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset, so both
  # engines install the same chart whether or not the platform's defaulting
  # middleware ran — mirror of the Pulumi module's DefaultChartVersion.
  # Chart and app versions move separately (chart 0.11.0 ships Valkey
  # 9.1.1); the chart pin governs.
  chart_version = coalesce(var.spec.chart_version, "0.11.0")

  namespace = var.spec.namespace

  # Resource-identity labels stamped on the module-created satellites
  # (namespace, the auth Secret — never injected into the chart's own
  # resources; Helm owns those).
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesValkey"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- topology / feature gates -----------------------------------------
  auth_enabled        = try(var.spec.auth, null) != null
  replication_enabled = try(var.spec.replication, null) != null

  # The chart renders the read Service by default in replication mode;
  # only an explicit enabled=false suppresses it.
  read_service_enabled = local.replication_enabled ? coalesce(try(var.spec.replication.read_service.enabled, null), true) : false

  # Deterministic name of the module-materialized password Secret (one key
  # per ACL username). The chart consumes it via auth.usersExistingSecret,
  # so passwords never appear in rendered chart values.
  auth_secret_name = "${var.metadata.name}-auth"

  # The two halves of the auth Secret's data, split by whether the user
  # brought a password. An absent or empty password means "mint one":
  # random_password.user is keyed by exactly these usernames, and
  # declared_passwords carries the rest verbatim. Both are empty maps when
  # auth is off, so every consumer can reference them unconditionally.
  generated_password_users = local.auth_enabled ? {
    for user in var.spec.auth.users : user.name => user
    if try(coalesce(user.password), "") == ""
  } : {}
  declared_passwords = local.auth_enabled ? {
    for user in var.spec.auth.users : user.name => user.password
    if try(coalesce(user.password), "") != ""
  } : {}

  # The write Service port, resolved to the chart default when the service
  # block or its port is unset — feeds the endpoint outputs. port is an
  # OPTIONAL scalar: unset arrives as null, and try(coalesce(x), null)
  # is the null-safe read (coalesce rejects a lone null; try turns that
  # rejection into the fallback).
  service_port = try(coalesce(var.spec.service.port), null) != null ? var.spec.service.port : 6379
  service_type = try(coalesce(var.spec.service.type), null) != null ? var.spec.service.type : "ClusterIP"

  # ---- valkey.conf (module-owned rendering) -------------------------------
  # The chart accepts valkey.conf only as ONE raw string (valkeyConfig,
  # mounted via ConfigMap and appended after the chart's own generated base
  # config). The typed config block renders that string deterministically:
  # appendonly, save points (or the disable directive), maxmemory,
  # maxmemory-policy, then extra_directives verbatim. The line order and
  # joining MUST stay identical to the Pulumi module's renderValkeyConfig.
  # Note log_level is deliberately NOT here — it is a first-class chart
  # value (valkeyLogLevel), not a valkey.conf directive.
  valkey_config_lines = try(var.spec.config, null) == null ? [] : concat(
    var.spec.config.append_only ? ["appendonly yes"] : [],
    [for save_point in var.spec.config.rdb_save_points : "save ${save_point}"],
    var.spec.config.snapshots_disabled ? ["save \"\""] : [],
    var.spec.config.max_memory != "" ? ["maxmemory ${var.spec.config.max_memory}"] : [],
    var.spec.config.max_memory_policy != "" ? ["maxmemory-policy ${var.spec.config.max_memory_policy}"] : [],
    trimspace(var.spec.config.extra_directives) != "" ? [trimspace(var.spec.config.extra_directives)] : [],
  )
  valkey_config = length(local.valkey_config_lines) > 0 ? join("\n", local.valkey_config_lines) : null

  # ---- acl users -----------------------------------------------------------
  # Declared ACL users render WITHOUT passwords: the chart reads each
  # user's password from the usersExistingSecret key named after the
  # username (its init script falls back passwordKey -> username; the
  # module leaves passwordKey unset). The chart REQUIRES a permissions
  # value on every entry, so the spec default (full access) is resolved
  # here — mirror of the Pulumi module's defaultAclPermissions.
  acl_users = local.auth_enabled ? {
    for user in var.spec.auth.users : user.name => {
      permissions = coalesce(user.permissions, "~* &* +@all")
    }
  } : null

  # ---- container resources (shared ContainerResources shape) --------------
  valkey_resources = try(var.spec.resources, null) == null ? null : {
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

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = {
    for k, v in {
      # Pin the chart's fullname to the release name (= metadata.name):
      # every chart object then carries a deterministic, manifest-derived
      # name — the write Service renders as `<name>`, pod discovery as
      # `<name>-headless`, and the replication read Service as
      # `<name>-read`, which is exactly what the stack outputs promise.
      fullnameOverride = local.release_name

      # Chart defaults: docker.io / valkey/valkey at the chart's app
      # version.
      image = try(var.spec.image, null) == null ? null : {
        for ik, iv in {
          registry   = var.spec.image.registry != "" ? var.spec.image.registry : null
          repository = var.spec.image.repository != "" ? var.spec.image.repository : null
          tag        = var.spec.image.tag != "" ? var.spec.image.tag : null
        } : ik => iv if iv != null
      }

      # The chart consumes pull secrets as a plain list of Secret NAMES
      # (its imagePullSecrets helper renders `- name: <entry>` per string
      # entry).
      imagePullSecrets = length(var.spec.image_pull_secrets) > 0 ? var.spec.image_pull_secrets : null

      # Server log verbosity is a first-class chart value (injected as the
      # VALKEY_LOG_LEVEL env var) — NOT a valkey.conf directive.
      valkeyLogLevel = try(coalesce(var.spec.log_level), null)

      valkeyConfig = local.valkey_config

      auth = local.auth_enabled ? {
        enabled             = true
        usersExistingSecret = local.auth_secret_name
        aclUsers            = local.acl_users
      } : null

      # Replication: one primary plus N replicas from a StatefulSet. Every
      # scalar renders with its resolved default so both engines emit the
      # identical replica block.
      replica = local.replication_enabled ? {
        enabled            = true
        replicas           = coalesce(var.spec.replication.replicas, 2)
        replicationUser    = coalesce(var.spec.replication.replication_user, "default")
        disklessSync       = var.spec.replication.diskless_sync
        minReplicasToWrite = var.spec.replication.min_replicas_to_write
        minReplicasMaxLag  = coalesce(var.spec.replication.min_replicas_max_lag, 10)
        persistence = {
          for pk, pv in {
            size         = var.spec.replication.persistence.size
            storageClass = var.spec.replication.persistence.storage_class != "" ? var.spec.replication.persistence.storage_class : null
          } : pk => pv if pv != null
        }
        service = {
          for sk, sv in {
            enabled     = local.read_service_enabled
            type        = try(coalesce(var.spec.replication.read_service.type), null) != null ? var.spec.replication.read_service.type : "ClusterIP"
            annotations = length(try(var.spec.replication.read_service.annotations, {})) > 0 ? var.spec.replication.read_service.annotations : null
          } : sk => sv if sv != null
        }
      } : null

      # Standalone persistence: the chart's dataStorage PVC (only read by
      # the standalone Deployment; replication uses volumeClaimTemplates
      # instead — the spec CEL rule keeps the two declarations apart).
      dataStorage = (!local.replication_enabled && try(var.spec.persistence, null) != null) ? {
        for dk, dv in {
          enabled       = true
          requestedSize = var.spec.persistence.size
          className     = var.spec.persistence.storage_class != "" ? var.spec.persistence.storage_class : null
          keepPvc       = var.spec.persistence.keep_on_uninstall
        } : dk => dv if dv != null
      } : null

      # The chart's key-name defaults (server.crt/server.key/ca.crt)
      # predate the kubernetes.io/tls convention; cert-manager
      # Certificates store their material as tls.crt/tls.key/ca.crt. The
      # spec's certificate seam is cert-manager, so the module pins the
      # chart's key names to the kubernetes.io/tls layout whenever TLS is
      # enabled.
      tls = try(var.spec.tls.enabled, false) ? {
        enabled                  = true
        existingSecret           = var.spec.tls.certificate_secret
        requireClientCertificate = var.spec.tls.require_client_certificate
        serverPublicKey          = "tls.crt"
        serverKey                = "tls.key"
        caPublicKey              = "ca.crt"
      } : null

      service = try(var.spec.service, null) == null ? null : {
        for sk, sv in {
          type        = local.service_type
          port        = local.service_port
          annotations = length(var.spec.service.annotations) > 0 ? var.spec.service.annotations : null
        } : sk => sv if sv != null
      }

      resources = local.valkey_resources

      # The ServiceMonitor toggle lives at metrics.serviceMonitor.enabled;
      # the chart additionally gates it on metrics.service.enabled, which
      # defaults to true — the metrics Service (`<name>-metrics`) always
      # accompanies the exporter here.
      metrics = try(var.spec.metrics.enabled, false) ? {
        for mk, mv in {
          enabled        = true
          serviceMonitor = var.spec.metrics.service_monitor_enabled ? { enabled = true } : null
        } : mk => mv if mv != null
      } : null

      # ---- scheduling ----------------------------------------------------
      nodeSelector = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
      tolerations = length(try(var.spec.scheduling.tolerations, [])) > 0 ? [
        for t in var.spec.scheduling.tolerations : {
          for tk, tv in {
            key               = t.key != "" ? t.key : null
            operator          = t.operator != "" ? t.operator : null
            value             = t.value != "" ? t.value : null
            effect            = t.effect != "" ? t.effect : null
            tolerationSeconds = try(t.toleration_seconds, null)
          } : tk => tv if tv != null
        }
      ] : null
      priorityClassName = try(var.spec.scheduling.priority_class_name, "") != "" ? var.spec.scheduling.priority_class_name : null

      # Rendered ONLY in replication mode: the chart's PDB template is
      # gated on replica.enabled, so a standalone PDB declaration would be
      # a silent no-op in the release — the module omits it instead of
      # rendering dead values. Exactly one bound is set (spec CEL rule);
      # with neither, the chart's own default (maxUnavailable: 1) applies.
      podDisruptionBudget = (local.replication_enabled && try(var.spec.pod_disruption_budget.enabled, false)) ? {
        for pk, pv in {
          enabled        = true
          maxUnavailable = var.spec.pod_disruption_budget.max_unavailable > 0 ? var.spec.pod_disruption_budget.max_unavailable : null
          minAvailable   = var.spec.pod_disruption_budget.min_available > 0 ? var.spec.pod_disruption_budget.min_available : null
        } : pk => pv if pv != null
      } : null
    } : k => v if v != null && v != {}
  }
}
