# Computed values for the KubernetesOpenFga module.
# Every resolution here has an exact twin in the Pulumi module — keep
# them in lockstep: same rendered chart values, same outputs.
#
# HCL DISCIPLINE: conditional keys are contributed with the null-prune
# idiom — `key = cond ? value : null` inside a single for-comprehension
# that drops nulls. Never `cond ? {} : {...}` ternaries (differently
# shaped object branches fail plan-time type unification); optional
# scalars inside present blocks are read via try()/coalesce() where they
# feed string interpolation (HCL's && does NOT short-circuit).

locals {
  # Pinned chart identity; chart_version resolves to the pinned default
  # when unset so both engines install the same chart whether or not the
  # platform's defaulting middleware ran. Chart and app versions move in
  # lockstep (chart 0.3.10 = OpenFGA v1.18.1).
  helm_chart_name       = "openfga"
  helm_chart_repo       = "https://openfga.github.io/helm-charts"
  default_chart_version = "0.3.10"
  chart_version         = try(var.spec.chart_version, "") != "" ? var.spec.chart_version : local.default_chart_version

  namespace    = var.spec.namespace
  release_name = var.metadata.name

  # The chart's ClusterIP Service is openfga.fullname — pinned to the
  # resource name via fullnameOverride, so the endpoints below are
  # deterministic. HTTP 8080 / plaintext gRPC 8081 are the chart's
  # fixed ports (this module never moves them).
  service_name = var.metadata.name

  # Planton governance labels: stamped on the module-created satellites
  # (the namespace and the authn-keys Secret) and on every OpenFGA pod
  # through the chart's own podExtraLabels (the migration runs as an init
  # container in the same pod) -- so a log line, a metric or an alert from
  # it names its organization and environment. The chart's selector is its
  # own fixed labels; these never reach it.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesOpenFga"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- datastore arms ---------------------------------------------------------
  # URIs are engine-native DSNs WITHOUT userinfo: username and password
  # reach the server as OPENFGA_DATASTORE_USERNAME / _PASSWORD env vars,
  # which the server prefers over any URI-embedded credentials — nothing
  # credential-bearing lands in rendered values.
  is_postgres = try(var.spec.datastore.postgres, null) != null
  is_mysql    = try(var.spec.datastore.mysql, null) != null
  is_memory   = !local.is_postgres && !local.is_mysql

  datastore_engine = local.is_postgres ? "postgres" : (local.is_mysql ? "mysql" : "memory")

  postgres_port     = try(coalesce(var.spec.datastore.postgres.port, 5432), 5432)
  postgres_ssl_mode = try(coalesce(var.spec.datastore.postgres.ssl_mode, "disable"), "disable")
  mysql_port        = try(coalesce(var.spec.datastore.mysql.port, 3306), 3306)

  # parseTime=true is required by the server's mysql storage adapter.
  datastore_uri = local.is_postgres ? "postgres://${var.spec.datastore.postgres.host}:${local.postgres_port}/${var.spec.datastore.postgres.database}?sslmode=${local.postgres_ssl_mode}" : (
    local.is_mysql ? "tcp(${var.spec.datastore.mysql.host}:${local.mysql_port})/${var.spec.datastore.mysql.database}?parseTime=true" : null
  )

  datastore_username = local.is_postgres ? var.spec.datastore.postgres.username : (
    local.is_mysql ? var.spec.datastore.mysql.username : null
  )

  password_secret_name = local.is_postgres ? var.spec.datastore.postgres.password_secret.secret_name : (
    local.is_mysql ? var.spec.datastore.mysql.password_secret.secret_name : null
  )

  password_secret_key = local.is_postgres ? try(coalesce(var.spec.datastore.postgres.password_secret.secret_key, "password"), "password") : (
    local.is_mysql ? try(coalesce(var.spec.datastore.mysql.password_secret.secret_key, "password"), "password") : null
  )

  # SECRET DISCIPLINE (chart precedence, _helpers.tpl envConfig): each
  # credential field resolves its own branch independently, so mixing a
  # plain username with existingSecret+passwordKey is legal and exact —
  # username lands in the chart-owned `<fullname>-datastore-secret`
  # (usernames are not secrets), while the password rides a secretKeyRef
  # into the referenced existing Secret.
  #
  # MIGRATIONS RUN AS AN INIT CONTAINER, ALWAYS (`openfga migrate` is
  # idempotent with embedded migrations). The chart's default hook-Job
  # mode is deliberately not used: (1) it DEADLOCKS engines that wait on
  # rollout readiness — Helm --wait waits for the Deployment, whose
  # wait-for-migration init container waits for a post-install hook Job
  # that Helm only runs after --wait; (2) its hook list includes
  # post-delete, which dials the database during uninstall.
  datastore_block = {
    for k, v in {
      engine          = local.datastore_engine
      applyMigrations = true
      migrationType   = "initContainer"
      uri             = local.datastore_uri
      username        = local.datastore_username
      existingSecret  = local.password_secret_name
      secretKeys      = local.password_secret_name != null ? { passwordKey = local.password_secret_key } : null
      maxOpenConns    = try(var.spec.datastore.max_open_conns, null)
      maxIdleConns    = try(var.spec.datastore.max_idle_conns, null)
      connMaxIdleTime = try(var.spec.datastore.conn_max_idle_time, "") != "" ? var.spec.datastore.conn_max_idle_time : null
      connMaxLifetime = try(var.spec.datastore.conn_max_lifetime, "") != "" ? var.spec.datastore.conn_max_lifetime : null
    } : k => v if v != null
  }

  # migrate.timeout is consumed by the initContainer branch as
  # OPENFGA_TIMEOUT (chart-truth: deployment.yaml renders it into the
  # migrate-database init container's env) — how long `openfga migrate`
  # retries an unreachable database before failing the pod. Meaningless
  # on the memory arm (no migrations run).
  migrate_block = (!local.is_memory && try(var.spec.datastore.migration_timeout, "") != "") ? {
    timeout = var.spec.datastore.migration_timeout
  } : null

  # ---- replicas -----------------------------------------------------------------
  # Skipped under autoscaling (the chart omits the Deployment replicas
  # field entirely when autoscaling.enabled — the HPA owns the count).
  # The memory arm renders 1 EXPLICITLY: the chart forces it anyway
  # (ternary on engine == "memory" — each replica would hold its own
  # divergent authorization world), rendered here so the manifest states
  # the intent.
  replica_count = local.is_memory ? 1 : (
    try(var.spec.hpa.enabled, false) ? null : try(var.spec.replicas, null)
  )

  # ---- pre-shared authn keys ---------------------------------------------------------
  # Declared keys materialize into the module-owned
  # `<metadata.name>-authn-keys` Secret (authn_secret.tf); an existing
  # Secret is referenced by name. Either way only a Secret NAME ever
  # renders — the chart's plain-values key list (authn.preshared.keys),
  # which would render every key into the Deployment manifest, is
  # deliberately never used.
  materialize_authn_secret = length(try(var.spec.authn.preshared.keys, [])) > 0
  authn_keys_secret_name   = local.materialize_authn_secret ? "${var.metadata.name}-authn-keys" : ""

  preshared_keys_secret_ref = local.materialize_authn_secret ? local.authn_keys_secret_name : try(coalesce(var.spec.authn.preshared.existing_keys_secret_name, ""), "")

  authn_method = try(var.spec.authn.preshared, null) != null ? "preshared" : (
    try(var.spec.authn.oidc, null) != null ? "oidc" : null
  )

  # Unset renders NOTHING (server default: no authentication).
  authn_block = {
    for k, v in {
      method    = local.authn_method
      preshared = local.authn_method == "preshared" ? { keysSecret = local.preshared_keys_secret_ref } : null
      oidc = local.authn_method == "oidc" ? {
        issuer   = var.spec.authn.oidc.issuer
        audience = var.spec.authn.oidc.audience
      } : null
    } : k => v if v != null
  }

  # ---- telemetry ------------------------------------------------------------------------
  # metrics.enabled is rendered EXPLICITLY in both states (chart default
  # true) so the manifest states the intent; the ServiceMonitor requires
  # the Prometheus Operator CRDs — the install fails without them.
  metrics_enabled = try(coalesce(var.spec.metrics.enabled, true), true)

  telemetry_metrics_block = {
    for k, v in {
      enabled             = local.metrics_enabled
      serviceMonitor      = try(var.spec.metrics.service_monitor_enabled, false) ? { enabled = true } : null
      enableRPCHistograms = try(var.spec.metrics.enable_rpc_histograms, false) ? true : null
    } : k => v if v != null
  }

  # sampleRatio is typed as a NUMBER in the chart's closed schema — a
  # string would fail install validation; the proto pattern guarantees a
  # parseable 0.0–1.0 (Pulumi twin: strconv.ParseFloat).
  telemetry_trace_block = try(var.spec.tracing.enabled, false) ? {
    for k, v in {
      enabled     = true
      otlp        = { endpoint = var.spec.tracing.otlp_endpoint }
      sampleRatio = try(var.spec.tracing.sample_ratio, "") != "" ? tonumber(var.spec.tracing.sample_ratio) : null
    } : k => v if v != null
  } : null

  telemetry_block = {
    for k, v in {
      metrics = local.telemetry_metrics_block
      trace   = local.telemetry_trace_block
    } : k => v if v != null
  }

  # ---- log ------------------------------------------------------------------------------------
  log_block = {
    for k, v in {
      level  = try(var.spec.log.level, "") != "" ? var.spec.log.level : null
      format = try(var.spec.log.format, "") != "" ? var.spec.log.format : null
    } : k => v if v != null
  }

  # ---- check-query cache -----------------------------------------------------------------------
  check_query_cache_block = try(var.spec.tuning.check_query_cache, null) == null ? null : {
    for k, v in {
      enabled = try(var.spec.tuning.check_query_cache.enabled, false)
      limit   = try(var.spec.tuning.check_query_cache.limit, null)
      ttl     = try(var.spec.tuning.check_query_cache.ttl, "") != "" ? var.spec.tuning.check_query_cache.ttl : null
    } : k => v if v != null
  }

  # ---- autoscaling ------------------------------------------------------------------------------
  autoscaling_block = try(var.spec.hpa.enabled, false) ? {
    for k, v in {
      enabled                           = true
      minReplicas                       = try(var.spec.hpa.min_replicas, null)
      maxReplicas                       = try(var.spec.hpa.max_replicas, null)
      targetCPUUtilizationPercentage    = try(var.spec.hpa.target_cpu_utilization_percent, null)
      targetMemoryUtilizationPercentage = try(var.spec.hpa.target_memory_utilization_percent, null)
    } : k => v if v != null
  } : null

  # ---- shared renderers ---------------------------------------------------------------------------
  resources_block = {
    for k, v in {
      requests = try(var.spec.resources.requests, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.requests.cpu
          memory = var.spec.resources.requests.memory
        } : rk => rv if rv != null && rv != ""
      }
      limits = try(var.spec.resources.limits, null) == null ? null : {
        for rk, rv in {
          cpu    = var.spec.resources.limits.cpu
          memory = var.spec.resources.limits.memory
        } : rk => rv if rv != null && rv != ""
      }
    } : k => v if v != null
  }

  scheduling_tolerations = [
    for t in try(var.spec.scheduling.tolerations, []) : { for k, v in {
      key               = t.key != "" ? t.key : null
      operator          = t.operator != "" ? t.operator : null
      value             = t.value != "" ? t.value : null
      effect            = t.effect != "" ? t.effect : null
      tolerationSeconds = t.toleration_seconds
    } : k => v if v != null }
  ]

  # ---- typed chart values (Pulumi twin: buildHelmValues) --------------------------------------------
  # fullnameOverride pins the chart fullname to the resource name (and
  # is RE-PINNED as a third values document in helm_release.tf).
  #
  # PLAYGROUND: ALWAYS OFF. The chart ships its demo playground ENABLED
  # by default; this module unconditionally disables it. Verified at
  # OpenFGA v1.18.1: upstream turned the playground off by default for
  # security (GHSA-68m9-983m-f3v5), the server REFUSES TO START when the
  # playground combines with ANY authn method, and at this version it
  # binds 127.0.0.1 pod-local — the chart's playground Service port
  # cannot reach it anyway.
  #
  # Tuning scalars map 1:1 to top-level chart values and render ONLY
  # when set (chart `if` guards keep the server's own defaults); the two
  # MaxResults fields are `ne nil` guards in the chart — an explicit 0
  # (= unlimited) renders through the null-prune untouched. The
  # experimentals list REPLACES the server's own default experimental
  # set (server contract at v1.18.1).
  #
  # serviceAccount: only the annotations seam (cloud workload identity)
  # is modeled. serviceAccount.create=false is deliberately unsupported:
  # it would silently drop the Job-status RBAC the wait-for-migration
  # init container needs if anyone flips migrationType back to "job".
  typed_helm_values = {
    for k, v in {
      fullnameOverride = local.release_name
      replicaCount     = local.replica_count
      datastore        = local.datastore_block
      migrate          = local.migrate_block
      playground       = { enabled = false }
      authn            = length(local.authn_block) > 0 ? local.authn_block : null
      telemetry        = local.telemetry_block
      log              = length(local.log_block) > 0 ? local.log_block : null

      maxTuplesPerWrite             = try(var.spec.tuning.max_tuples_per_write, null)
      maxTypesPerAuthorizationModel = try(var.spec.tuning.max_types_per_authorization_model, null)
      maxChecksPerBatchCheck        = try(var.spec.tuning.max_checks_per_batch_check, null)
      listObjectsDeadline           = try(var.spec.tuning.list_objects_deadline, "") != "" ? var.spec.tuning.list_objects_deadline : null
      listObjectsMaxResults         = try(var.spec.tuning.list_objects_max_results, null)
      listUsersDeadline             = try(var.spec.tuning.list_users_deadline, "") != "" ? var.spec.tuning.list_users_deadline : null
      listUsersMaxResults           = try(var.spec.tuning.list_users_max_results, null)
      requestTimeout                = try(var.spec.tuning.request_timeout, "") != "" ? var.spec.tuning.request_timeout : null
      checkQueryCache               = local.check_query_cache_block
      experimentals                 = length(try(var.spec.tuning.experimentals, [])) > 0 ? var.spec.tuning.experimentals : null

      resources    = length(local.resources_block) > 0 ? local.resources_block : null
      autoscaling  = local.autoscaling_block
      nodeSelector = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
      tolerations  = length(local.scheduling_tolerations) > 0 ? local.scheduling_tolerations : null

      podExtraLabels = local.labels

      serviceAccount = length(try(var.spec.service_account_annotations, {})) > 0 ? {
        annotations = var.spec.service_account_annotations
      } : null
    } : k => v if v != null
  }
}
