locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels following OpenMCF conventions
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_name" = var.metadata.name
    "resource_kind" = "kubernetes_temporal"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge all labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Namespace from spec
  namespace = var.spec.namespace

  # Service names
  frontend_service_name = "${var.metadata.name}-frontend"
  ui_service_name       = "${var.metadata.name}-web"

  # Service ports
  frontend_grpc_port = 7233
  frontend_http_port = 7243
  ui_port            = 8080

  # Internal cluster endpoints
  frontend_endpoint = "${local.frontend_service_name}.${local.namespace}.svc.cluster.local:${local.frontend_grpc_port}"
  web_ui_endpoint   = "${local.ui_service_name}.${local.namespace}.svc.cluster.local:${local.ui_port}"

  # Port-forward commands for local access
  port_forward_frontend_command = "kubectl port-forward -n ${local.namespace} service/${local.frontend_service_name} 7233:7233"
  port_forward_ui_command       = "kubectl port-forward -n ${local.namespace} service/${local.ui_service_name} 8080:8080"

  # Ingress configuration
  frontend_ingress_enabled = try(var.spec.ingress.frontend.enabled, false)
  frontend_grpc_hostname   = try(var.spec.ingress.frontend.grpc_hostname, "")
  frontend_http_hostname   = try(var.spec.ingress.frontend.http_hostname, "")
  web_ui_ingress_enabled   = try(var.spec.ingress.web_ui.enabled, false)
  web_ui_hostname          = try(var.spec.ingress.web_ui.hostname, "")

  # External hostname outputs (only set if ingress is enabled)
  external_frontend_hostname = local.frontend_ingress_enabled ? local.frontend_grpc_hostname : ""
  external_ui_hostname       = local.web_ui_ingress_enabled ? local.web_ui_hostname : ""

  # Database configuration
  database_backend          = var.spec.database.backend
  has_external_database     = var.spec.database.external_database != null
  database_name             = var.spec.database.database_name
  visibility_name           = var.spec.database.visibility_name
  disable_auto_schema_setup = var.spec.database.disable_auto_schema_setup

  # Database backend booleans
  is_cassandra  = local.database_backend == "cassandra"
  is_postgresql = local.database_backend == "postgresql"
  is_mysql      = local.database_backend == "mysql"

  # SQL sub-driver mapping
  sql_driver = local.is_postgresql ? "postgres12" : (local.is_mysql ? "mysql8" : "")

  # External database details (when provided)
  external_db_host     = try(var.spec.database.external_database.host, "")
  external_db_port     = try(var.spec.database.external_database.port, 0)
  external_db_username = try(var.spec.database.external_database.username, "")

  # Password handling - check if using secret_ref or string_value
  external_db_password_secret_ref = try(var.spec.database.external_database.password.secret_ref, null)
  external_db_password_string     = try(var.spec.database.external_database.password.string_value, "")

  # Determine which secret to use for database password
  # If secret_ref is provided, use the existing secret; otherwise, use the secret we create
  use_existing_db_secret = local.external_db_password_secret_ref != null
  db_secret_name         = local.use_existing_db_secret ? local.external_db_password_secret_ref.name : local.database_secret_name
  db_secret_key          = local.use_existing_db_secret ? local.external_db_password_secret_ref.key : local.database_secret_key

  # Monitoring stack configuration
  has_external_elasticsearch = var.spec.external_elasticsearch != null
  enable_monitoring_stack    = var.spec.enable_monitoring_stack || local.has_external_elasticsearch

  # External Elasticsearch details
  external_es_host = try(var.spec.external_elasticsearch.host, "")
  external_es_port = try(var.spec.external_elasticsearch.port, 0)
  external_es_user = try(var.spec.external_elasticsearch.user, "")

  # Elasticsearch password handling - check if using secret_ref or string_value
  external_es_password_secret_ref = try(var.spec.external_elasticsearch.password.secret_ref, null)
  external_es_password_string     = try(var.spec.external_elasticsearch.password.string_value, "")
  use_existing_es_secret          = local.external_es_password_secret_ref != null

  # Helm chart configuration
  helm_chart_name       = "temporal"
  helm_chart_repository = "https://go.temporal.io/helm-charts"
  helm_chart_version    = var.spec.version

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  database_secret_name = "${var.metadata.name}-db-password"
  database_secret_key  = "password"

  # Gateway API resource names derived from hostname (flattened: dots replaced with dashes)
  # Since these resources are created in istio-ingress namespace, using full domain names
  # makes them easily identifiable (e.g., "temporal-app-prod-main-ui-planton-live-external")

  # Frontend HTTP ingress resources
  frontend_http_flattened = replace(local.frontend_http_hostname, ".", "-")
  frontend_http_cert_secret_name = (
    local.frontend_http_hostname != ""
    ? local.frontend_http_flattened
    : "${var.metadata.name}-frontend-http-cert"
  )
  frontend_http_gateway_name = (
    local.frontend_http_hostname != ""
    ? "${local.frontend_http_flattened}-external"
    : "${var.metadata.name}-frontend-http-external"
  )
  frontend_http_redirect_route_name = (
    local.frontend_http_hostname != ""
    ? "${local.frontend_http_flattened}-redirect"
    : "${var.metadata.name}-frontend-http-redirect"
  )
  frontend_https_route_name = (
    local.frontend_http_hostname != ""
    ? "${local.frontend_http_flattened}-https"
    : "${var.metadata.name}-frontend-https"
  )

  # Web UI ingress resources
  ui_flattened = replace(local.web_ui_hostname, ".", "-")
  ui_cert_secret_name = (
    local.web_ui_hostname != ""
    ? local.ui_flattened
    : "${var.metadata.name}-ui-cert"
  )
  ui_gateway_name = (
    local.web_ui_hostname != ""
    ? "${local.ui_flattened}-external"
    : "${var.metadata.name}-ui-external"
  )
  ui_http_redirect_route_name = (
    local.web_ui_hostname != ""
    ? "${local.ui_flattened}-redirect"
    : "${var.metadata.name}-ui-http-redirect"
  )
  ui_https_route_name = (
    local.web_ui_hostname != ""
    ? "${local.ui_flattened}-https"
    : "${var.metadata.name}-ui-https"
  )

  # Dynamic configuration
  has_dynamic_config        = var.spec.dynamic_config != null
  history_size_limit_error  = try(var.spec.dynamic_config.history_size_limit_error, null)
  history_count_limit_error = try(var.spec.dynamic_config.history_count_limit_error, null)
  history_size_limit_warn   = try(var.spec.dynamic_config.history_size_limit_warn, null)
  history_count_limit_warn  = try(var.spec.dynamic_config.history_count_limit_warn, null)

  # History shards configuration
  num_history_shards = var.spec.num_history_shards

  # Service configuration - Frontend
  has_frontend_config         = try(var.spec.services.frontend, null) != null
  frontend_replicas           = try(var.spec.services.frontend.replicas, null)
  frontend_resources_limits   = try(var.spec.services.frontend.resources.limits, null)
  frontend_resources_requests = try(var.spec.services.frontend.resources.requests, null)

  # Service configuration - History
  has_history_config         = try(var.spec.services.history, null) != null
  history_replicas           = try(var.spec.services.history.replicas, null)
  history_resources_limits   = try(var.spec.services.history.resources.limits, null)
  history_resources_requests = try(var.spec.services.history.resources.requests, null)

  # Service configuration - Matching
  has_matching_config         = try(var.spec.services.matching, null) != null
  matching_replicas           = try(var.spec.services.matching.replicas, null)
  matching_resources_limits   = try(var.spec.services.matching.resources.limits, null)
  matching_resources_requests = try(var.spec.services.matching.resources.requests, null)

  # Service configuration - Worker
  has_worker_config         = try(var.spec.services.worker, null) != null
  worker_replicas           = try(var.spec.services.worker.replicas, null)
  worker_resources_limits   = try(var.spec.services.worker.resources.limits, null)
  worker_resources_requests = try(var.spec.services.worker.resources.requests, null)

  # ---------------------------------------------------------------- Helm values
  # The following locals assemble the Temporal Helm chart values map. They mirror the
  # prior `set {}`/`dynamic "set"` keys exactly (helm provider v3 dropped those blocks),
  # rendered through the house values=[yamlencode(...)] idiom. Conditional fragments are
  # appended as single-element lists so conditional branches share a list type (bare-object
  # branches fail tofu's type unification). This mirrors the Pulumi module's values map.

  # Datastore enable flags (+ embedded cassandra config). Exactly one branch is active.
  temporal_db_values = merge(concat(
    local.has_external_database ? [{
      cassandra  = { enabled = false }
      mysql      = { enabled = false }
      postgresql = { enabled = false }
    }] : [],
    (!local.has_external_database && local.is_cassandra) ? [{
      cassandra = {
        enabled      = true
        replicaCount = var.spec.cassandra_replicas
        config = {
          dev          = true
          cluster_size = var.spec.cassandra_replicas
        }
      }
      mysql      = { enabled = false }
      postgresql = { enabled = false }
    }] : [],
    (!local.has_external_database && local.is_mysql) ? [{
      cassandra  = { enabled = false }
      mysql      = { enabled = true }
      postgresql = { enabled = false }
    }] : [],
    (!local.has_external_database && local.is_postgresql) ? [{
      cassandra  = { enabled = false }
      mysql      = { enabled = false }
      postgresql = { enabled = true }
    }] : [],
  )...)

  # server.config: frontend RPC ports always; external-DB SQL persistence when configured.
  temporal_server_config = merge(concat(
    [{
      services = {
        frontend = {
          rpc = {
            grpcPort = local.frontend_grpc_port
            httpPort = local.frontend_http_port
          }
        }
      }
    }],
    local.has_external_database ? [{
      persistence = {
        driver = "sql"
        default = {
          driver = "sql"
          sql = {
            driver            = local.sql_driver
            host              = local.external_db_host
            port              = local.external_db_port
            database          = local.database_name
            user              = local.external_db_username
            existingSecret    = local.db_secret_name
            existingSecretKey = local.db_secret_key
            tls = {
              enabled                = true
              enableHostVerification = false
            }
          }
        }
        visibility = {
          driver = "sql"
          sql = {
            driver            = local.sql_driver
            host              = local.external_db_host
            port              = local.external_db_port
            database          = local.visibility_name
            user              = local.external_db_username
            existingSecret    = local.db_secret_name
            existingSecretKey = local.db_secret_key
            tls = {
              enabled                = true
              enableHostVerification = false
            }
          }
        }
      }
    }] : [],
  )...)

  # server.dynamicConfig: each history limit override is a single-element list of {value}.
  temporal_dynamic_config = merge(concat(
    local.history_size_limit_error != null ? [{ "limit.historySize.error" = [{ value = local.history_size_limit_error }] }] : [],
    local.history_count_limit_error != null ? [{ "limit.historyCount.error" = [{ value = local.history_count_limit_error }] }] : [],
    local.history_size_limit_warn != null ? [{ "limit.historySize.warn" = [{ value = local.history_size_limit_warn }] }] : [],
    local.history_count_limit_warn != null ? [{ "limit.historyCount.warn" = [{ value = local.history_count_limit_warn }] }] : [],
  )...)

  # Per-service (frontend/history/matching/worker) replicaCount + resources, each leaf optional.
  temporal_frontend_values = merge(concat(
    local.frontend_replicas != null ? [{ replicaCount = local.frontend_replicas }] : [],
    (try(local.frontend_resources_limits.cpu, null) != null || try(local.frontend_resources_limits.memory, null) != null || try(local.frontend_resources_requests.cpu, null) != null || try(local.frontend_resources_requests.memory, null) != null) ? [{
      resources = merge(concat(
        (try(local.frontend_resources_limits.cpu, null) != null || try(local.frontend_resources_limits.memory, null) != null) ? [{
          limits = merge(concat(
            try(local.frontend_resources_limits.cpu, null) != null ? [{ cpu = local.frontend_resources_limits.cpu }] : [],
            try(local.frontend_resources_limits.memory, null) != null ? [{ memory = local.frontend_resources_limits.memory }] : [],
          )...)
        }] : [],
        (try(local.frontend_resources_requests.cpu, null) != null || try(local.frontend_resources_requests.memory, null) != null) ? [{
          requests = merge(concat(
            try(local.frontend_resources_requests.cpu, null) != null ? [{ cpu = local.frontend_resources_requests.cpu }] : [],
            try(local.frontend_resources_requests.memory, null) != null ? [{ memory = local.frontend_resources_requests.memory }] : [],
          )...)
        }] : [],
      )...)
    }] : [],
  )...)

  temporal_history_values = merge(concat(
    local.history_replicas != null ? [{ replicaCount = local.history_replicas }] : [],
    (try(local.history_resources_limits.cpu, null) != null || try(local.history_resources_limits.memory, null) != null || try(local.history_resources_requests.cpu, null) != null || try(local.history_resources_requests.memory, null) != null) ? [{
      resources = merge(concat(
        (try(local.history_resources_limits.cpu, null) != null || try(local.history_resources_limits.memory, null) != null) ? [{
          limits = merge(concat(
            try(local.history_resources_limits.cpu, null) != null ? [{ cpu = local.history_resources_limits.cpu }] : [],
            try(local.history_resources_limits.memory, null) != null ? [{ memory = local.history_resources_limits.memory }] : [],
          )...)
        }] : [],
        (try(local.history_resources_requests.cpu, null) != null || try(local.history_resources_requests.memory, null) != null) ? [{
          requests = merge(concat(
            try(local.history_resources_requests.cpu, null) != null ? [{ cpu = local.history_resources_requests.cpu }] : [],
            try(local.history_resources_requests.memory, null) != null ? [{ memory = local.history_resources_requests.memory }] : [],
          )...)
        }] : [],
      )...)
    }] : [],
  )...)

  temporal_matching_values = merge(concat(
    local.matching_replicas != null ? [{ replicaCount = local.matching_replicas }] : [],
    (try(local.matching_resources_limits.cpu, null) != null || try(local.matching_resources_limits.memory, null) != null || try(local.matching_resources_requests.cpu, null) != null || try(local.matching_resources_requests.memory, null) != null) ? [{
      resources = merge(concat(
        (try(local.matching_resources_limits.cpu, null) != null || try(local.matching_resources_limits.memory, null) != null) ? [{
          limits = merge(concat(
            try(local.matching_resources_limits.cpu, null) != null ? [{ cpu = local.matching_resources_limits.cpu }] : [],
            try(local.matching_resources_limits.memory, null) != null ? [{ memory = local.matching_resources_limits.memory }] : [],
          )...)
        }] : [],
        (try(local.matching_resources_requests.cpu, null) != null || try(local.matching_resources_requests.memory, null) != null) ? [{
          requests = merge(concat(
            try(local.matching_resources_requests.cpu, null) != null ? [{ cpu = local.matching_resources_requests.cpu }] : [],
            try(local.matching_resources_requests.memory, null) != null ? [{ memory = local.matching_resources_requests.memory }] : [],
          )...)
        }] : [],
      )...)
    }] : [],
  )...)

  temporal_worker_values = merge(concat(
    local.worker_replicas != null ? [{ replicaCount = local.worker_replicas }] : [],
    (try(local.worker_resources_limits.cpu, null) != null || try(local.worker_resources_limits.memory, null) != null || try(local.worker_resources_requests.cpu, null) != null || try(local.worker_resources_requests.memory, null) != null) ? [{
      resources = merge(concat(
        (try(local.worker_resources_limits.cpu, null) != null || try(local.worker_resources_limits.memory, null) != null) ? [{
          limits = merge(concat(
            try(local.worker_resources_limits.cpu, null) != null ? [{ cpu = local.worker_resources_limits.cpu }] : [],
            try(local.worker_resources_limits.memory, null) != null ? [{ memory = local.worker_resources_limits.memory }] : [],
          )...)
        }] : [],
        (try(local.worker_resources_requests.cpu, null) != null || try(local.worker_resources_requests.memory, null) != null) ? [{
          requests = merge(concat(
            try(local.worker_resources_requests.cpu, null) != null ? [{ cpu = local.worker_resources_requests.cpu }] : [],
            try(local.worker_resources_requests.memory, null) != null ? [{ memory = local.worker_resources_requests.memory }] : [],
          )...)
        }] : [],
      )...)
    }] : [],
  )...)

  # Merge per-service configs into the server map (only services with config).
  temporal_server_services = merge(concat(
    length(local.temporal_frontend_values) > 0 ? [{ frontend = local.temporal_frontend_values }] : [],
    length(local.temporal_history_values) > 0 ? [{ history = local.temporal_history_values }] : [],
    length(local.temporal_matching_values) > 0 ? [{ matching = local.temporal_matching_values }] : [],
    length(local.temporal_worker_values) > 0 ? [{ worker = local.temporal_worker_values }] : [],
  )...)

  # The full server map: config (always), numHistoryShards/dynamicConfig (optional),
  # and per-service overrides.
  temporal_server = merge(concat(
    [{ config = local.temporal_server_config }],
    local.num_history_shards != null ? [{ numHistoryShards = local.num_history_shards }] : [],
    length(local.temporal_dynamic_config) > 0 ? [{ dynamicConfig = local.temporal_dynamic_config }] : [],
    [local.temporal_server_services],
  )...)

  # External Elasticsearch values (only meaningful when external ES is configured).
  temporal_es_external = merge(concat(
    [{
      enabled = false
      host    = local.external_es_host
      port    = local.external_es_port
      scheme  = "http"
    }],
    local.external_es_user != "" ? [{ username = local.external_es_user }] : [],
    (!local.use_existing_es_secret && local.external_es_password_string != "") ? [{ password = local.external_es_password_string }] : [],
    local.use_existing_es_secret ? [{
      existingSecret    = local.external_es_password_secret_ref.name
      existingSecretKey = local.external_es_password_secret_ref.key
    }] : [],
  )...)

  # Top-level Helm values.
  helm_values = merge(concat(
    [{
      fullnameOverride = var.metadata.name
      schema = {
        createDatabase = {
          enabled = !local.disable_auto_schema_setup
        }
        setup  = { enabled = true }
        update = { enabled = true }
      }
      prometheus          = { enabled = local.enable_monitoring_stack }
      grafana             = { enabled = local.enable_monitoring_stack }
      kubePrometheusStack = { enabled = local.enable_monitoring_stack }
      server              = local.temporal_server
    }],
    [local.temporal_db_values],
    var.spec.disable_web_ui ? [{ web = { enabled = false } }] : [],
    local.has_external_elasticsearch ? [{ elasticsearch = local.temporal_es_external }] : [],
    (!local.has_external_elasticsearch && !var.spec.enable_embedded_elasticsearch) ? [{ elasticsearch = { enabled = false } }] : [],
  )...)
}

