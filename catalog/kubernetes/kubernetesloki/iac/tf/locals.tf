# Computed values for the KubernetesLoki module. Every resolution here has
# an exact twin in the Pulumi module's locals.go / values.go — keep them in
# lockstep.
#
# SECRET DISCIPLINE (load-bearing): the chart renders the Loki configuration
# into a ConfigMap visible to anyone with read access on the namespace.
# Declared object-store credentials therefore NEVER appear in these values:
# they travel as environment variables sourced from the referenced Secrets
# (defaults.extraEnv secretKeyRefs), the config references them as ${VAR}
# placeholders, and -config.expand-env=true makes Loki expand them at
# process start.
#
# HCL DISCIPLINE (applies to every conditional object below): conditional
# entries are written as `key = cond ? value : null` inside ONE object
# literal, pruned with `{ for k, v in {...} : k => v if v != null }`.
# `cond ? {...} : {}` ternaries fail plan-time type unification when
# branches carry different attributes. Optional nested blocks are read with
# try() (HCL's && does NOT short-circuit).

locals {
  # Chart identity — must stay byte-identical with the Pulumi module's
  # vars. KNOW THIS about the repo URL: the loki chart's canonical home is
  # the grafana-community index (the vendor moved its community charts
  # there — the same move the grafana chart made). Never "fix" this back.
  helm_chart_name = "loki"
  helm_chart_repo = "https://grafana-community.github.io/helm-charts"

  # Release name — metadata.name, NOT a fixed chart name: several Loki
  # instances can coexist. fullnameOverride pins every chart child name to
  # it.
  release_name = var.metadata.name

  # Chart version resolved to the pinned default when unset. Chart 18.5.4
  # ships Loki 3.7.4.
  chart_version = coalesce(var.spec.chart_version, "18.5.4")

  namespace = var.spec.namespace

  # The chart composes child names like `<fullname>-backend-headless` (16
  # chars of suffix) and truncates the fullname at 63 — a resource name
  # longer than this corrupts the naming contract the outputs promise. The
  # helm_release precondition (main.tf) fails the plan loudly; this flag is
  # its condition (twin: the Pulumi module's MaxNameLength guard).
  max_name_length    = 40
  name_within_budget = length(local.release_name) <= local.max_name_length

  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesLoki"
    },
    var.metadata.id != null && var.metadata.id != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    var.metadata.org != null && var.metadata.org != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # ---- storage backend resolution ----------------------------------------
  # One derivation feeds three surfaces: loki.storage.type, the derived
  # schema's object_store, and the compactor's delete_request_store.
  backend_type = (
    try(var.spec.storage.s3, null) != null || try(var.spec.storage.r2, null) != null ? "s3" :
    try(var.spec.storage.gcs, null) != null ? "gcs" :
    try(var.spec.storage.azure, null) != null ? "azure" : "filesystem"
  )

  # ---- deployment mode + write-path replicas -----------------------------
  is_ssd = try(var.spec.simple_scalable, null) != null

  write_path_replicas = local.is_ssd ? coalesce(try(var.spec.simple_scalable.write_replicas, null), 3) : (
    try(var.spec.monolithic, null) != null ? coalesce(try(var.spec.monolithic.replicas, null), 1) : 1
  )
  # Loki refuses a replication_factor above the number of ingesting
  # replicas; 3 is the HA ceiling worth defaulting to.
  replication_factor = min(local.write_path_replicas, 3)

  # Per-mode disk/class/resources.
  mono_disk_size     = coalesce(try(var.spec.monolithic.disk_size, null), "10Gi")
  mono_storage_class = try(var.spec.monolithic.storage_class, "")
  ssd_disk_size      = coalesce(try(var.spec.simple_scalable.disk_size, null), "10Gi")
  ssd_storage_class  = try(var.spec.simple_scalable.storage_class, "")

  # Per-scope resource maps. Inlined (HCL has no user functions) with the
  # null-prune idiom; null when nothing is declared so the chart keeps its
  # defaults (twin: the Pulumi module's resourcesMap).
  mono_resources = local.is_ssd ? null : (try(var.spec.monolithic.resources, null) == null ? null : { for rk, rv in {
    requests = try(var.spec.monolithic.resources.requests, null) == null ? null : { for qk, qv in {
      cpu    = try(var.spec.monolithic.resources.requests.cpu, "") != "" ? var.spec.monolithic.resources.requests.cpu : null
      memory = try(var.spec.monolithic.resources.requests.memory, "") != "" ? var.spec.monolithic.resources.requests.memory : null
    } : qk => qv if qv != null }
    limits = try(var.spec.monolithic.resources.limits, null) == null ? null : { for lk, lv in {
      cpu    = try(var.spec.monolithic.resources.limits.cpu, "") != "" ? var.spec.monolithic.resources.limits.cpu : null
      memory = try(var.spec.monolithic.resources.limits.memory, "") != "" ? var.spec.monolithic.resources.limits.memory : null
    } : lk => lv if lv != null }
  } : rk => rv if rv != null && rv != {} })

  ssd_resources = !local.is_ssd ? null : (try(var.spec.simple_scalable.resources, null) == null ? null : { for rk, rv in {
    requests = try(var.spec.simple_scalable.resources.requests, null) == null ? null : { for qk, qv in {
      cpu    = try(var.spec.simple_scalable.resources.requests.cpu, "") != "" ? var.spec.simple_scalable.resources.requests.cpu : null
      memory = try(var.spec.simple_scalable.resources.requests.memory, "") != "" ? var.spec.simple_scalable.resources.requests.memory : null
    } : qk => qv if qv != null }
    limits = try(var.spec.simple_scalable.resources.limits, null) == null ? null : { for lk, lv in {
      cpu    = try(var.spec.simple_scalable.resources.limits.cpu, "") != "" ? var.spec.simple_scalable.resources.limits.cpu : null
      memory = try(var.spec.simple_scalable.resources.limits.memory, "") != "" ? var.spec.simple_scalable.resources.limits.memory : null
    } : lk => lv if lv != null }
  } : rk => rv if rv != null && rv != {} })

  gateway_resources = try(var.spec.gateway.resources, null) == null ? null : { for rk, rv in {
    requests = try(var.spec.gateway.resources.requests, null) == null ? null : { for qk, qv in {
      cpu    = try(var.spec.gateway.resources.requests.cpu, "") != "" ? var.spec.gateway.resources.requests.cpu : null
      memory = try(var.spec.gateway.resources.requests.memory, "") != "" ? var.spec.gateway.resources.requests.memory : null
    } : qk => qv if qv != null }
    limits = try(var.spec.gateway.resources.limits, null) == null ? null : { for lk, lv in {
      cpu    = try(var.spec.gateway.resources.limits.cpu, "") != "" ? var.spec.gateway.resources.limits.cpu : null
      memory = try(var.spec.gateway.resources.limits.memory, "") != "" ? var.spec.gateway.resources.limits.memory : null
    } : lk => lv if lv != null }
  } : rk => rv if rv != null && rv != {} }

  # Persistence blocks (pruned) for each tier — computed once, referenced
  # conditionally below so the workload-block ternaries never put
  # differently-shaped objects on their two branches (an HCL
  # type-unification error).
  mono_persistence = { for pk, pv in {
    enabled      = true
    size         = local.mono_disk_size
    storageClass = local.mono_storage_class != "" ? local.mono_storage_class : null
  } : pk => pv if pv != null }
  ssd_persistence = { for pk, pv in {
    enabled      = true
    size         = local.ssd_disk_size
    storageClass = local.ssd_storage_class != "" ? local.ssd_storage_class : null
  } : pk => pv if pv != null }

  # Workload blocks: exactly one mode renders live, every other mode's
  # workloads are explicitly zeroed (the half-running-mode trap). Each is a
  # SINGLE pruned object with per-key conditionals — never a ternary whose
  # branches carry different attribute sets (the grafana exemplar pattern).
  single_binary = { for k, v in {
    replicas    = local.is_ssd ? 0 : (try(var.spec.monolithic, null) != null ? coalesce(try(var.spec.monolithic.replicas, null), 1) : 1)
    persistence = local.is_ssd ? null : local.mono_persistence
    resources   = local.is_ssd ? null : local.mono_resources
  } : k => v if v != null }

  write_block = { for k, v in {
    replicas    = local.is_ssd ? coalesce(try(var.spec.simple_scalable.write_replicas, null), 3) : 0
    persistence = local.is_ssd ? local.ssd_persistence : null
    resources   = local.is_ssd ? local.ssd_resources : null
  } : k => v if v != null }

  backend_block = { for k, v in {
    replicas    = local.is_ssd ? coalesce(try(var.spec.simple_scalable.backend_replicas, null), 3) : 0
    persistence = local.is_ssd ? local.ssd_persistence : null
    resources   = local.is_ssd ? local.ssd_resources : null
  } : k => v if v != null }

  read_block = { for k, v in {
    replicas  = local.is_ssd ? coalesce(try(var.spec.simple_scalable.read_replicas, null), 3) : 0
    resources = local.is_ssd ? local.ssd_resources : null
  } : k => v if v != null }

  # The chart's microservices-mode workloads, zeroed in EVERY rendering.
  distributed_components = [
    "ingester", "querier", "queryFrontend", "queryScheduler", "distributor",
    "compactor", "indexGateway", "bloomCompactor", "bloomGateway",
  ]
  distributed_zeroed = { for c in local.distributed_components : c => { replicas = 0 } }

  # ---- schema ------------------------------------------------------------
  schema_from = coalesce(var.spec.schema_from_date != "" ? var.spec.schema_from_date : null, "2024-04-01")

  # ---- the r2 arm (twin of the Pulumi module's r2.go) ---------------------
  # R2 speaks S3, so backend_type is "s3" and the S3 values are composed
  # from R2's vocabulary. The host table is pkg/cloudflare/r2's (the Go
  # source of truth): a jurisdictional bucket is served only through
  # <account>.<jurisdiction>.r2.cloudflarestorage.com. The key pair goes
  # into the module-owned `<name>-r2-credentials` Secret (main.tf).
  r2              = try(var.spec.storage.r2, null)
  r2_jurisdiction = local.r2 == null ? "default" : coalesce(try(coalesce(local.r2.jurisdiction), ""), "default")
  r2_endpoint = local.r2 == null ? null : (
    local.r2_jurisdiction == "default"
    ? "https://${local.r2.account_id}.r2.cloudflarestorage.com"
    : "https://${local.r2.account_id}.${local.r2_jurisdiction}.r2.cloudflarestorage.com"
  )
  r2_secret_name = "${local.release_name}-r2-credentials"
  r2_secret_data = local.r2 == null ? {} : {
    "access-key-id"     = local.r2.credentials.access_key_id
    "secret-access-key" = local.r2.credentials.secret_access_key
  }

  # ---- object-store credential env + volumes -----------------------------
  s3_creds_declared  = local.backend_type == "s3" && try(var.spec.storage.s3.credentials, null) != null
  gcs_key_declared   = local.backend_type == "gcs" && try(var.spec.storage.gcs.service_account_key_secret, null) != null
  azure_key_declared = local.backend_type == "azure" && try(var.spec.storage.azure.account_key_secret, null) != null

  gcs_key_mount_path = "/var/secrets/gcs"
  gcs_key_volume     = "gcs-service-account"

  credential_env = concat(
    local.r2 != null ? [
      {
        name      = "LOKI_S3_ACCESS_KEY_ID"
        valueFrom = { secretKeyRef = { name = local.r2_secret_name, key = "access-key-id" } }
      },
      {
        name      = "LOKI_S3_SECRET_ACCESS_KEY"
        valueFrom = { secretKeyRef = { name = local.r2_secret_name, key = "secret-access-key" } }
      },
    ] : [],
    local.s3_creds_declared ? [
      {
        name = "LOKI_S3_ACCESS_KEY_ID"
        valueFrom = { secretKeyRef = {
          name = var.spec.storage.s3.credentials.access_key_id_secret.name
          key  = var.spec.storage.s3.credentials.access_key_id_secret.key
        } }
      },
      {
        name = "LOKI_S3_SECRET_ACCESS_KEY"
        valueFrom = { secretKeyRef = {
          name = var.spec.storage.s3.credentials.secret_access_key_secret.name
          key  = var.spec.storage.s3.credentials.secret_access_key_secret.key
        } }
      },
    ] : [],
    local.azure_key_declared ? [
      {
        name = "LOKI_AZURE_ACCOUNT_KEY"
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

  extra_volumes = local.gcs_key_declared ? [
    {
      name   = local.gcs_key_volume
      secret = { secretName = var.spec.storage.gcs.service_account_key_secret.name }
    },
  ] : []
  extra_volume_mounts = local.gcs_key_declared ? [
    {
      name      = local.gcs_key_volume
      mountPath = local.gcs_key_mount_path
      readOnly  = true
    },
  ] : []

  # ---- storage block -----------------------------------------------------
  # One object for both arms that speak S3 (s3 and r2), each attribute
  # choosing its own value: a conditional between two differently shaped
  # objects makes OpenTofu unify them to map(string), and the chart copies
  # these values into Loki's config verbatim -- a quoted "true" for
  # s3ForcePathStyle is a string Loki refuses to parse, and it never starts.
  storage_s3 = local.backend_type != "s3" ? null : { for k, v in {
    endpoint         = local.r2 != null ? local.r2_endpoint : (try(var.spec.storage.s3.endpoint, "") != "" ? var.spec.storage.s3.endpoint : null)
    region           = local.r2 != null ? "auto" : (try(var.spec.storage.s3.region, "") != "" ? var.spec.storage.s3.region : null)
    s3ForcePathStyle = local.r2 != null || try(var.spec.storage.s3.force_path_style, false) ? true : null
    insecure         = local.r2 == null && try(var.spec.storage.s3.insecure, false) ? true : null
    accessKeyId      = local.r2 != null || local.s3_creds_declared ? "$${LOKI_S3_ACCESS_KEY_ID}" : null
    secretAccessKey  = local.r2 != null || local.s3_creds_declared ? "$${LOKI_S3_SECRET_ACCESS_KEY}" : null
  } : k => v if v != null }

  storage_azure = local.backend_type != "azure" ? null : { for k, v in {
    accountName       = var.spec.storage.azure.account_name
    accountKey        = local.azure_key_declared ? "$${LOKI_AZURE_ACCOUNT_KEY}" : null
    useFederatedToken = local.azure_key_declared ? null : true
  } : k => v if v != null }

  storage_block = { for k, v in {
    type = local.backend_type
    bucketNames = local.r2 != null ? {
      chunks = local.r2.bucket
      ruler  = local.r2.bucket
      } : local.backend_type == "s3" ? {
      chunks = var.spec.storage.s3.bucket
      ruler  = try(var.spec.storage.s3.ruler_bucket, "") != "" ? var.spec.storage.s3.ruler_bucket : var.spec.storage.s3.bucket
      } : local.backend_type == "gcs" ? {
      chunks = var.spec.storage.gcs.bucket
      ruler  = try(var.spec.storage.gcs.ruler_bucket, "") != "" ? var.spec.storage.gcs.ruler_bucket : var.spec.storage.gcs.bucket
      } : local.backend_type == "azure" ? {
      chunks = var.spec.storage.azure.container
      ruler  = try(var.spec.storage.azure.ruler_container, "") != "" ? var.spec.storage.azure.ruler_container : var.spec.storage.azure.container
    } : null
    s3    = local.storage_s3
    azure = local.storage_azure
  } : k => v if v != null }

  # ---- limits + retention ------------------------------------------------
  limits_config = { for k, v in {
    retention_period            = var.spec.retention_period != "" ? var.spec.retention_period : null
    ingestion_rate_mb           = try(var.spec.limits.ingestion_rate_mb, null)
    ingestion_burst_size_mb     = try(var.spec.limits.ingestion_burst_size_mb, null)
    max_global_streams_per_user = try(var.spec.limits.max_global_streams_per_user, null)
    max_query_lookback          = try(var.spec.limits.max_query_lookback, "") != "" ? var.spec.limits.max_query_lookback : null
  } : k => v if v != null }

  compactor_block = var.spec.retention_period != "" ? {
    retention_enabled    = true
    delete_request_store = local.backend_type
  } : null

  # ---- ruler -------------------------------------------------------------
  ruler_enabled = try(var.spec.ruler.enabled, false)
  ruler_config = local.ruler_enabled ? {
    alertmanager_url = try(var.spec.ruler.alertmanager_url, "")
    storage = {
      type  = "local"
      local = { directory = "/rules" }
    }
  } : null

  # ---- multi-tenancy -----------------------------------------------------
  mt_enabled      = try(var.spec.multi_tenancy.enabled, false)
  mt_tenants      = try(var.spec.multi_tenancy.tenants, [])
  mt_existing_htp = try(var.spec.multi_tenancy.existing_htpasswd_secret, "")
  loki_tenants = local.mt_enabled && length(local.mt_tenants) > 0 ? [
    for t in local.mt_tenants : {
      name         = t.name
      passwordHash = t.password_hash
    }
  ] : null

  # ---- the loki config block ---------------------------------------------
  loki_config = { for k, v in {
    auth_enabled = local.mt_enabled
    commonConfig = { replication_factor = local.replication_factor }
    schemaConfig = {
      configs = [
        {
          from         = local.schema_from
          store        = "tsdb"
          object_store = local.backend_type
          schema       = "v13"
          index        = { prefix = "loki_index_", period = "24h" }
        }
      ]
    }
    analytics     = try(var.spec.usage_reporting, false) ? null : { reporting_enabled = false }
    storage       = local.storage_block
    limits_config = length(local.limits_config) > 0 ? local.limits_config : null
    compactor     = local.compactor_block
    rulerConfig   = local.ruler_config
    tenants       = local.loki_tenants
    # Loki reads its object-store credentials only at start: the Secret's
    # fingerprint changes every Loki pod's template when the key rotates.
    # sha256(jsonencode(map)) is the Pulumi twin's credentialsChecksum.
    podAnnotations = local.r2 != null ? { "checksum/credentials" = sha256(jsonencode(local.r2_secret_data)) } : null
  } : k => v if v != null }

  # ---- gateway -----------------------------------------------------------
  gateway_enabled = try(var.spec.gateway.enabled, null) != null ? var.spec.gateway.enabled : true
  gateway_block = { for k, v in {
    enabled   = local.gateway_enabled ? null : false
    replicas  = try(var.spec.gateway.replicas, null) != null && try(var.spec.gateway.replicas, 1) != 1 ? var.spec.gateway.replicas : null
    resources = local.gateway_resources
    basicAuth = local.mt_enabled && (length(local.mt_tenants) > 0 || local.mt_existing_htp != "") ? { for bk, bv in {
      enabled        = true
      existingSecret = local.mt_existing_htp != "" ? local.mt_existing_htp : null
    } : bk => bv if bv != null } : null
  } : k => v if v != null }

  # ---- caches ------------------------------------------------------------
  chunks_cache = { for k, v in {
    enabled         = try(var.spec.caching.chunks_cache_enabled, null) != null && !var.spec.caching.chunks_cache_enabled ? false : null
    allocatedMemory = try(var.spec.caching.chunks_cache_memory_mb, null)
  } : k => v if v != null }
  results_cache = { for k, v in {
    enabled         = try(var.spec.caching.results_cache_enabled, null) != null && !var.spec.caching.results_cache_enabled ? false : null
    allocatedMemory = try(var.spec.caching.results_cache_memory_mb, null)
  } : k => v if v != null }

  # ---- defaults (shared Loki-workload scheduling + credential env) -------
  defaults_block = { for k, v in {
    extraEnv          = length(local.credential_env) > 0 ? local.credential_env : null
    extraArgs         = length(local.credential_env) > 0 ? ["-config.expand-env=true"] : null
    extraVolumes      = length(local.extra_volumes) > 0 ? local.extra_volumes : null
    extraVolumeMounts = length(local.extra_volume_mounts) > 0 ? local.extra_volume_mounts : null
    nodeSelector      = length(try(var.spec.scheduling.node_selector, {})) > 0 ? var.spec.scheduling.node_selector : null
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
  } : k => v if v != null }

  # ---- images ------------------------------------------------------------
  # The chart MIXES image forms: loki/gateway/canary/sidecar images are
  # SPLIT (registry + repository — global.imageRegistry overrides all their
  # registries at once) while the memcached caches run the docker-library
  # `memcached` image (repository-only, COMBINED form) the global override
  # does not reach — its repository is re-pointed explicitly. Twin: the
  # Pulumi module's image handling.
  global_block    = var.spec.image_registry != "" ? { imageRegistry = var.spec.image_registry } : null
  memcached_block = var.spec.image_registry != "" ? { image = { repository = "${var.spec.image_registry}/memcached" } } : null

  # ---- typed chart values (twin of the Pulumi module's buildHelmValues) --
  helm_values = { for k, v in {
    fullnameOverride = local.release_name
    deploymentMode   = local.is_ssd ? "SimpleScalable" : "Monolithic"

    singleBinary = local.single_binary
    write        = local.write_block
    read         = local.read_block
    backend      = local.backend_block

    loki = local.loki_config

    gateway      = length(local.gateway_block) > 0 ? local.gateway_block : null
    chunksCache  = length(local.chunks_cache) > 0 ? local.chunks_cache : null
    resultsCache = length(local.results_cache) > 0 ? local.results_cache : null
    lokiCanary   = try(var.spec.canary_enabled, null) != null && !var.spec.canary_enabled ? { enabled = false } : null
    # The chart's Helm test reads the canary's metrics and refuses to render without it.
    test             = try(var.spec.canary_enabled, null) != null && !var.spec.canary_enabled ? { enabled = false } : null
    monitoring       = var.spec.service_monitor_enabled ? { serviceMonitor = { enabled = true } } : null
    global           = local.global_block
    memcached        = local.memcached_block
    imagePullSecrets = length(var.spec.image_pull_secrets) > 0 ? var.spec.image_pull_secrets : null
    defaults         = length(local.defaults_block) > 0 ? local.defaults_block : null
  } : k => v if v != null }
}
