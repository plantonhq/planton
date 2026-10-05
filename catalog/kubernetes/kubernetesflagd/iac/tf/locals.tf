# Computed values for the KubernetesFlagd module. Every resolution here has an
# exact twin in the Pulumi module - keep them in lockstep: the same arguments,
# the same sources document byte for byte (and so the same checksum), the same
# mounts, the same objects.

locals {
  name      = var.metadata.name
  namespace = var.spec.namespace

  labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      "planton.ai/kind"     = "KubernetesFlagd"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )
  selector_labels = {
    "app.kubernetes.io/name"     = "flagd"
    "app.kubernetes.io/instance" = var.metadata.name
  }
  pod_labels = merge(try(var.spec.pod_labels, {}), local.labels, local.selector_labels)

  image_repository = try(var.spec.image.repository, "") != "" && try(var.spec.image.repository, null) != null ? var.spec.image.repository : "ghcr.io/open-feature/flagd"
  image_tag        = try(var.spec.image.tag, "") != "" && try(var.spec.image.tag, null) != null ? var.spec.image.tag : "v0.17.0"
  image            = "${local.image_repository}:${local.image_tag}${try(var.spec.image.fips, false) == true ? "-fips" : ""}"
  pull_policy      = try(var.spec.image.pull_policy, "") != "" && try(var.spec.image.pull_policy, null) != null ? var.spec.image.pull_policy : "IfNotPresent"

  create_service_account = try(var.spec.service_account.existing_name, "") == ""
  service_account_name   = local.create_service_account ? var.metadata.name : var.spec.service_account.existing_name

  port            = try(var.spec.server.port, null) != null ? var.spec.server.port : 8013
  management_port = try(var.spec.server.management_port, null) != null ? var.spec.server.management_port : 8014
  sync_port       = try(var.spec.server.sync_port, null) != null ? var.spec.server.sync_port : 8015
  ofrep_port      = try(var.spec.server.ofrep_port, null) != null ? var.spec.server.ofrep_port : 8016
  service_type    = try(var.spec.server.service_type, "") != "" && try(var.spec.server.service_type, null) != null ? var.spec.server.service_type : "ClusterIP"

  tls_secret_name        = try(var.spec.server.tls_secret_name, "")
  otel_ca                = try(var.spec.telemetry.otel_ca_cert_secret, null)
  otel_client_tls_secret = try(var.spec.telemetry.otel_client_tls_secret_name, "")
  context_values         = try(var.spec.evaluation.context_values, {})
  context_from_header    = try(var.spec.evaluation.context_from_header, {})
  cors_origins           = try(var.spec.evaluation.cors_origins, [])

  # ---- flagd arguments (Pulumi twin: buildFlagdConfig) --------------------------
  args = concat(
    [
      "start",
      "--port=${local.port}",
      "--management-port=${local.management_port}",
      "--sync-port=${local.sync_port}",
      "--ofrep-port=${local.ofrep_port}",
      "--log-format=${try(var.spec.log.format, "") != "" && try(var.spec.log.format, null) != null ? var.spec.log.format : "json"}",
    ],
    try(var.spec.log.debug, false) == true ? ["--debug"] : [],
    local.tls_secret_name != "" ? ["--server-cert-path=/etc/flagd/tls/tls.crt", "--server-key-path=/etc/flagd/tls/tls.key"] : [],
    [for k in sort(keys(local.context_values)) : "--context-value=${k}=${local.context_values[k]}"],
    [for k in sort(keys(local.context_from_header)) : "--context-from-header=${k}=${local.context_from_header[k]}"],
    length(local.cors_origins) > 0 ? ["--cors-origin=${join(",", local.cors_origins)}"] : [],
    try(var.spec.ofrep_sse.enabled, null) != null ? ["--ofrep-sse-enabled=${var.spec.ofrep_sse.enabled}"] : [],
    try(var.spec.ofrep_sse.inactivity_delay_seconds, null) != null ? ["--ofrep-sse-inactivity-delay=${var.spec.ofrep_sse.inactivity_delay_seconds}"] : [],
    try(var.spec.ofrep_sse.public_url, "") != "" ? ["--ofrep-sse-public-url=${var.spec.ofrep_sse.public_url}"] : [],
    try(var.spec.sync.http_enabled, null) != null ? ["--sync-http-enabled=${var.spec.sync.http_enabled}"] : [],
    try(var.spec.sync.disable_metadata, false) == true ? ["--disable-sync-metadata"] : [],
    try(var.spec.sync.stream_deadline, "") != "" ? ["--stream-deadline=${var.spec.sync.stream_deadline}"] : [],
    try(var.spec.telemetry.metrics_exporter, "") == "otel" ? ["--metrics-exporter=otel"] : [],
    try(var.spec.telemetry.otel_collector_uri, "") != "" ? ["--otel-collector-uri=${var.spec.telemetry.otel_collector_uri}"] : [],
    local.otel_ca != null ? ["--otel-ca-path=/etc/flagd/otel-ca/${local.otel_ca.key}"] : [],
    local.otel_client_tls_secret != "" ? ["--otel-cert-path=/etc/flagd/otel-tls/tls.crt", "--otel-key-path=/etc/flagd/otel-tls/tls.key"] : [],
    try(var.spec.telemetry.otel_reload_interval, "") != "" ? ["--otel-reload-interval=${var.spec.telemetry.otel_reload_interval}"] : [],
    try(var.spec.limits.max_request_body_bytes, null) != null ? ["--max-request-body=${var.spec.limits.max_request_body_bytes}"] : [],
    try(var.spec.limits.max_request_header_bytes, null) != null ? ["--max-request-header=${var.spec.limits.max_request_header_bytes}"] : [],
  )

  # ---- sources (Pulumi twin: buildFlagdConfig) ----------------------------------
  sources = [for i, s in var.spec.sources : { for k, v in {
    uri = (
      s.config_map != null ? "/etc/flagd/sources/${i}/${s.config_map.key}" :
      s.http != null ? s.http.url :
      s.grpc != null ? s.grpc.target :
      s.feature_flag != null ? "${try(s.feature_flag.namespace, "") != "" ? s.feature_flag.namespace : local.namespace}/${s.feature_flag.name}" :
      s.google_storage != null ? "gs://${s.google_storage.bucket}/${s.google_storage.object}" :
      s.azure_blob != null ? "azblob://${s.azure_blob.bucket}/${s.azure_blob.object}" :
      s.s3 != null ? "s3://${s.s3.bucket}/${s.s3.object}" : null
    )
    provider = (
      s.config_map != null ? (try(s.config_map.watcher, "") != "" ? s.config_map.watcher : "file") :
      s.http != null ? "http" :
      s.grpc != null ? "grpc" :
      s.feature_flag != null ? "kubernetes" :
      s.google_storage != null ? "gcs" :
      s.azure_blob != null ? "azblob" :
      s.s3 != null ? "s3" : null
    )
    authHeader = try(s.http.auth_header, "") != "" ? s.http.auth_header : null
    headers    = length(merge(try(s.http.headers, {}), try(s.http.sensitive_headers, {}), try(s.grpc.headers, {}), try(s.grpc.sensitive_headers, {}))) > 0 ? merge(try(s.http.headers, {}), try(s.http.sensitive_headers, {}), try(s.grpc.headers, {}), try(s.grpc.sensitive_headers, {})) : null
    interval   = try(s.http.interval_seconds, s.google_storage.interval_seconds, s.azure_blob.interval_seconds, s.s3.interval_seconds, null)
    timeoutS   = try(s.http.timeout_seconds, null)
    oauth = try(s.http.oauth, null) == null ? null : {
      clientID     = s.http.oauth.client_id
      clientSecret = s.http.oauth.client_secret
      tokenUrl     = s.http.oauth.token_url
    }
    intervalSeed       = try(s.http.interval_seed, s.google_storage.interval_seed, s.azure_blob.interval_seed, s.s3.interval_seed, "") != "" ? try(s.http.interval_seed, s.google_storage.interval_seed, s.azure_blob.interval_seed, s.s3.interval_seed, null) : null
    tls                = try(s.grpc.tls, false) == true ? true : null
    certPath           = try(s.grpc.ca_cert_secret, null) != null ? "/etc/flagd/grpc-ca/${i}/${s.grpc.ca_cert_secret.key}" : null
    providerID         = try(s.grpc.provider_id, "") != "" ? s.grpc.provider_id : null
    maxMsgSize         = try(s.grpc.max_msg_size, null)
    incrementalUpdates = try(s.grpc.incremental_updates, false) == true ? true : null
    selector           = try(s.grpc.selector, "") != "" ? s.grpc.selector : null
  } : k => v if v != null }]

  sources_json     = jsonencode(local.sources)
  sources_checksum = sha256(local.sources_json)

  config_map_mounts = [for i, s in var.spec.sources : { volume = "source-${i}", config_map = s.config_map.config_map_name, mount_path = "/etc/flagd/sources/${i}" } if s.config_map != null]
  grpc_ca_mounts    = [for i, s in var.spec.sources : { volume = "grpc-ca-${i}", secret = s.grpc.ca_cert_secret.name, key = s.grpc.ca_cert_secret.key, mount_path = "/etc/flagd/grpc-ca/${i}" } if try(s.grpc.ca_cert_secret, null) != null]

  feature_flag_namespaces = toset([for s in var.spec.sources : (try(s.feature_flag.namespace, "") != "" ? s.feature_flag.namespace : local.namespace) if s.feature_flag != null])

  # A header declared plain and sensitive (Pulumi twin: mergeHeaders).
  header_conflicts = flatten([for s in var.spec.sources : concat(
    [for h in keys(try(s.http.sensitive_headers, {})) : h if contains([for x in keys(try(s.http.headers, {})) : lower(x)], lower(h))],
    [for h in keys(try(s.grpc.sensitive_headers, {})) : h if contains([for x in keys(try(s.grpc.headers, {})) : lower(x)], lower(h))]
  )])

  replicas = try(var.spec.hpa.enabled, false) == true ? null : (try(var.spec.replicas, null) != null ? var.spec.replicas : 1)

  # Empty topology-spread selectors self-spread on flagd's own selector.
  topology_spread_constraints = [for c in try(var.spec.scheduling.topology_spread_constraints, []) : merge(c, {
    match_labels = length(try(c.match_labels, {})) > 0 ? c.match_labels : local.selector_labels
  })]
}
