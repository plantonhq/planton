# Computed values for the KubernetesTempo module. Every resolution here has
# an exact twin in the Pulumi module's locals.go / values.go — keep them in
# lockstep.
#
# SECRET DISCIPLINE (load-bearing): the chart renders the Tempo config into
# a ConfigMap. Declared object-store credentials therefore NEVER appear in
# these values: they travel as environment variables sourced from the
# referenced Secrets (tempo.extraEnv secretKeyRefs), the config references
# them as ${VAR} placeholders, and -config.expand-env=true expands them at
# process start.
#
# HCL DISCIPLINE: conditional entries use the null-prune idiom; optional
# nested blocks are read with try() (HCL's && does NOT short-circuit).

locals {
  helm_chart_name = "tempo"
  helm_chart_repo = "https://grafana-community.github.io/helm-charts"

  release_name  = var.metadata.name
  chart_version = coalesce(var.spec.chart_version, "2.2.3")
  namespace     = var.spec.namespace

  max_name_length    = 45
  name_within_budget = length(local.release_name) <= local.max_name_length

  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesTempo"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- storage backend ---------------------------------------------------
  backend_type = (
    try(var.spec.storage.s3, null) != null || try(var.spec.storage.r2, null) != null ? "s3" :
    try(var.spec.storage.gcs, null) != null ? "gcs" :
    try(var.spec.storage.azure, null) != null ? "azure" : "local"
  )
  s3_creds_declared = local.backend_type == "s3" && try(var.spec.storage.s3.credentials, null) != null

  # ---- the r2 arm (twin of the Pulumi module's r2.go) ---------------------
  # R2 speaks S3, so the trace backend is "s3" and the S3 values are
  # composed from R2's vocabulary. The host table is pkg/cloudflare/r2's
  # (the Go source of truth): a jurisdictional bucket is served only through
  # <account>.<jurisdiction>.r2.cloudflarestorage.com; Tempo's S3 client
  # takes the host without a scheme and speaks TLS by default. The key pair
  # goes into the module-owned `<name>-r2-credentials` Secret (main.tf).
  r2              = try(var.spec.storage.r2, null)
  r2_jurisdiction = local.r2 == null ? "default" : coalesce(try(coalesce(local.r2.jurisdiction), ""), "default")
  r2_endpoint = local.r2 == null ? null : (
    local.r2_jurisdiction == "default"
    ? "${local.r2.account_id}.r2.cloudflarestorage.com"
    : "${local.r2.account_id}.${local.r2_jurisdiction}.r2.cloudflarestorage.com"
  )
  r2_secret_name = "${local.release_name}-r2-credentials"
  r2_secret_data = local.r2 == null ? {} : {
    "access-key-id"     = local.r2.credentials.access_key_id
    "secret-access-key" = local.r2.credentials.secret_access_key
  }
  gcs_key_declared   = local.backend_type == "gcs" && try(var.spec.storage.gcs.service_account_key_secret, null) != null
  azure_key_declared = local.backend_type == "azure" && try(var.spec.storage.azure.account_key_secret, null) != null

  gcs_key_mount_path = "/var/secrets/gcs"
  gcs_key_volume     = "gcs-service-account"

  trace_s3 = local.r2 != null ? {
    bucket         = local.r2.bucket
    endpoint       = local.r2_endpoint
    region         = "auto"
    forcepathstyle = true
    access_key     = "$${TEMPO_S3_ACCESS_KEY_ID}"
    secret_key     = "$${TEMPO_S3_SECRET_ACCESS_KEY}"
    } : local.backend_type != "s3" ? null : { for k, v in {
      bucket         = var.spec.storage.s3.bucket
      endpoint       = var.spec.storage.s3.endpoint
      region         = try(var.spec.storage.s3.region, "") != "" ? var.spec.storage.s3.region : null
      forcepathstyle = try(var.spec.storage.s3.force_path_style, false) ? true : null
      insecure       = try(var.spec.storage.s3.insecure, false) ? true : null
      access_key     = local.s3_creds_declared ? "$${TEMPO_S3_ACCESS_KEY_ID}" : null
      secret_key     = local.s3_creds_declared ? "$${TEMPO_S3_SECRET_ACCESS_KEY}" : null
  } : k => v if v != null }

  trace_azure = local.backend_type != "azure" ? null : { for k, v in {
    storage_account_name = var.spec.storage.azure.account_name
    container_name       = var.spec.storage.azure.container
    storage_account_key  = local.azure_key_declared ? "$${TEMPO_AZURE_ACCOUNT_KEY}" : null
    use_federated_token  = local.azure_key_declared ? null : true
  } : k => v if v != null }

  trace_block = { for k, v in {
    backend = local.backend_type
    wal     = { path = "/var/tempo/wal" }
    local   = local.backend_type == "local" ? { path = "/var/tempo/traces" } : null
    s3      = local.trace_s3
    gcs     = local.backend_type == "gcs" ? { bucket_name = var.spec.storage.gcs.bucket } : null
    azure   = local.trace_azure
  } : k => v if v != null }

  # ---- credential env + volumes ------------------------------------------
  credential_env = concat(
    local.r2 != null ? [
      {
        name      = "TEMPO_S3_ACCESS_KEY_ID"
        valueFrom = { secretKeyRef = { name = local.r2_secret_name, key = "access-key-id" } }
      },
      {
        name      = "TEMPO_S3_SECRET_ACCESS_KEY"
        valueFrom = { secretKeyRef = { name = local.r2_secret_name, key = "secret-access-key" } }
      },
    ] : [],
    local.s3_creds_declared ? [
      {
        name = "TEMPO_S3_ACCESS_KEY_ID"
        valueFrom = { secretKeyRef = {
          name = var.spec.storage.s3.credentials.access_key_id_secret.name
          key  = var.spec.storage.s3.credentials.access_key_id_secret.key
        } }
      },
      {
        name = "TEMPO_S3_SECRET_ACCESS_KEY"
        valueFrom = { secretKeyRef = {
          name = var.spec.storage.s3.credentials.secret_access_key_secret.name
          key  = var.spec.storage.s3.credentials.secret_access_key_secret.key
        } }
      },
    ] : [],
    local.azure_key_declared ? [
      {
        name = "TEMPO_AZURE_ACCOUNT_KEY"
        valueFrom = { secretKeyRef = {
          name = var.spec.storage.azure.account_key_secret.name
          key  = var.spec.storage.azure.account_key_secret.key
        } }
      },
    ] : [],
    local.gcs_key_declared ? [
      {
        name  = "GOOGLE_APPLICATION_CREDENTIALS"
        value = "${local.gcs_key_mount_path}/${var.spec.storage.gcs.service_account_key_secret.key}"
      },
    ] : [],
  )
  extra_volume_mounts = local.gcs_key_declared ? [
    { name = local.gcs_key_volume, mountPath = local.gcs_key_mount_path, readOnly = true },
  ] : []
  extra_volumes = local.gcs_key_declared ? [
    { name = local.gcs_key_volume, secret = { secretName = var.spec.storage.gcs.service_account_key_secret.name } },
  ] : []

  # ---- receivers ---------------------------------------------------------
  # OTLP always on; the four legacy Jaeger protocols opt-in.
  receivers = merge(
    {
      otlp = { protocols = {
        grpc = { endpoint = "0.0.0.0:4317" }
        http = { endpoint = "0.0.0.0:4318" }
      } }
    },
    try(var.spec.jaeger_receivers_enabled, false) ? {
      jaeger = { protocols = {
        grpc           = { endpoint = "0.0.0.0:14250" }
        thrift_binary  = { endpoint = "0.0.0.0:6832" }
        thrift_compact = { endpoint = "0.0.0.0:6831" }
        thrift_http    = { endpoint = "0.0.0.0:14268" }
      } }
    } : {}
  )

  # ---- metrics generator -------------------------------------------------
  mg_enabled              = try(var.spec.metrics_generator.enabled, false)
  mg_remote_write_url_raw = local.mg_enabled ? try(var.spec.metrics_generator.remote_write_url, "") : ""
  # Append Prometheus' standard remote-write path when the URL carries none
  # (a bare Service endpoint like the stack's prometheus_endpoint output).
  mg_remote_write_url = local.mg_remote_write_url_raw == "" ? "" : (
    can(regex("://[^/]+/", local.mg_remote_write_url_raw)) ? local.mg_remote_write_url_raw : "${local.mg_remote_write_url_raw}/api/v1/write"
  )
  mg_processors_raw = local.mg_enabled ? try(var.spec.metrics_generator.processors, []) : []
  mg_processors = length(local.mg_processors_raw) > 0 ? [
    for p in local.mg_processors_raw : p == "service_graphs" ? "service-graphs" : "span-metrics"
  ] : ["service-graphs", "span-metrics"]

  # ---- resources ---------------------------------------------------------
  tempo_resources = try(var.spec.resources, null) == null ? null : { for rk, rv in {
    requests = try(var.spec.resources.requests, null) == null ? null : { for qk, qv in {
      cpu    = try(var.spec.resources.requests.cpu, "") != "" ? var.spec.resources.requests.cpu : null
      memory = try(var.spec.resources.requests.memory, "") != "" ? var.spec.resources.requests.memory : null
    } : qk => qv if qv != null }
    limits = try(var.spec.resources.limits, null) == null ? null : { for lk, lv in {
      cpu    = try(var.spec.resources.limits.cpu, "") != "" ? var.spec.resources.limits.cpu : null
      memory = try(var.spec.resources.limits.memory, "") != "" ? var.spec.resources.limits.memory : null
    } : lk => lv if lv != null }
  } : rk => rv if rv != null && rv != {} }

  # ---- the tempo block ---------------------------------------------------
  tempo_block = { for k, v in {
    retention           = coalesce(var.spec.retention, "24h")
    reportingEnabled    = try(var.spec.usage_reporting, false) ? null : false
    multitenancyEnabled = var.spec.multi_tenancy_enabled ? true : null
    receivers           = local.receivers
    storage             = { trace = local.trace_block }
    resources           = local.tempo_resources
    metricsGenerator = local.mg_enabled ? {
      enabled        = true
      remoteWriteUrl = local.mg_remote_write_url
    } : null
    overrides = local.mg_enabled ? {
      defaults = { metrics_generator = { processors = local.mg_processors } }
    } : null
    extraEnv          = length(local.credential_env) > 0 ? local.credential_env : null
    extraArgs         = length(local.credential_env) > 0 ? { "config.expand-env" = "true" } : null
    extraVolumeMounts = length(local.extra_volume_mounts) > 0 ? local.extra_volume_mounts : null
    pullSecrets       = length(var.spec.image_pull_secrets) > 0 ? var.spec.image_pull_secrets : null
  } : k => v if v != null }

  # ---- persistence -------------------------------------------------------
  persistence_block = var.spec.ephemeral ? { enabled = false } : { for k, v in {
    enabled          = true
    size             = coalesce(var.spec.disk_size, "10Gi")
    storageClassName = try(var.spec.storage_class, "") != "" ? var.spec.storage_class : null
  } : k => v if v != null }

  # ---- tempoQuery --------------------------------------------------------
  tempo_query_block = var.spec.tempo_query_enabled ? { for k, v in {
    enabled    = true
    repository = var.spec.image_registry != "" ? "${var.spec.image_registry}/grafana/tempo-query" : null
  } : k => v if v != null } : null

  # ---- typed chart values (twin of buildHelmValues) ----------------------
  helm_values = { for k, v in {
    fullnameOverride = local.release_name
    # Tempo reads its object-store credentials only at start: the Secret's
    # fingerprint changes the pod template when the key rotates.
    # sha256(jsonencode(map)) is the Pulumi twin's credentialsChecksum.
    podAnnotations = local.r2 != null ? { "checksum/credentials" = sha256(jsonencode(local.r2_secret_data)) } : null
    replicas       = coalesce(var.spec.replicas, 1)
    persistence    = local.persistence_block
    tempo          = local.tempo_block
    tempoQuery     = local.tempo_query_block
    serviceMonitor = var.spec.service_monitor_enabled ? { enabled = true } : null
    global         = var.spec.image_registry != "" ? { imageRegistry = var.spec.image_registry } : null
    nodeSelector   = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
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
    extraVolumes      = length(local.extra_volumes) > 0 ? local.extra_volumes : null
  } : k => v if v != null }
}
