# Computed values for the KubernetesAgentSandboxController module. Every
# resolution and patch here has an exact twin in the Pulumi module
# (settings.go, flags.go, workloads.go, compose.go); a parity test renders
# both engines' documents for the same specs and holds them equal.
#
# DEFAULTS: every default equals the spec's declared default, which equals
# upstream's own, and is resolved here so the module installs the same thing
# whether or not the platform's defaulting ran. A flag is written onto
# upstream's arguments only when its value differs, so upstream's defaults
# stay upstream's.
#
# HCL DISCIPLINE: optional blocks are read with try(), and conditional keys
# are contributed as merge() of `cond ? { key = value } : {}` singleton maps.

locals {
  # ---- the upstream contract (vars.go) ----------------------------------------
  # Pinned kubernetes-sigs/agent-sandbox release. The catalog's Sandbox,
  # SandboxTemplate, SandboxWarmPool and SandboxClaim kinds mirror its
  # resource definitions field for field, so moving it is a deliberate
  # catalog update. MUST equal the Pulumi module's vars.Release (a test holds
  # them equal).
  release           = "v1.0.7"
  bundle_name       = "agent-sandbox"
  namespace         = "agent-sandbox-system"
  upstream_registry = "registry.k8s.io"
  kind_name         = "KubernetesAgentSandboxController"
  owner_label       = "planton.ai/kind"

  controller_deployment_name = "agent-sandbox-controller"
  controller_container_name  = "agent-sandbox-controller"
  controller_service_name    = "agent-sandbox-controller"
  config_map_name            = "agent-sandbox-config"
  config_map_key             = "allowed-label-domains"
  config_checksum_annotation = "planton.ai/config-sha256"
  metrics_tls_mount_path     = "/etc/agent-sandbox/metrics-tls"

  # ---- who owns the namespace (settings.go) -------------------------------------
  # The namespace is upstream's; the spec only says who owns it. A reference
  # arrives here resolved to a name, so the preflight refuses any other name
  # the schema could not see.
  requested_namespace = var.spec.namespace
  create_namespace    = try(var.spec.create_namespace, false)

  # Identity labels on the Namespace this module creates; the kind label is
  # also the ownership mark the namespace check reads.
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

  router_deployment_name = "sandbox-router"
  router_container_name  = "sandbox-router"
  router_service_name    = "sandbox-router-svc"
  router_image_repo      = "registry.k8s.io/agent-sandbox/sandbox-router-go"
  router_tls_mount       = "/etc/sandbox-router/tls"
  router_client_ca_mount = "/etc/sandbox-router/client-ca"
  router_scoped_v1_mount = "/etc/sandbox-router/scoped-token-v1"
  router_scoped_v2_mount = "/etc/sandbox-router/scoped-token-v2"

  release_asset_base_url = "https://github.com/kubernetes-sigs/agent-sandbox/releases/download/"
  router_files_base_url  = "https://raw.githubusercontent.com/kubernetes-sigs/agent-sandbox/${local.release}/sandbox-router/deploy/"
  release_asset          = local.extensions ? "sandbox-with-extensions.yaml" : "sandbox.yaml"
  release_asset_url      = "${local.release_asset_base_url}${local.release}/${local.release_asset}"

  tls_versions = { tls10 = "VersionTLS10", tls11 = "VersionTLS11", tls12 = "VersionTLS12", tls13 = "VersionTLS13" }

  # ---- resolved settings (settings.go) ----------------------------------------
  extensions            = try(var.spec.extensions, null) != null ? coalesce(try(var.spec.extensions.enabled, null), true) : false
  allowed_label_domains = local.extensions ? try(var.spec.extensions.allowed_label_domains, []) : []
  config_map_content    = join("\n", local.allowed_label_domains)
  cluster_domain        = coalesce(try(var.spec.cluster_domain, null), "cluster.local")
  image_registry        = try(var.spec.image_registry, "")
  image_pull_secrets    = try(var.spec.image_pull_secrets, [])

  claim_workers                     = coalesce(try(var.spec.extensions.claim.concurrent_workers, null), 50)
  claim_warm_candidate_grace_period = coalesce(try(var.spec.extensions.claim.warm_candidate_grace_period, null), "2s")
  claim_events_disabled             = try(var.spec.extensions.claim.events_disabled, false)
  claim_observability_disabled      = try(var.spec.extensions.claim.observability_annotations_disabled, false)
  pool_workers                      = coalesce(try(var.spec.extensions.warm_pool.concurrent_workers, null), 1)
  pool_max_batch_size               = coalesce(try(var.spec.extensions.warm_pool.max_batch_size, null), 300)
  pool_replenish_delay              = coalesce(try(var.spec.extensions.warm_pool.replenish_delay, null), "0s")
  pool_max_refill_rate              = coalesce(try(var.spec.extensions.warm_pool.max_refill_rate, null), 0)
  pool_readiness_grace_period       = coalesce(try(var.spec.extensions.warm_pool.readiness_grace_period, null), "5m")
  pool_unschedulable_recheck        = coalesce(try(var.spec.extensions.warm_pool.unschedulable_recheck_interval, null), "1m")
  pool_eviction_enabled             = coalesce(try(var.spec.extensions.warm_pool.eviction_enabled, null), true)
  template_workers                  = coalesce(try(var.spec.extensions.template.concurrent_workers, null), 1)

  c_image             = local.folded_images.controller
  c_image_pull_secret = try(var.spec.controller.image.pull_secret_name, "")
  c_replicas          = coalesce(try(var.spec.controller.replicas, null), 1)
  c_resources         = local.resources.controller
  c_node_selector     = try(var.spec.controller.node_selector, {})
  c_tolerations       = local.tolerations.controller
  c_leader_elect      = coalesce(try(var.spec.controller.leader_election.enabled, null), true)
  c_leader_namespace  = try(var.spec.controller.leader_election.namespace, "")
  c_metrics_port      = coalesce(try(var.spec.controller.metrics.port, null), 8080)
  c_rest_metrics      = coalesce(try(var.spec.controller.metrics.rest_client_metrics_enabled, null), true)
  c_metrics_tls_on    = try(var.spec.controller.metrics.tls, null) != null ? coalesce(try(var.spec.controller.metrics.tls.enabled, null), true) : false
  c_metrics_tls_cert  = local.c_metrics_tls_on ? try(var.spec.controller.metrics.tls.certificate_secret, "") : ""
  c_metrics_tls_min   = local.c_metrics_tls_on ? lookup(local.tls_versions, try(var.spec.controller.metrics.tls.min_version, ""), "") : ""
  c_metrics_tls_ciph  = local.c_metrics_tls_on ? try(var.spec.controller.metrics.tls.cipher_suites, []) : []
  c_health_port       = coalesce(try(var.spec.controller.health_probe_port, null), 8081)
  c_sandbox_workers   = coalesce(try(var.spec.controller.sandbox.concurrent_workers, null), 100)
  c_sandbox_no_events = try(var.spec.controller.sandbox.events_disabled, false)
  c_write_behind      = coalesce(try(var.spec.controller.sandbox.write_behind_window, null), "0s")
  c_cache_selectors   = try(var.spec.controller.cache_label_selectors, false)
  c_qps               = try(var.spec.controller.kube_api.qps, null)
  c_burst             = coalesce(try(var.spec.controller.kube_api.burst, null), 10)
  c_connections       = coalesce(try(var.spec.controller.kube_api.connections, null), 1)
  c_separate_watch    = try(var.spec.controller.kube_api.separate_watch_connection, false)
  c_tracing           = try(var.spec.controller.tracing, null)
  c_pprof_cpu         = try(var.spec.controller.profiling.cpu_enabled, false)
  c_pprof_debug       = try(var.spec.controller.profiling.debug_enabled, false)
  c_block_rate        = coalesce(try(var.spec.controller.profiling.block_profile_rate, null), 1000000)
  c_mutex_fraction    = coalesce(try(var.spec.controller.profiling.mutex_profile_fraction, null), 10)
  c_log_devel         = try(var.spec.controller.logging.development, false)
  c_log_encoder       = try(var.spec.controller.logging.encoder, "")
  c_log_level         = try(var.spec.controller.logging.level, "")
  c_log_stacktrace    = try(var.spec.controller.logging.stacktrace_level, "")
  c_log_time          = try(var.spec.controller.logging.time_encoding, "")

  router_on = try(var.spec.router, null) != null ? coalesce(try(var.spec.router.enabled, null), true) : false

  r_image             = local.folded_images.router
  r_image_pull_secret = try(var.spec.router.image.pull_secret_name, "")
  r_replicas          = coalesce(try(var.spec.router.replicas, null), 2)
  r_resources         = local.resources.router
  r_node_selector     = try(var.spec.router.node_selector, {})
  r_tolerations       = local.tolerations.router
  r_min_available     = coalesce(try(var.spec.router.min_available, null), 1)
  r_token_review      = try(var.spec.router.authorization.token_review, null)
  r_scoped_token      = try(var.spec.router.authorization.scoped_token, null)
  r_authz_mode        = local.r_token_review != null ? "tokenreview" : (local.r_scoped_token != null ? "scoped-token" : "allow-all")
  r_tr_ttl            = coalesce(try(var.spec.router.authorization.token_review.ttl, null), "30s")
  r_tr_cache_size     = coalesce(try(var.spec.router.authorization.token_review.cache_size, null), 2048)
  r_tr_require_token  = try(var.spec.router.authorization.token_review.require_token, false)
  r_tr_audiences      = try(var.spec.router.authorization.token_review.audiences, [])
  r_scoped_v1         = try(var.spec.router.authorization.scoped_token.v1_secret, null)
  r_scoped_v2         = try(var.spec.router.authorization.scoped_token.verification_keys, null)
  r_scoped_until      = try(var.spec.router.authorization.scoped_token.v1_accept_until, "")
  r_cookie            = try(var.spec.router.authorization.cookie, null)
  r_tls_on            = try(var.spec.router.tls, null) != null ? coalesce(try(var.spec.router.tls.enabled, null), true) : false
  r_tls_cert          = local.r_tls_on ? var.spec.router.tls.certificate_secret : ""
  r_tls_port          = coalesce(try(var.spec.router.tls.port, null), 8443)
  r_mtls_mode         = local.r_tls_on ? try(var.spec.router.tls.mtls_mode, "") : ""
  r_client_ca         = local.r_tls_on ? try(var.spec.router.tls.client_ca, null) : null
  r_tls_min           = local.r_tls_on ? lookup(local.tls_versions, try(var.spec.router.tls.min_version, ""), "") : ""
  r_tls_ciphers       = local.r_tls_on ? try(var.spec.router.tls.cipher_suites, []) : []
  r_http_enabled      = local.r_tls_on ? coalesce(try(var.spec.router.tls.http_enabled, null), true) : true
  r_http_port         = coalesce(try(var.spec.router.ports.http, null), 8080)
  r_metrics_port      = coalesce(try(var.spec.router.ports.metrics, null), 9090)
  r_health_port       = coalesce(try(var.spec.router.ports.health_probe, null), 8081)
  r_proxy_timeout     = coalesce(try(var.spec.router.proxy.timeout, null), "180s")
  r_header_timeout    = coalesce(try(var.spec.router.proxy.response_header_timeout, null), "30s")
  r_shutdown_timeout  = coalesce(try(var.spec.router.proxy.shutdown_timeout, null), "30s")
  r_max_body_bytes    = try(var.spec.router.proxy.max_request_body_bytes, 0)
  r_allow_loopback    = try(var.spec.router.proxy.allow_loopback_pod_ip, false)
  r_path_prefix       = try(var.spec.router.proxy.path_routing_prefix, "")
  r_max_retries       = coalesce(try(var.spec.router.proxy.upstream_max_retries, null), 3)
  r_retry_initial     = coalesce(try(var.spec.router.proxy.upstream_retry_initial_delay, null), "200ms")
  r_retry_max         = coalesce(try(var.spec.router.proxy.upstream_retry_max_delay, null), "800ms")
  r_cache_enabled     = coalesce(try(var.spec.router.cache.enabled, null), true)
  r_cache_namespace   = try(var.spec.router.cache.namespace, "")
  r_telemetry         = try(var.spec.router.telemetry, null)
  r_traces_enabled    = coalesce(try(var.spec.router.telemetry.traces_enabled, null), true)
  r_otel_metrics      = coalesce(try(var.spec.router.telemetry.metrics_enabled, null), true)
  r_access_log        = coalesce(try(var.spec.router.access_log_enabled, null), true)
  r_netpol_on         = coalesce(try(var.spec.router.network_policy.enabled, null), true)
  r_client_ns_labels  = try(var.spec.router.network_policy.client_namespace_labels, {})
  r_metrics_namespace = coalesce(try(var.spec.router.network_policy.metrics_namespace, null), "monitoring")
  r_sandbox_ports     = length(try(var.spec.router.network_policy.sandbox_ports, [])) > 0 ? var.spec.router.network_policy.sandbox_ports : [8888]

  # Image overrides fold repo:tag; resources and tolerations drop empty keys
  # (settings.go imageOverride, containerResources, tolerations).
  folded_images = {
    for name, image in { controller = try(var.spec.controller.image, null), router = try(var.spec.router.image, null) } :
    name => (try(image.repo, "") != "" && try(image.tag, "") != "" ? "${image.repo}:${image.tag}" : "${try(image.repo, "")}${try(image.tag, "")}")
  }
  resources = {
    for name, res in { controller = try(var.spec.controller.resources, null), router = try(var.spec.router.resources, null) } :
    name => res == null ? null : (length(local.resource_parts[name]) == 0 ? null : local.resource_parts[name])
  }
  resource_parts = {
    for name, res in { controller = try(var.spec.controller.resources, null), router = try(var.spec.router.resources, null) } :
    name => { for part, values in {
      requests = { for k, v in { cpu = try(res.requests.cpu, ""), memory = try(res.requests.memory, "") } : k => v if v != "" }
      limits   = { for k, v in { cpu = try(res.limits.cpu, ""), memory = try(res.limits.memory, "") } : k => v if v != "" }
    } : part => values if length(values) > 0 }
  }
  tolerations = {
    for name, list in { controller = try(var.spec.controller.tolerations, []), router = try(var.spec.router.tolerations, []) } :
    name => [for t in list : { for k, v in {
      key               = t.key != "" ? t.key : null
      operator          = t.operator != "" ? t.operator : null
      value             = t.value != "" ? t.value : null
      effect            = t.effect != "" ? t.effect : null
      tolerationSeconds = t.toleration_seconds
    } : k => v if v != null }]
  }

  # ---- controller flags (flags.go controllerFlags) ------------------------------
  controller_flag_candidates = concat(
    [
      { on = !local.c_leader_elect, name = "leader-elect", value = "false" },
      { on = local.c_leader_namespace != "", name = "leader-election-namespace", value = local.c_leader_namespace },
      { on = local.cluster_domain != "cluster.local", name = "cluster-domain", value = local.cluster_domain },
      { on = local.c_metrics_port != 8080, name = "metrics-bind-address", value = ":${local.c_metrics_port}" },
      { on = local.c_health_port != 8081, name = "health-probe-bind-address", value = ":${local.c_health_port}" },
      { on = !local.c_rest_metrics, name = "enable-rest-client-metrics", value = "false" },
      { on = local.c_metrics_tls_on, name = "metrics-secure-serving", value = "true" },
      { on = local.c_metrics_tls_cert != "", name = "metrics-cert-dir", value = local.metrics_tls_mount_path },
      { on = local.c_metrics_tls_min != "", name = "tls-min-version", value = local.c_metrics_tls_min },
      { on = length(local.c_metrics_tls_ciph) > 0, name = "tls-cipher-suites", value = join(",", local.c_metrics_tls_ciph) },
      { on = local.c_sandbox_workers != 100, name = "sandbox-concurrent-workers", value = tostring(local.c_sandbox_workers) },
      { on = local.c_sandbox_no_events, name = "disable-sandbox-events", value = "true" },
      { on = local.c_write_behind != "0s", name = "sandbox-write-behind-window", value = local.c_write_behind },
      { on = local.c_cache_selectors, name = "cache-label-selectors", value = "true" },
      { on = local.c_qps != null, name = "kube-api-qps", value = local.c_qps != null ? tostring(local.c_qps) : "" },
      { on = local.c_burst != 10, name = "kube-api-burst", value = tostring(local.c_burst) },
      { on = local.c_connections != 1, name = "api-connections", value = tostring(local.c_connections) },
      { on = local.c_separate_watch, name = "separate-watch-connection", value = "true" },
      { on = local.c_tracing != null, name = "enable-tracing", value = "true" },
      { on = local.c_pprof_cpu, name = "enable-pprof", value = "true" },
      { on = local.c_pprof_debug, name = "enable-pprof-debug", value = "true" },
      { on = local.c_block_rate != 1000000, name = "pprof-block-profile-rate", value = tostring(local.c_block_rate) },
      { on = local.c_mutex_fraction != 10, name = "pprof-mutex-profile-fraction", value = tostring(local.c_mutex_fraction) },
      { on = local.c_log_devel, name = "zap-devel", value = "true" },
      { on = local.c_log_encoder != "", name = "zap-encoder", value = local.c_log_encoder },
      { on = local.c_log_level != "", name = "zap-log-level", value = local.c_log_level },
      { on = local.c_log_stacktrace != "", name = "zap-stacktrace-level", value = local.c_log_stacktrace },
      { on = local.c_log_time != "", name = "zap-time-encoding", value = local.c_log_time },
    ],
    local.extensions ? [
      { on = local.claim_workers != 50, name = "sandbox-claim-concurrent-workers", value = tostring(local.claim_workers) },
      { on = local.claim_warm_candidate_grace_period != "2s", name = "sandbox-claim-warm-candidate-grace-period", value = local.claim_warm_candidate_grace_period },
      { on = local.claim_events_disabled, name = "disable-claim-events", value = "true" },
      { on = local.claim_observability_disabled, name = "disable-claim-observability-annotations", value = "true" },
      { on = local.pool_workers != 1, name = "sandbox-warm-pool-concurrent-workers", value = tostring(local.pool_workers) },
      { on = local.pool_max_batch_size != 300, name = "sandbox-warm-pool-max-batch-size", value = tostring(local.pool_max_batch_size) },
      { on = local.pool_replenish_delay != "0s", name = "sandbox-warm-pool-replenish-delay", value = local.pool_replenish_delay },
      { on = local.pool_max_refill_rate != 0, name = "sandbox-warm-pool-max-refill-rate", value = tostring(local.pool_max_refill_rate) },
      { on = local.pool_readiness_grace_period != "5m", name = "sandbox-warm-pool-readiness-grace-period", value = local.pool_readiness_grace_period },
      { on = local.pool_unschedulable_recheck != "1m", name = "sandbox-warm-pool-unschedulable-recheck-interval", value = local.pool_unschedulable_recheck },
      { on = !local.pool_eviction_enabled, name = "enable-warm-pool-eviction", value = "false" },
      { on = local.template_workers != 1, name = "sandbox-template-concurrent-workers", value = tostring(local.template_workers) },
    ] : [],
  )
  controller_flags = [for f in local.controller_flag_candidates : f if f.on]

  # ---- router flags (flags.go routerFlags) --------------------------------------
  # The first seven are arguments upstream's deployment file already passes,
  # so they always take the resolved value; the rest only when they differ.
  router_flag_candidates = local.router_on ? [
    { on = true, name = "http-bind-address", value = local.r_http_enabled ? ":${local.r_http_port}" : "" },
    { on = true, name = "metrics-bind-address", value = ":${local.r_metrics_port}" },
    { on = true, name = "health-probe-bind-address", value = ":${local.r_health_port}" },
    { on = true, name = "cluster-domain", value = local.cluster_domain },
    { on = true, name = "proxy-timeout", value = local.r_proxy_timeout },
    { on = true, name = "upstream-max-retries", value = tostring(local.r_max_retries) },
    { on = true, name = "cache-enabled", value = tostring(local.r_cache_enabled) },
    { on = local.r_tls_on, name = "https-bind-address", value = ":${local.r_tls_port}" },
    { on = local.r_tls_on, name = "tls-cert-file", value = "${local.router_tls_mount}/tls.crt" },
    { on = local.r_tls_on, name = "tls-key-file", value = "${local.router_tls_mount}/tls.key" },
    { on = local.r_mtls_mode != "" && local.r_mtls_mode != "off", name = "mtls-mode", value = local.r_mtls_mode },
    { on = local.r_client_ca != null, name = "tls-client-ca-file", value = local.r_client_ca != null ? "${local.router_client_ca_mount}/${local.r_client_ca.key}" : "" },
    { on = local.r_tls_min != "", name = "tls-min-version", value = local.r_tls_min },
    { on = length(local.r_tls_ciphers) > 0, name = "tls-cipher-suites", value = join(",", local.r_tls_ciphers) },
    { on = local.r_header_timeout != "30s", name = "response-header-timeout", value = local.r_header_timeout },
    { on = local.r_shutdown_timeout != "30s", name = "shutdown-timeout", value = local.r_shutdown_timeout },
    { on = local.r_max_body_bytes != 0, name = "max-request-body-bytes", value = tostring(local.r_max_body_bytes) },
    { on = local.r_allow_loopback, name = "allow-loopback-pod-ip", value = "true" },
    { on = local.r_path_prefix != "", name = "path-routing-prefix", value = local.r_path_prefix },
    { on = local.r_retry_initial != "200ms", name = "upstream-retry-initial-delay", value = local.r_retry_initial },
    { on = local.r_retry_max != "800ms", name = "upstream-retry-max-delay", value = local.r_retry_max },
    # Both are stated whenever telemetry is configured: upstream turns them
    # on by itself once an OTLP endpoint is in its environment.
    { on = local.r_telemetry != null, name = "enable-tracing", value = tostring(local.r_traces_enabled) },
    { on = local.r_telemetry != null, name = "enable-otel-metrics", value = tostring(local.r_otel_metrics) },
    { on = !local.r_access_log, name = "access-log", value = "false" },
    { on = local.r_authz_mode != "allow-all", name = "authz-mode", value = local.r_authz_mode },
    { on = local.r_authz_mode == "tokenreview" && local.r_tr_ttl != "30s", name = "authz-tokenreview-ttl", value = local.r_tr_ttl },
    { on = local.r_authz_mode == "tokenreview" && local.r_tr_cache_size != 2048, name = "authz-tokenreview-cache-size", value = tostring(local.r_tr_cache_size) },
    { on = local.r_authz_mode == "tokenreview" && local.r_tr_require_token, name = "authz-tokenreview-require-token", value = "true" },
    { on = local.r_authz_mode == "tokenreview" && length(local.r_tr_audiences) > 0, name = "authz-tokenreview-audiences", value = join(",", local.r_tr_audiences) },
    { on = local.r_scoped_v1 != null, name = "authz-scoped-token-secret-file", value = local.r_scoped_v1 != null ? "${local.router_scoped_v1_mount}/${local.r_scoped_v1.key}" : "" },
    { on = local.r_scoped_v2 != null, name = "authz-scoped-token-verification-keys-file", value = local.r_scoped_v2 != null ? "${local.router_scoped_v2_mount}/${local.r_scoped_v2.key}" : "" },
    { on = local.r_scoped_until != "", name = "authz-scoped-token-v1-accept-until", value = local.r_scoped_until },
    { on = local.r_cookie != null, name = "authz-cookie-name", value = try(local.r_cookie.name, "") },
    { on = try(local.r_cookie.query_param, "") != "", name = "authz-cookie-query-param", value = try(local.r_cookie.query_param, "") },
    { on = !contains(["", "lax"], try(local.r_cookie.same_site, "")), name = "authz-cookie-samesite", value = try(local.r_cookie.same_site, "") },
    { on = try(local.r_cookie.insecure, false), name = "authz-cookie-insecure", value = "true" },
    { on = length(try(local.r_cookie.allowed_origins, [])) > 0, name = "authz-cookie-allowed-origins", value = join(",", try(local.r_cookie.allowed_origins, [])) },
    { on = try(local.r_cookie.trust_forwarded_proto, false), name = "authz-trust-forwarded-proto", value = "true" },
    { on = local.r_cache_namespace != "", name = "cache-namespace", value = local.r_cache_namespace },
  ] : []
  router_flags = [for f in local.router_flag_candidates : f if f.on]

  # applyFlags (flags.go): a flag upstream already passes keeps its position
  # and takes the new value; every other flag is appended in order.
  controller_args = local.controller_upstream_container == null ? [] : concat(
    [for a in try(local.controller_upstream_container.args, []) :
      contains(keys({ for f in local.controller_flags : f.name => f.value }), trimprefix(split("=", a)[0], "--"))
      ? "--${trimprefix(split("=", a)[0], "--")}=${ { for f in local.controller_flags : f.name => f.value }[trimprefix(split("=", a)[0], "--")]}"
      : a
    ],
    [for f in local.controller_flags : "--${f.name}=${f.value}"
    if !contains([for a in try(local.controller_upstream_container.args, []) : trimprefix(split("=", a)[0], "--")], f.name)],
  )
  router_args = local.router_upstream_container == null ? [] : concat(
    [for a in try(local.router_upstream_container.args, []) :
      contains(keys({ for f in local.router_flags : f.name => f.value }), trimprefix(split("=", a)[0], "--"))
      ? "--${trimprefix(split("=", a)[0], "--")}=${ { for f in local.router_flags : f.name => f.value }[trimprefix(split("=", a)[0], "--")]}"
      : a
    ],
    [for f in local.router_flags : "--${f.name}=${f.value}"
    if !contains([for a in try(local.router_upstream_container.args, []) : trimprefix(split("=", a)[0], "--")], f.name)],
  )

  # ---- upstream documents -------------------------------------------------------
  # Split on separator lines only, keyed by composed identity
  # (apiVersion//kind//name[//namespace]), the kubectl importer's ID form.
  release_documents = [
    for doc in split("\n---", "\n${data.http.release_asset.response_body}") : yamldecode(doc)
    if trimspace(doc) != "" && can(yamldecode(doc).kind)
  ]
  router_files = local.router_on ? concat(
    ["serviceaccount.yaml"],
    local.r_cache_enabled ? ["rbac.yaml"] : [],
    local.r_authz_mode == "tokenreview" ? ["rbac-tokenreview.yaml"] : [],
    ["deployment.yaml", "service.yaml"],
    local.r_min_available > 0 ? ["pdb.yaml"] : [],
    local.r_netpol_on ? ["networkpolicy.yaml"] : [],
  ) : []
  router_documents = flatten([
    for file in local.router_files : [
      for doc in split("\n---", "\n${data.http.router_file[file].response_body}") : yamldecode(doc)
      if trimspace(doc) != "" && can(yamldecode(doc).kind)
    ]
  ])
  documents = {
    for doc in concat(local.release_documents, local.router_documents) :
    join("//", concat([doc.apiVersion, doc.kind, doc.metadata.name], try(doc.metadata.namespace, "") != "" ? [doc.metadata.namespace] : [])) => doc
    if doc.kind != "CustomResourceDefinition"
  }

  namespace_id             = "v1//Namespace//${local.namespace}"
  controller_deployment_id = "apps/v1//Deployment//${local.controller_deployment_name}//${local.namespace}"
  controller_service_id    = "v1//Service//${local.controller_service_name}//${local.namespace}"
  config_map_id            = "v1//ConfigMap//${local.config_map_name}//${local.namespace}"
  router_deployment_id     = "apps/v1//Deployment//${local.router_deployment_name}//${local.namespace}"
  router_service_id        = "v1//Service//${local.router_service_name}//${local.namespace}"
  router_pdb_id            = "policy/v1//PodDisruptionBudget//${local.router_deployment_name}//${local.namespace}"
  router_netpol_id         = "networking.k8s.io/v1//NetworkPolicy//${local.router_deployment_name}//${local.namespace}"

  controller_upstream             = try(local.documents[local.controller_deployment_id], null)
  controller_upstream_container   = try(one([for c in local.controller_upstream.spec.template.spec.containers : c if c.name == local.controller_container_name]), null)
  router_upstream                 = try(local.documents[local.router_deployment_id], null)
  router_upstream_container       = try(one([for c in local.router_upstream.spec.template.spec.containers : c if c.name == local.router_container_name]), null)
  router_upstream_image_repo      = local.router_upstream_container == null ? "" : split(":", local.router_upstream_container.image)[0]
  router_netpol_upstream          = try(local.documents[local.router_netpol_id], null)
  router_netpol_shape_as_expected = local.router_netpol_upstream == null ? true : (length(local.router_netpol_upstream.spec.ingress) == 3 && length(local.router_netpol_upstream.spec.egress) == 4 && try(local.router_netpol_upstream.spec.ingress[0].ports[0].port, 0) == 8080 && try(local.router_netpol_upstream.spec.ingress[1].ports[0].port, 0) == 9090 && try(local.router_netpol_upstream.spec.ingress[2].ports[0].port, 0) == 8081 && try(local.router_netpol_upstream.spec.egress[1].ports[0].port, 0) == 8888)

  # resolveImage (workloads.go): an override wins, then the mirror moves an
  # upstream image, else upstream's stays.
  controller_image = local.c_image != "" ? local.c_image : (
    local.image_registry != "" && startswith(try(local.controller_upstream_container.image, ""), "${local.upstream_registry}/")
    ? "${local.image_registry}${trimprefix(local.controller_upstream_container.image, local.upstream_registry)}"
    : try(local.controller_upstream_container.image, "")
  )
  router_pinned_image = "${local.router_image_repo}:${local.release}"
  router_image = local.r_image != "" ? local.r_image : (
    local.image_registry != "" ? "${local.image_registry}${trimprefix(local.router_pinned_image, local.upstream_registry)}" : local.router_pinned_image
  )

  # telemetryEnv (workloads.go): the OpenTelemetry SDK's own variables.
  controller_env = concat(
    local.c_tracing != null ? [{ name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = local.c_tracing.otlp_endpoint }] : [],
    try(local.c_tracing.insecure, false) ? [{ name = "OTEL_EXPORTER_OTLP_INSECURE", value = "true" }] : [],
    try(local.c_tracing.headers, null) != null ? [{ name = "OTEL_EXPORTER_OTLP_HEADERS", valueFrom = { secretKeyRef = { name = local.c_tracing.headers.secret, key = local.c_tracing.headers.key } } }] : [],
  )
  router_env = concat(
    local.r_telemetry != null ? [{ name = "OTEL_EXPORTER_OTLP_ENDPOINT", value = local.r_telemetry.otlp_endpoint }] : [],
    try(local.r_telemetry.insecure, false) ? [{ name = "OTEL_EXPORTER_OTLP_INSECURE", value = "true" }] : [],
    try(local.r_telemetry.headers, null) != null ? [{ name = "OTEL_EXPORTER_OTLP_HEADERS", valueFrom = { secretKeyRef = { name = local.r_telemetry.headers.secret, key = local.r_telemetry.headers.key } } }] : [],
  )

  # pullSecrets (workloads.go): the spec's pull secrets plus an override's own,
  # each once, in that order.
  controller_pull_secrets = [for n in distinct(compact(concat(local.image_pull_secrets, [local.c_image_pull_secret]))) : { name = n }]
  router_pull_secrets     = [for n in distinct(compact(concat(local.image_pull_secrets, [local.r_image_pull_secret]))) : { name = n }]

  # The router's drain budget plus upstream's 15s margin (upstream pairs its
  # 30s default with 45s); durations are summed from their Go-duration parts.
  duration_unit_seconds = { ns = 0.000000001, us = 0.000001, ms = 0.001, s = 1, m = 60, h = 3600 }
  router_shutdown_seconds = sum(concat([0], [
    for part in regexall("([0-9]+(?:\\.[0-9]+)?)(ns|us|ms|s|m|h)", local.r_shutdown_timeout) :
    tonumber(part[0]) * local.duration_unit_seconds[part[1]]
  ]))
  router_termination_grace = ceil(local.router_shutdown_seconds) + 15
}
