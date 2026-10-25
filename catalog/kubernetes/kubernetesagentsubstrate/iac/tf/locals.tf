# Computed values for the KubernetesAgentSubstrate module. Every resolution
# here has an exact twin in the Pulumi module (vars.go, settings.go,
# config.go, workloads.go); a gated parity test plans this module and holds
# its documents and outputs equal to the Pulumi composition.
#
# DEFAULTS: every default equals the spec's declared default, which equals
# upstream's own, and is resolved here so the module installs the same thing
# whether or not the platform's defaulting ran. A flag is written onto
# upstream's arguments only when its value differs, so upstream's defaults
# stay upstream's.
#
# HCL DISCIPLINE: optional blocks are read with try(); conditional keys and
# list items are contributed as filtered candidate lists or merge() of
# `cond ? { key = value } : {}` singletons, never a conditional between two
# differently shaped objects.

locals {
  # ---- the upstream contract (vars.go) ------------------------------------------
  # The agent-substrate/substrate release this module composes: its install
  # shape, the version installed when spec.version is empty, and the release
  # whose resource definitions the kind applies. Any release Planton publishes
  # with this shape installs with this module (images.json's shape names it).
  # MUST equal the Pulumi module's vars.Release (a test holds them equal).
  release                    = "v0.6.0"
  bundle_name                = "agent-substrate"
  namespace                  = "ate-system"
  pod_certificates_namespace = "podcertificate-controller-system"
  kind_name                  = "KubernetesAgentSubstrate"
  owner_label                = "planton.ai/kind"

  api_server_name          = "ate-api-server"
  controller_name          = "ate-controller"
  node_agent_name          = "atelet"
  router_name              = "atenet-router"
  egress_name              = "atenet-egress"
  pod_certificates_name    = "podcertificate-controller"
  credential_provider_name = "k8s-credential-provider"
  postgres_name            = "postgres"
  identity_server_name     = "ate-idp-server"
  postgres_volume_claim    = "data-postgres-0"

  api_server_host        = "api.ate-system.svc"
  api_trust_bundle_name  = "servicedns.podcert.ate.dev:identity:primary-bundle"
  api_trust_bundle_sign  = "servicedns.podcert.ate.dev/identity"
  router_client_signer   = "podidentity.podcert.ate.dev/identity"
  api_access_label_key   = "ate.dev/api-access"
  api_access_label_value = "true"

  version_annotation        = "planton.ai/agent-substrate-version"
  minor_label               = "planton.ai/agent-substrate-minor"
  env_hash_annotation       = "ate.dev/env-hash"
  config_hash_annotation    = "planton.ai/config-hash"
  postgres_setup_config_map = "postgres-setup"
  postgres_setup_mount_path = "/etc/postgres-setup"

  release_base_url = "https://downloads.planton.ai/releases/agent-substrate/"

  # The release installed (settings.go version), and where its files are.
  version     = coalesce(try(var.spec.version, null), local.release)
  version_url = "${local.release_base_url}${local.version}/"

  # The release's seal (vars.go fileChecksums), read first.
  checksums_file = "SHA256SUMS"

  # The release files the composition reads, by role (vars.go). images.json
  # is read on its own, before the rest, because it names the release's
  # install shape.
  images_file = "images.json"
  release_files = {
    namespace           = "manifests/ate-install/ate-system-namespace.yaml"
    controller_role     = "manifests/ate-install/generated/role.yaml"
    pod_certificates    = "manifests/ate-install/pod-certificate-controller.yaml"
    sandbox_validation  = "manifests/ate-install/sandboxconfig-validation.yaml"
    sandbox_gvisor      = "manifests/ate-install/sandboxconfig-gvisor.yaml"
    otel_config         = "manifests/ate-install/ate-otel-config.yaml"
    postgres            = "manifests/ate-install/postgres/postgres.yaml"
    postgres_size10     = "manifests/ate-install/postgres-size10/postgres-config-patch.yaml"
    podcert_size10      = "manifests/ate-install/podcert-size10/kustomization.yaml"
    api_server          = "manifests/ate-install/ate-api-server.yaml"
    controller          = "manifests/ate-install/ate-controller.yaml"
    identity_server     = "manifests/ate-install/ate-idp-server.yaml"
    node_agent          = "manifests/ate-install/atelet.yaml"
    router              = "manifests/ate-install/atenet-router.yaml"
    router_monitoring   = "manifests/ate-install/atenet-router-monitoring.yaml"
    network_policies    = "manifests/ate-install/ate-system-networkpolicies.yaml"
    egress              = "manifests/ate-install/atenet-egress.yaml"
    cloudsql_patch      = "manifests/ate-install/cloudsql/proxy-sidecar-patch.yaml"
    agentgateway        = "manifests/ate-install/components/agentgateway/kustomization.yaml"
    agentgateway_config = "manifests/ate-install/components/agentgateway/configmap.yaml"
    agentgateway_egress = "manifests/ate-install/components/agentgateway-egress/kustomization.yaml"
    cordon              = "manifests/ate-install/components/cordon-control-plane/kustomization.yaml"
    credential_provider = "manifests/egress-credential-injection/k8s-credential-provider.yaml"
    setup_sql           = "setup.sql"
  }

  # ---- who owns the namespace ---------------------------------------------------
  requested_namespace = var.spec.namespace
  create_namespace    = try(var.spec.create_namespace, false)

  # Identity labels on the namespaces and Secrets this module creates; the
  # kind label is also the ownership mark the namespace check reads.
  identity_labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      (local.owner_label)   = local.kind_name
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {},
  )

  # ---- settings (settings.go) ---------------------------------------------------
  crds_install = try(var.spec.crds.install, null) == null ? true : var.spec.crds.install
  crds_keep    = try(var.spec.crds.keep_on_uninstall, null) == null ? true : var.spec.crds.keep_on_uninstall

  gke                    = try(var.spec.platform.gke, null) != null
  gke_cluster_id         = try(var.spec.platform.gke.cluster_id, "")
  api_server_gsa         = try(var.spec.platform.gke.api_server_service_account, "")
  node_agent_gsa         = try(var.spec.platform.gke.node_agent_service_account, "")
  transfer_gsa           = try(var.spec.platform.gke.snapshot_transfer_service_account, "")
  service_account_issuer = try(var.spec.platform.generic.service_account_issuer, "")

  preview  = try(var.spec.preview_gates, [])
  size10   = try(var.spec.cluster_size, "") == "size10"
  cordon   = try(var.spec.cordon_control_plane, false)
  registry = try(var.spec.image_registry, "")
  pull_secrets = [
    for name in try(var.spec.image_pull_secrets, []) : { name = name }
  ]

  db_kind = (
    try(var.spec.database.external, null) != null ? "external" :
    try(var.spec.database.cloud_sql, null) != null ? "cloudsql" : "bundled"
  )
  db_roles = try(var.spec.database.external.roles, null) != null ? var.spec.database.external.roles : try(var.spec.database.cloud_sql.roles, null)
  db = {
    pool_max_conns   = coalesce(try(var.spec.database.pool_max_connections, null), 0)
    storage_class    = try(var.spec.database.bundled.storage_class, "")
    volume_size      = coalesce(try(var.spec.database.bundled.volume_size, null), "500Gi")
    resources        = try(var.spec.database.bundled.resources, null)
    role_read_write  = coalesce(try(local.db_roles.read_write, null), "substrate_readwrite")
    role_owner       = coalesce(try(local.db_roles.owner, null), "substrate_owner")
    schema           = coalesce(try(local.db_roles.schema, null), "substrate")
    server_ca        = try(var.spec.database.external.server_ca_certificate, "")
    cloudsql         = try(var.spec.database.cloud_sql.instance_connection_name, "")
    cloudsql_iam     = try(var.spec.database.cloud_sql.iam_authentication, null) == null ? true : var.spec.database.cloud_sql.iam_authentication
    cloudsql_ip_type = trimsuffix(coalesce(try(var.spec.database.cloud_sql.ip_type, ""), "private_ip"), "_ip")
  }
  db_external_passwords = local.db_kind == "external" || (local.db_kind == "cloudsql" && !local.db.cloudsql_iam)
  db_read_write_secret  = try(var.spec.database.external.read_write_connection_string, var.spec.database.cloud_sql.read_write_connection_string, null)
  db_owner_secret = coalesce(
    try(var.spec.database.external.owner_connection_string, null),
    try(var.spec.database.cloud_sql.owner_connection_string, null),
    local.db_read_write_secret,
    { name = "", key = "" },
  )

  s3 = try(var.spec.storage.s3, null)
  storage = {
    is_s3      = local.s3 != null
    gcs_bucket = try(var.spec.storage.gcs.bucket, "")
    endpoint   = try(local.s3.endpoint, "")
    bucket     = try(local.s3.bucket, "")
    region     = coalesce(try(local.s3.region, null), "us-east-1")
    path_style = try(local.s3.force_path_style, null) == null ? true : local.s3.force_path_style
    secret     = try(local.s3.credentials.secret_name, "")
    id_key     = coalesce(try(local.s3.credentials.access_key_id_key, null), "access_key_id")
    secret_key = coalesce(try(local.s3.credentials.secret_access_key_key, null), "secret_access_key")
  }
  # Where actor snapshots are stored, the API server's snapshot location: the
  # bucket, under the prefix when one is set (storageSettings.location).
  snapshot_prefix  = try(var.spec.storage.snapshot_prefix, "")
  storage_location = "${local.storage.is_s3 ? "s3://${local.storage.bucket}" : "gs://${local.storage.gcs_bucket}"}${local.snapshot_prefix != "" ? "/${local.snapshot_prefix}" : ""}"

  api = {
    replicas         = coalesce(try(var.spec.api_server.replicas, null), 2)
    resources        = try(var.spec.api_server.resources, null)
    log_level        = coalesce(try(var.spec.api_server.log_level, ""), "info")
    bootstrap_owners = coalescelist(try(var.spec.api_server.authorization.bootstrap_owners, []), ["system:serviceaccount:ate-system:ate-client"])
    jwt_algorithm    = upper(coalesce(try(var.spec.api_server.actor_jwt.algorithm, ""), "es256"))
    actor_issuer     = coalesce(try(var.spec.api_server.actor_jwt.issuer, ""), "https://idp.${local.namespace}.svc")
    drain_delay      = coalesce(try(var.spec.api_server.drain_delay, null), "13s")
    drain_timeout    = coalesce(try(var.spec.api_server.drain_timeout, null), "15s")
    resync_interval  = coalesce(try(var.spec.api_server.template_resync_interval, null), "20s")
  }

  identity = {
    replicas    = coalesce(try(var.spec.identity_server.replicas, null), 2)
    resources   = try(var.spec.identity_server.resources, null)
    log_level   = coalesce(try(var.spec.identity_server.log_level, ""), "info")
    drain_grace = coalesce(try(var.spec.identity_server.drain_grace, null), "5s")
  }

  controller = {
    resources   = try(var.spec.controller.resources, null)
    log_level   = coalesce(try(var.spec.controller.log_level, ""), "info")
    worker_sync = coalesce(try(var.spec.controller.worker_sync_concurrency, null), 2)
  }

  node_agent = {
    resources     = try(var.spec.node_agent.resources, null)
    log_level     = coalesce(try(var.spec.node_agent.log_level, ""), "info")
    node_selector = try(var.spec.node_agent.node_selector, {})
    tolerations = [
      for t in try(var.spec.node_agent.tolerations, []) : merge(
        try(t.key, "") != "" ? { key = t.key } : {},
        try(t.operator, "") != "" ? { operator = t.operator } : {},
        try(t.value, "") != "" ? { value = t.value } : {},
        try(t.effect, "") != "" ? { effect = t.effect } : {},
        try(t.toleration_seconds, null) != null ? { tolerationSeconds = t.toleration_seconds } : {},
      )
    ]
    stats_poll    = coalesce(try(var.spec.node_agent.actor_stats_poll_interval, null), "1m")
    drain_delay   = coalesce(try(var.spec.node_agent.drain_delay, null), "0s")
    drain_timeout = coalesce(try(var.spec.node_agent.drain_timeout, null), "5m")
    gc_period     = coalesce(try(var.spec.node_agent.image_cache.gc_period, null), "5m")
    high_percent  = coalesce(try(var.spec.node_agent.image_cache.high_percent, null), 85)
    low_percent   = coalesce(try(var.spec.node_agent.image_cache.low_percent, null), 80)
    max_bytes     = coalesce(try(var.spec.node_agent.image_cache.max_bytes, null), 0)
    min_age       = coalesce(try(var.spec.node_agent.image_cache.min_age, null), "2m")
    gc_dry_run    = try(var.spec.node_agent.image_cache.dry_run, false)
    staging_gib   = coalesce(try(var.spec.node_agent.asset_staging_capacity_gib, null), 8)
  }
  node_snapshots = {
    upload_delay   = coalesce(try(var.spec.node_agent.snapshots.upload_delay, null), "5m")
    detach_delay   = coalesce(try(var.spec.node_agent.snapshots.detach_delay, null), "30m")
    upload_timeout = coalesce(try(var.spec.node_agent.snapshots.upload_timeout, null), "2m")
    upload_mib     = coalesce(try(var.spec.node_agent.snapshots.upload_throughput_mib, null), 100)
    upload_workers = coalesce(try(var.spec.node_agent.snapshots.upload_workers, null), 2)
    detach_workers = coalesce(try(var.spec.node_agent.snapshots.detach_workers, null), 2)
    report_workers = coalesce(try(var.spec.node_agent.snapshots.report_workers, null), 1)
    report_give_up = coalesce(try(var.spec.node_agent.snapshots.report_give_up_after, null), "10m")
    upload_memory  = coalesce(try(var.spec.node_agent.snapshots.upload_memory_budget, null), "512Mi")
  }
  # The node's two file caches (resolveFileCache): the same flags under each
  # cache's prefix, with the fetch timeout named apart.
  file_caches = {
    "sandbox-asset-cache" = {
      timeout_flag    = "sandbox-asset-fetch-timeout"
      disabled        = false
      min_age         = coalesce(try(var.spec.node_agent.sandbox_asset_cache.min_age, null), "1h")
      min_age_default = "1h"
      timeout         = coalesce(try(var.spec.node_agent.sandbox_asset_cache.fetch_timeout, null), "5m")
      timeout_default = "5m"
      gc_period       = coalesce(try(var.spec.node_agent.sandbox_asset_cache.eviction.gc_period, null), "5m")
      max_bytes       = coalesce(try(var.spec.node_agent.sandbox_asset_cache.eviction.max_bytes, null), 0)
      gc_dry_run      = try(var.spec.node_agent.sandbox_asset_cache.eviction.dry_run, false)
    }
    "snapshot-cache" = {
      timeout_flag    = "snapshot-cache-fetch-timeout"
      disabled        = try(var.spec.node_agent.snapshot_cache.disabled, false)
      min_age         = coalesce(try(var.spec.node_agent.snapshot_cache.min_age, null), "10m")
      min_age_default = "10m"
      timeout         = coalesce(try(var.spec.node_agent.snapshot_cache.fetch_timeout, null), "10m")
      timeout_default = "10m"
      gc_period       = coalesce(try(var.spec.node_agent.snapshot_cache.eviction.gc_period, null), "5m")
      max_bytes       = coalesce(try(var.spec.node_agent.snapshot_cache.eviction.max_bytes, null), 0)
      gc_dry_run      = try(var.spec.node_agent.snapshot_cache.eviction.dry_run, false)
    }
  }

  router = {
    agentgateway        = try(var.spec.router.dataplane, "") == "agentgateway"
    replicas            = coalesce(try(var.spec.router.replicas, null), 1)
    resources           = try(var.spec.router.resources, null)
    log_level           = coalesce(try(var.spec.router.log_level, ""), "info")
    spiffe_ids          = try(var.spec.router.allowed_client_spiffe_ids, [])
    route_timeout       = coalesce(try(var.spec.router.route_timeout, null), "5m")
    park_budget         = coalesce(try(var.spec.router.parked_requests.budget, null), "5s")
    park_max            = coalesce(try(var.spec.router.parked_requests.max, null), 1024)
    park_retry_interval = coalesce(try(var.spec.router.parked_requests.retry_interval, null), "100ms")
    park_retry_factor   = coalesce(try(var.spec.router.parked_requests.retry_factor, null), 1.1)
    park_retry_jitter   = coalesce(try(var.spec.router.parked_requests.retry_jitter, null), 0.1)
    extproc_max         = coalesce(try(var.spec.router.ext_proc_max_requests, null), 0)
    drain_delay         = coalesce(try(var.spec.router.drain_delay, null), "13s")
    drain_timeout       = try(var.spec.router.drain_timeout, "")
  }

  provider_kind = (
    try(var.spec.egress.credential_provider.kubernetes_secrets, null) != null ? "kubernetes" :
    try(var.spec.egress.credential_provider.external, null) != null ? "external" : "disabled"
  )
  ksp = try(var.spec.egress.credential_provider.kubernetes_secrets, null)
  egress = {
    provider = local.provider_kind
    provider_name = (
      local.provider_kind == "kubernetes" ? "k8s.io" :
      local.provider_kind == "external" ? var.spec.egress.credential_provider.external.name : ""
    )
    provider_address = (
      local.provider_kind == "kubernetes" ? "${local.credential_provider_name}.${local.namespace}.svc:50051" :
      local.provider_kind == "external" ? var.spec.egress.credential_provider.external.address : ""
    )
    grants             = try(local.ksp.grants, [])
    provider_resources = try(local.ksp.resources, null)
    provider_log_level = coalesce(try(local.ksp.log_level, ""), "info")
    provider_qps       = coalesce(try(local.ksp.kube_api_qps, null), 50)
    provider_burst     = coalesce(try(local.ksp.kube_api_burst, null), 100)
    provider_drain     = coalesce(try(local.ksp.drain_grace, null), "5s")
    replicas           = coalesce(try(var.spec.egress.replicas, null), 1)
    resources          = try(var.spec.egress.resources, null)
    log_level          = coalesce(try(var.spec.egress.log_level, ""), "info")
    policy_ttl         = coalesce(try(var.spec.egress.policy_cache_ttl, null), "10s")
    intercept_ttl      = coalesce(try(var.spec.egress.intercept_certificate_ttl, null), "15m")
    envoy_concurrency  = coalesce(try(var.spec.egress.envoy_concurrency, null), 0)
    extproc_namespace  = try(var.spec.egress.additional_ext_proc_service.namespace, "")
    extproc_service    = try(var.spec.egress.additional_ext_proc_service.service, "")
    extproc_port       = try(var.spec.egress.additional_ext_proc_service.port, 0)
    drain_delay        = coalesce(try(var.spec.egress.drain_delay, null), "13s")
  }
  # The name the gateway verifies a provider as: its address without the port.
  provider_server_name = length(regexall(":[^:]*$", local.egress.provider_address)) > 0 ? replace(local.egress.provider_address, "/:[^:]*$/", "") : local.egress.provider_address

  pod_certs = {
    resources          = try(var.spec.pod_certificates.resources, null)
    workers_per_signer = coalesce(try(var.spec.pod_certificates.workers_per_signer, null), 1)
    qps                = coalesce(try(var.spec.pod_certificates.kube_api_qps, null), 0)
    burst              = coalesce(try(var.spec.pod_certificates.kube_api_burst, null), 0)
  }

  telemetry = {
    endpoint        = try(var.spec.telemetry.otlp_endpoint, "")
    traces          = try(var.spec.telemetry.traces_exporters, [])
    metrics         = try(var.spec.telemetry.metrics_exporters, [])
    logs            = try(var.spec.telemetry.logs_exporters, [])
    sampler         = try(var.spec.telemetry.traces_sampler, "")
    sampler_ratio   = try(var.spec.telemetry.traces_sampler_ratio, null)
    export_interval = coalesce(try(var.spec.telemetry.metric_export_interval_ms, null), 0)
    export_timeout  = coalesce(try(var.spec.telemetry.metric_export_timeout_ms, null), 0)
    worker_traces   = try(var.spec.telemetry.worker_traces_exporters, [])
    worker_metrics  = try(var.spec.telemetry.worker_metrics_exporters, [])
    worker_logs     = try(var.spec.telemetry.worker_logs_exporters, [])
  }

  # ---- derived configuration (config.go) ----------------------------------------
  bundled_host = "postgres.ate-system.svc:5432"
  bundled_dsn = {
    for user in ["substrate_readwrite_user", "substrate_owner_user"] :
    user => "postgresql://${user}@${local.bundled_host}/atepg?sslmode=verify-full&sslrootcert=/run/servicedns.podcert.ate.dev/trust-bundle.pem&sslcert=/run/postgres.podcert.ate.dev/${user}.pem&sslkey=/run/postgres.podcert.ate.dev/${user}.pem"
  }
  cloudsql_dsn = "user=${trimsuffix(local.api_server_gsa, ".gserviceaccount.com")} host=127.0.0.1 port=5432 dbname=atepg sslmode=disable"

  api_env_config_map = merge(
    {
      "ATE_API_POSTGRES_READ_WRITE_ROLE"          = local.db.role_read_write
      "ATE_API_POSTGRES_OWNER_ROLE"               = local.db.role_owner
      "ATE_API_ACTOR_JWT_ISSUER"                  = local.api.actor_issuer
      "ATE_API_SNAPSHOT_LOCATION"                 = local.storage_location
      "ATE_API_SNAPSHOT_TRANSFER_SERVICE_ACCOUNT" = local.transfer_gsa
    },
    local.db.pool_max_conns > 0 ? { "ATE_API_POSTGRES_POOL_MAX_CONNS" = tostring(local.db.pool_max_conns) } : {},
    local.db_kind == "cloudsql" ? {
      "ATE_API_POSTGRES_CLOUDSQL_INSTANCE" = local.db.cloudsql
      "CSQL_PROXY_AUTO_IAM_AUTHN"          = tostring(local.db.cloudsql_iam)
      "CSQL_PROXY_PORT"                    = "5432"
      "CSQL_PROXY_HEALTH_CHECK"            = "true"
      "CSQL_PROXY_HTTP_ADDRESS"            = "0.0.0.0"
      "CSQL_PROXY_HTTP_PORT"               = "9801"
      "CSQL_PROXY_STRUCTURED_LOGS"         = "true"
    } : {},
    local.db_kind == "cloudsql" && local.db.cloudsql_ip_type == "private" ? { "CSQL_PROXY_PRIVATE_IP" = "true" } : {},
    local.db_kind == "cloudsql" && local.db.cloudsql_ip_type == "psc" ? { "CSQL_PROXY_PSC" = "true" } : {},
  )

  api_env_secret = merge(
    { "ATE_API_POSTGRES_SCHEMA" = local.db.schema },
    local.db_kind == "bundled" ? {
      "ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING" = "${local.bundled_dsn["substrate_readwrite_user"]}${local.size10 ? "&pool_max_conns=64&pool_min_conns=4" : ""}"
      "ATE_API_POSTGRES_OWNER_CONNECTION_STRING"      = local.bundled_dsn["substrate_owner_user"]
    } : {},
    local.db_kind == "cloudsql" && local.db.cloudsql_iam ? {
      "ATE_API_POSTGRES_READ_WRITE_CONNECTION_STRING" = local.cloudsql_dsn
      "ATE_API_POSTGRES_OWNER_CONNECTION_STRING"      = local.cloudsql_dsn
    } : {},
  )

  # The API server's configuration file (config.go apiConfigText): its log
  # level, read live, so it rolls nothing.
  api_config_text = "logging: {\n  level: \"${local.api.log_level}\"\n}\n"

  # authentication.yaml, byte for byte upstream's buildAuthenticationConfig.
  api_issuer = local.gke ? "https://container.googleapis.com/v1/${local.gke_cluster_id}" : local.service_account_issuer
  authentication_yaml = trimsuffix(join("", concat(
    ["jwtProviders:\n- name: kubernetes\n  issuer: ${local.api_issuer}\n  audiences: [${local.api_server_host}]\n"],
    contains(["https://kubernetes.default.svc", "https://kubernetes.default.svc.cluster.local"], local.api_issuer) ? [
      "  certificateAuthorityFile: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt\n  discoveryTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token\n",
    ] : [],
  )), "\n")

  # The credential provider's namespace policy (upstream's format).
  policy_yaml = length(local.egress.grants) == 0 ? "policies: []\n" : join("", concat(["policies:\n"], flatten([
    for g in local.egress.grants : concat(
      ["- atespace: ${jsonencode(g.atespace)}\n  allowedNamespaces:\n"],
      [for ns in g.allowed_namespaces : "  - ${jsonencode(ns)}\n"],
    )
  ])))

  # ate-otel-config: upstream's data with the spec's settings; on a generic
  # cluster with no endpoint nothing is pushed.
  otel_push_off = !local.gke && local.telemetry.endpoint == ""
  otel_traces   = local.otel_push_off && length(local.telemetry.traces) == 0 ? ["none"] : local.telemetry.traces
  otel_metrics  = local.otel_push_off && length(local.telemetry.metrics) == 0 ? ["prometheus"] : local.telemetry.metrics
  otel_upstream = try(local.documents_by_file.otel_config[0].data, {})
  otel_data = merge(
    { for k, v in local.otel_upstream : k => v if !(local.otel_push_off && k == "OTEL_EXPORTER_OTLP_ENDPOINT") },
    local.telemetry.endpoint != "" ? { "OTEL_EXPORTER_OTLP_ENDPOINT" = local.telemetry.endpoint } : {},
    length(local.otel_traces) > 0 ? { "OTEL_TRACES_EXPORTER" = join(",", local.otel_traces) } : {},
    length(local.otel_metrics) > 0 ? { "OTEL_METRICS_EXPORTER" = join(",", local.otel_metrics) } : {},
    length(local.telemetry.logs) > 0 ? { "OTEL_LOGS_EXPORTER" = join(",", local.telemetry.logs) } : {},
    length(local.telemetry.worker_traces) > 0 ? { "ATE_WORKER_OTEL_TRACES_EXPORTER" = join(",", local.telemetry.worker_traces) } : {},
    length(local.telemetry.worker_metrics) > 0 ? { "ATE_WORKER_OTEL_METRICS_EXPORTER" = join(",", local.telemetry.worker_metrics) } : {},
    length(local.telemetry.worker_logs) > 0 ? { "ATE_WORKER_OTEL_LOGS_EXPORTER" = join(",", local.telemetry.worker_logs) } : {},
    local.telemetry.sampler != "" ? { "OTEL_TRACES_SAMPLER" = local.telemetry.sampler } : {},
    local.telemetry.sampler_ratio != null ? { "OTEL_TRACES_SAMPLER_ARG" = tostring(local.telemetry.sampler_ratio) } : {},
    local.telemetry.export_interval > 0 ? { "OTEL_METRIC_EXPORT_INTERVAL" = tostring(local.telemetry.export_interval) } : {},
    local.telemetry.export_timeout > 0 ? { "OTEL_METRIC_EXPORT_TIMEOUT" = tostring(local.telemetry.export_timeout) } : {},
  )

  # Rollout hashes (config.go envHash and configHash).
  env_hash = sha256(join("", concat(
    ["configmap\n"], [for k in sort(keys(local.api_env_config_map)) : "${k}=${local.api_env_config_map[k]}\n"],
    ["secret\n"], [for k in sort(keys(local.api_env_secret)) : "${k}=${local.api_env_secret[k]}\n"],
  )))
  config_maps = merge(
    {
      "ate-api-authentication" = { "authentication.yaml" = local.authentication_yaml }
      "ate-otel-config"        = local.otel_data
    },
    local.egress.provider == "kubernetes" ? { "k8s-credential-provider-namespace-policy" = { "namespace-policy.yaml" = local.policy_yaml } } : {},
    local.db_kind == "bundled" ? {
      "postgres-config"                 = local.postgres_config_data
      (local.postgres_setup_config_map) = { "setup.sql" = local.release_bodies["setup_sql"] }
    } : {},
  )
  config_readers = {
    (local.api_server_name)          = ["ate-api-authentication", "ate-otel-config"]
    (local.controller_name)          = ["ate-otel-config"]
    (local.node_agent_name)          = ["ate-otel-config"]
    (local.router_name)              = ["ate-otel-config"]
    (local.egress_name)              = ["ate-otel-config"]
    (local.credential_provider_name) = ["k8s-credential-provider-namespace-policy"]
    (local.postgres_name)            = ["postgres-config", local.postgres_setup_config_map]
  }
  config_hashes = {
    for workload, names in local.config_readers : workload => sha256(join("", flatten([
      for name in names : concat(
        ["configmap ${name}\n"],
        [for k in sort(keys(try(local.config_maps[name], {}))) : "${k}=${local.config_maps[name][k]}\n"],
      )
    ])))
  }

  # ---- flags per component (workloads.go) ---------------------------------------
  # The bootstrap owners are always written, as upstream's installer writes
  # them; the log level is the configuration file's.
  api_flags = [for f in [
    { on = true, name = "authz-bootstrap-owners", value = join(",", local.api.bootstrap_owners) },
    { on = local.api.drain_delay != "13s", name = "drain-delay", value = local.api.drain_delay },
    { on = local.api.drain_timeout != "15s", name = "drain-timeout", value = local.api.drain_timeout },
    { on = local.api.resync_interval != "20s", name = "template-resync-interval", value = local.api.resync_interval },
  ] : f if f.on]
  identity_flags = [for f in [
    { on = true, name = "issuer", value = local.api.actor_issuer },
    { on = local.identity.log_level != "info", name = "log-level", value = local.identity.log_level },
    { on = local.identity.drain_grace != "5s", name = "drain-grace", value = local.identity.drain_grace },
  ] : f if f.on]
  controller_flags = [for f in [
    { on = local.controller.log_level != "info", name = "log-level", value = local.controller.log_level },
    { on = local.controller.worker_sync != 2, name = "worker-sync-concurrency", value = tostring(local.controller.worker_sync) },
  ] : f if f.on]
  node_agent_flags = [for f in concat([
    { on = local.node_agent.log_level != "info", name = "log-level", value = local.node_agent.log_level },
    { on = local.node_agent.stats_poll != "1m", name = "actor-stats-poll-interval", value = local.node_agent.stats_poll },
    { on = local.node_agent.drain_delay != "0s", name = "drain-delay", value = local.node_agent.drain_delay },
    { on = local.node_agent.drain_timeout != "5m", name = "drain-timeout", value = local.node_agent.drain_timeout },
    { on = local.node_agent.gc_period != "5m", name = "image-cache-gc-period", value = local.node_agent.gc_period },
    { on = local.node_agent.high_percent != 85, name = "image-cache-high-percent", value = tostring(local.node_agent.high_percent) },
    { on = local.node_agent.low_percent != 80, name = "image-cache-low-percent", value = tostring(local.node_agent.low_percent) },
    { on = local.node_agent.max_bytes != 0, name = "image-cache-max-bytes", value = tostring(local.node_agent.max_bytes) },
    { on = local.node_agent.min_age != "2m", name = "image-cache-min-age", value = local.node_agent.min_age },
    { on = local.node_agent.gc_dry_run, name = "image-cache-gc-dry-run", value = "true" },
    { on = local.node_snapshots.upload_delay != "5m", name = "snapshot-upload-delay", value = local.node_snapshots.upload_delay },
    { on = local.node_snapshots.detach_delay != "30m", name = "snapshot-detach-delay", value = local.node_snapshots.detach_delay },
    { on = local.node_snapshots.upload_timeout != "2m", name = "snapshot-upload-timeout", value = local.node_snapshots.upload_timeout },
    { on = local.node_snapshots.upload_mib != 100, name = "snapshot-upload-est-throughput-mib", value = tostring(local.node_snapshots.upload_mib) },
    { on = local.node_snapshots.upload_workers != 2, name = "snapshot-upload-workers", value = tostring(local.node_snapshots.upload_workers) },
    { on = local.node_snapshots.detach_workers != 2, name = "snapshot-detach-workers", value = tostring(local.node_snapshots.detach_workers) },
    { on = local.node_snapshots.report_workers != 1, name = "snapshot-report-workers", value = tostring(local.node_snapshots.report_workers) },
    { on = local.node_snapshots.report_give_up != "10m", name = "snapshot-report-give-up-after", value = local.node_snapshots.report_give_up },
    ], flatten([for prefix, c in local.file_caches : [
      { on = c.min_age != c.min_age_default, name = "${prefix}-min-age", value = c.min_age },
      { on = c.timeout != c.timeout_default, name = c.timeout_flag, value = c.timeout },
      { on = c.gc_period != "5m", name = "${prefix}-gc-period", value = c.gc_period },
      { on = c.max_bytes != 0, name = "${prefix}-max-bytes", value = tostring(c.max_bytes) },
      { on = c.gc_dry_run, name = "${prefix}-gc-dry-run", value = "true" },
    ]]), [
    # An empty cache directory turns the snapshot cache off.
    { on = local.file_caches["snapshot-cache"].disabled, name = "snapshot-cache-dir", value = "" },
  ]) : f if f.on]
  staging_flags = [for f in [
    { on = local.node_agent.staging_gib != 8, name = "asset-staging-capacity", value = "${local.node_agent.staging_gib}Gi" },
    { on = local.node_snapshots.upload_memory != "512Mi", name = "upload-memory-budget", value = local.node_snapshots.upload_memory },
  ] : f if f.on]
  router_flags = [for f in [
    { on = local.router.log_level != "info", name = "log-level", value = local.router.log_level },
    { on = length(local.router.spiffe_ids) > 0, name = "ingress-allowed-spiffe-ids", value = join(",", local.router.spiffe_ids) },
    { on = local.router.route_timeout != "5m", name = "route-timeout", value = local.router.route_timeout },
    { on = local.router.park_budget != "5s", name = "parked-request-budget", value = local.router.park_budget },
    { on = local.router.park_max != 1024, name = "parked-request-max", value = tostring(local.router.park_max) },
    { on = local.router.park_retry_interval != "100ms", name = "parked-request-retry-interval", value = local.router.park_retry_interval },
    { on = local.router.park_retry_factor != 1.1, name = "parked-request-retry-factor", value = tostring(local.router.park_retry_factor) },
    { on = local.router.park_retry_jitter != 0.1, name = "parked-request-retry-jitter", value = tostring(local.router.park_retry_jitter) },
    { on = local.router.extproc_max != 0, name = "extproc-max-requests", value = tostring(local.router.extproc_max) },
    { on = local.router.drain_delay != "13s", name = "drain-delay", value = local.router.drain_delay },
    { on = local.router.drain_timeout != "", name = "drain-timeout", value = local.router.drain_timeout },
  ] : f if f.on]
  egress_flags = [for f in [
    { on = local.egress.log_level != "info", name = "log-level", value = local.egress.log_level },
    { on = local.egress.policy_ttl != "10s", name = "egress-policy-cache-ttl", value = local.egress.policy_ttl },
    { on = local.egress.drain_delay != "13s", name = "drain-delay", value = local.egress.drain_delay },
    { on = local.egress.provider != "disabled", name = "credential-provider-name", value = local.egress.provider_name },
    { on = local.egress.provider != "disabled", name = "credential-provider-address", value = local.egress.provider_address },
    { on = local.egress.provider != "disabled", name = "credential-provider-ca-file", value = "/run/servicedns.podcert.ate.dev/trust-bundle.pem" },
    { on = local.egress.provider != "disabled", name = "credential-provider-client-cert", value = "/run/podidentity.podcert.ate.dev/credential-bundle.pem" },
    { on = local.egress.provider != "disabled", name = "credential-provider-server-name", value = local.provider_server_name },
  ] : f if f.on]
  minter_flags = [for f in [
    { on = local.egress.intercept_ttl != "15m", name = "leaf-cert-ttl", value = local.egress.intercept_ttl },
    { on = local.egress.log_level != "info", name = "log-level", value = local.egress.log_level },
  ] : f if f.on]
  provider_flags = [for f in [
    { on = local.egress.provider_log_level != "info", name = "log-level", value = local.egress.provider_log_level },
    { on = local.egress.provider_qps != 50, name = "kube-api-qps", value = tostring(local.egress.provider_qps) },
    { on = local.egress.provider_burst != 100, name = "kube-api-burst", value = tostring(local.egress.provider_burst) },
    { on = local.egress.provider_drain != "5s", name = "drain-grace", value = local.egress.provider_drain },
  ] : f if f.on]
  pod_cert_flags = [for f in [
    { on = local.pod_certs.qps > 0, name = "kube-api-qps", value = tostring(local.pod_certs.qps) },
    { on = local.pod_certs.burst > 0, name = "kube-api-burst", value = tostring(local.pod_certs.burst) },
  ] : f if f.on]

  # Storage environment on both objectstore-plugin sidecars (storageEnv).
  storage_env = concat(
    [{ name = "ATE_STORAGE_BACKEND", value = local.storage.is_s3 ? "s3" : "gcs" }],
    [for e in [
      { name = "AWS_REGION", value = local.storage.region },
      { name = "AWS_ENDPOINT_URL", value = local.storage.endpoint },
    ] : e if local.storage.is_s3],
    [for e in [{ name = "AWS_S3_USE_PATH_STYLE", value = "true" }] : e if local.storage.is_s3 && local.storage.path_style],
    [for e in [
      { name = "AWS_ACCESS_KEY_ID", valueFrom = { secretKeyRef = { name = local.storage.secret, key = local.storage.id_key } } },
      { name = "AWS_SECRET_ACCESS_KEY", valueFrom = { secretKeyRef = { name = local.storage.secret, key = local.storage.secret_key } } },
    ] : e if local.storage.is_s3],
  )

  # The database start hook's setup invocation (workloads.go setupCommand).
  setup_command = "psql --no-psqlrc --set=ON_ERROR_STOP=1 --single-transaction --username \"$${POSTGRES_USER}\" --dbname atepg --set=substrate_schema=substrate --set=substrate_owner_role=substrate_owner --set=substrate_owner_user=substrate_owner_user --set=substrate_owner_password= --set=substrate_readwrite_role=substrate_readwrite --set=substrate_readwrite_user=substrate_readwrite_user --set=substrate_readwrite_password= --file=/etc/postgres-setup/setup.sql"
}
