# Computed values for the KubernetesKarapace module. Every resolution
# here has an exact twin in the Pulumi module's locals.go /
# deployment.go — keep them in lockstep (same env keys rendered and
# omitted, same mount paths, same defaults).
#
# HCL DISCIPLINE: conditional entries are written as
# `cond ? value : null` and pruned (`if e != null` / `if v != ""`) —
# the null-prune idiom ONLY. Optional nested blocks are read with
# try(): HCL's && does not short-circuit, so chained null checks still
# dereference the null.

locals {
  namespace = var.spec.namespace

  # Registry objects are metadata.name; the optional REST-proxy role is
  # "<metadata.name>-rest".
  registry_name = var.metadata.name
  rest_name     = "${var.metadata.name}-rest"

  # The module's pinned upstream image, used when spec.image is empty.
  # The tag pins the aiven-open/karapace 6.2.1 release; bump it
  # deliberately, in lockstep with the Pulumi module's
  # vars.KarapaceImage.
  default_image = "ghcr.io/aiven-open/karapace:6.2.1"
  image         = try(coalesce(var.spec.image), "") != "" ? var.spec.image : local.default_image

  # ---- identity labels (kubernetesdeployment conventions) -------------------
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_labels = merge(
    {
      "resource"      = "true"
      "resource_id"   = local.resource_id
      "resource_kind" = "KubernetesKarapace"
      "resource_name" = var.metadata.name
    },
    var.metadata.org != null && var.metadata.org != "" ? { "organization" = var.metadata.org } : {},
    var.metadata.env != null && var.metadata.env != "" ? { "environment" = var.metadata.env } : {}
  )

  # Immutable pod-selection identity per role. Both Deployments run the
  # SAME image in the same namespace — the role-specific "app" value is
  # what keeps each Service from selecting the other role's pods
  # (selectors AND all their labels).
  registry_selector_labels = {
    "app"           = local.registry_name
    "resource_name" = var.metadata.name
  }
  rest_selector_labels = {
    "app"           = local.rest_name
    "resource_name" = var.metadata.name
  }

  registry_labels = merge(local.base_labels, local.registry_selector_labels)
  rest_labels     = merge(local.base_labels, local.rest_selector_labels)

  # ---- resolved spec defaults -------------------------------------------------
  registry_replicas = try(var.spec.replicas, null) != null ? var.spec.replicas : 1
  registry_port     = try(var.spec.port, null) != null ? var.spec.port : 8081

  rest_enabled  = try(var.spec.rest_proxy.enabled, false)
  rest_replicas = try(coalesce(var.spec.rest_proxy.replicas), 0) != 0 ? var.spec.rest_proxy.replicas : 1
  rest_port     = try(coalesce(var.spec.rest_proxy.port), 0) != 0 ? var.spec.rest_proxy.port : 8082

  # Scheme the registry serves on: https when spec.server_tls is set.
  # Drives the endpoint output, the probe scheme, the advertised
  # protocol (leader forwarding), and the REST proxy's registry_scheme
  # wiring. The REST proxy itself always serves plain HTTP.
  server_tls = try(var.spec.server_tls, null)
  scheme     = local.server_tls != null ? "https" : "http"

  topic_name    = try(coalesce(var.spec.registry.topic_name), "") != "" ? var.spec.registry.topic_name : "_schemas"
  compatibility = try(coalesce(var.spec.registry.compatibility), "") != "" ? var.spec.registry.compatibility : "BACKWARD"

  # LEADER-ELECTION CONTRACT: replicas of ONE installation coordinate
  # leadership through this Kafka consumer group, so the group id must
  # be unique per installation sharing a Kafka cluster — two
  # installations sharing a group id fight over leadership and corrupt
  # each other's view of the schemas topic. metadata.name is the natural
  # per-installation default.
  group_id = try(coalesce(var.spec.registry.group_id), "") != "" ? var.spec.registry.group_id : local.registry_name

  master_election_strategy = try(coalesce(var.spec.registry.master_election_strategy), "") != "" ? var.spec.registry.master_election_strategy : "lowest"

  log_level         = try(coalesce(var.spec.log_level), "") != "" ? var.spec.log_level : "INFO"
  security_protocol = try(coalesce(var.spec.kafka.security_protocol), "") != "" ? var.spec.kafka.security_protocol : "PLAINTEXT"

  # ---- Secret mount points -----------------------------------------------------
  # Configuration reaches the engine as file PATHS (ssl_cafile,
  # server_tls_certfile, registry_authfile, ...), so every Secret mounts
  # at a fixed directory and the env vars point inside it.
  kafka_ca_mount_path          = "/etc/karapace/kafka-ca"
  kafka_client_cert_mount_path = "/etc/karapace/kafka-cert"
  server_tls_mount_path        = "/etc/karapace/server-tls"
  authfile_mount_path          = "/etc/karapace/auth"

  # ---- Kafka connection blocks ---------------------------------------------------
  kafka_tls  = try(var.spec.kafka.tls, null)
  kafka_sasl = try(var.spec.kafka.sasl, null)

  kafka_tls_ca_certificate     = try(coalesce(local.kafka_tls.ca_certificate), "") != "" ? local.kafka_tls.ca_certificate : "ca.crt"
  kafka_tls_client_cert_secret = try(coalesce(local.kafka_tls.client_cert_secret_name), "")
  kafka_tls_client_certificate = try(coalesce(local.kafka_tls.client_certificate), "") != "" ? local.kafka_tls.client_certificate : "user.crt"
  kafka_tls_client_key         = try(coalesce(local.kafka_tls.client_key), "") != "" ? local.kafka_tls.client_key : "user.key"

  # SASL password wiring. The password NEVER lands in the pod spec as a
  # plaintext env value: when the spec declares a literal password the
  # module materializes it into the Secret "<metadata.name>-sasl" (key
  # "password") and the env var references that Secret — pod specs are
  # world-readable to anyone with get-pod RBAC, Secret values have their
  # own ACL. When the spec references an existing Secret
  # (password_secret), the env var references it directly.
  sasl_secret_name   = "${var.metadata.name}-sasl"
  create_sasl_secret = local.kafka_sasl != null && try(local.kafka_sasl.password_secret, null) == null && try(local.kafka_sasl.password, "") != ""

  sasl_password_secret_name = (
    local.kafka_sasl == null ? "" :
    try(local.kafka_sasl.password_secret, null) != null ? local.kafka_sasl.password_secret.secret_name :
    local.sasl_secret_name
  )
  sasl_password_secret_key = (
    local.kafka_sasl == null ? "" :
    try(local.kafka_sasl.password_secret, null) != null ? (
      try(coalesce(local.kafka_sasl.password_secret.key), "") != "" ? local.kafka_sasl.password_secret.key : "password"
    ) :
    "password"
  )

  # ---- server TLS / HTTP auth key resolution -------------------------------------
  server_tls_certificate = try(coalesce(local.server_tls.certificate), "") != "" ? local.server_tls.certificate : "tls.crt"
  server_tls_key         = try(coalesce(local.server_tls.key), "") != "" ? local.server_tls.key : "tls.key"

  http_basic   = try(var.spec.http_authentication.basic, null)
  http_oidc    = try(var.spec.http_authentication.oidc, null)
  authfile_key = try(coalesce(local.http_basic.key), "") != "" ? local.http_basic.key : "authfile.json"

  # ---- env lists (plain name/value pairs; ORDER mirrors the Pulumi module) --------
  # Entries needing valueFrom — the per-pod advertised hostname fieldRef
  # and the SASL password secretKeyRef — cannot ride a name/value list,
  # so they render as dedicated blocks in main.tf and each role's env
  # splits into head (before the fieldRef) and tail (after the
  # connection env + password ref) segments. main.tf interleaves:
  # head → ADVERTISED_HOSTNAME → kafka_connection_env → SASL password →
  # tail, matching the Pulumi module's rendered order exactly.

  # Env shared by both roles: the Kafka connection.
  kafka_connection_env = [
    for e in [
      { name = "KARAPACE_BOOTSTRAP_URI", value = var.spec.kafka.bootstrap_servers },
      { name = "KARAPACE_SECURITY_PROTOCOL", value = local.security_protocol },
      { name = "KARAPACE_LOG_LEVEL", value = local.log_level },
      # Kafka TLS file paths point into the Secret mounts (config.py:
      # ssl_cafile / ssl_certfile / ssl_keyfile).
      local.kafka_tls != null ? { name = "KARAPACE_SSL_CAFILE", value = "${local.kafka_ca_mount_path}/${local.kafka_tls_ca_certificate}" } : null,
      local.kafka_tls != null && local.kafka_tls_client_cert_secret != "" ? { name = "KARAPACE_SSL_CERTFILE", value = "${local.kafka_client_cert_mount_path}/${local.kafka_tls_client_certificate}" } : null,
      local.kafka_tls != null && local.kafka_tls_client_cert_secret != "" ? { name = "KARAPACE_SSL_KEYFILE", value = "${local.kafka_client_cert_mount_path}/${local.kafka_tls_client_key}" } : null,
      local.kafka_sasl != null ? { name = "KARAPACE_SASL_MECHANISM", value = local.kafka_sasl.mechanism } : null,
      local.kafka_sasl != null ? { name = "KARAPACE_SASL_PLAIN_USERNAME", value = local.kafka_sasl.username } : null,
    ] : e if e != null
  ]

  # Registry-role head: role flags first (this process serves ONLY the
  # schema-registry API), then the serve address.
  registry_env_head = [
    { name = "KARAPACE_KARAPACE_REGISTRY", value = "true" },
    { name = "KARAPACE_KARAPACE_REST", value = "false" },
    { name = "KARAPACE_HOST", value = "0.0.0.0" },
    { name = "KARAPACE_PORT", value = tostring(local.registry_port) },
  ]

  # Registry-role tail: behavior knobs, server TLS, HTTP authentication.
  registry_env_tail = [
    for e in [
      { name = "KARAPACE_TOPIC_NAME", value = local.topic_name },
      try(var.spec.registry.replication_factor, null) != null ? { name = "KARAPACE_REPLICATION_FACTOR", value = tostring(var.spec.registry.replication_factor) } : null,
      { name = "KARAPACE_COMPATIBILITY", value = local.compatibility },
      { name = "KARAPACE_GROUP_ID", value = local.group_id },
      { name = "KARAPACE_MASTER_ELECTION_STRATEGY", value = local.master_election_strategy },
      # The advertised protocol must follow the serving scheme: the
      # leader publishes `protocol://advertised_hostname:port` as the
      # master URL and followers forward writes to it — an https server
      # advertising http would break follower forwarding.
      local.server_tls != null ? { name = "KARAPACE_ADVERTISED_PROTOCOL", value = "https" } : null,
      local.server_tls != null ? { name = "KARAPACE_SERVER_TLS_CERTFILE", value = "${local.server_tls_mount_path}/${local.server_tls_certificate}" } : null,
      local.server_tls != null ? { name = "KARAPACE_SERVER_TLS_KEYFILE", value = "${local.server_tls_mount_path}/${local.server_tls_key}" } : null,
      local.http_basic != null ? { name = "KARAPACE_REGISTRY_AUTHFILE", value = "${local.authfile_mount_path}/${local.authfile_key}" } : null,
      local.http_oidc != null ? { name = "KARAPACE_SASL_OAUTHBEARER_AUTHENTICATION_ENABLED", value = "true" } : null,
      local.http_oidc != null ? { name = "KARAPACE_SASL_OAUTHBEARER_JWKS_ENDPOINT_URL", value = local.http_oidc.jwks_endpoint_url } : null,
      local.http_oidc != null && try(local.http_oidc.expected_issuer, "") != "" ? { name = "KARAPACE_SASL_OAUTHBEARER_EXPECTED_ISSUER", value = local.http_oidc.expected_issuer } : null,
      local.http_oidc != null && try(local.http_oidc.expected_audience, "") != "" ? { name = "KARAPACE_SASL_OAUTHBEARER_EXPECTED_AUDIENCE", value = local.http_oidc.expected_audience } : null,
    ] : e if e != null
  ]

  # REST-proxy-role head: role flags flipped, same engine image.
  rest_env_head = [
    { name = "KARAPACE_KARAPACE_REST", value = "true" },
    { name = "KARAPACE_KARAPACE_REGISTRY", value = "false" },
    { name = "KARAPACE_HOST", value = "0.0.0.0" },
    { name = "KARAPACE_PORT", value = tostring(local.rest_port) },
  ]

  # REST-proxy-role tail: the schema-registry wiring at the registry
  # Service (scheme follows the registry's server_tls posture).
  rest_env_tail = [
    { name = "KARAPACE_REGISTRY_SCHEME", value = local.scheme },
    { name = "KARAPACE_REGISTRY_HOST", value = "${local.registry_name}.${local.namespace}.svc.cluster.local" },
    { name = "KARAPACE_REGISTRY_PORT", value = tostring(local.registry_port) },
  ]

  # ---- Secret-backed volumes per role ----------------------------------------------
  # Kafka-side mounts are shared by both roles; server TLS and the Basic
  # authfile are registry-only.
  kafka_volumes = [
    for v in [
      local.kafka_tls != null ? { name = "kafka-ca", secret_name = local.kafka_tls.ca_secret_name, mount_path = local.kafka_ca_mount_path } : null,
      local.kafka_tls != null && local.kafka_tls_client_cert_secret != "" ? { name = "kafka-cert", secret_name = local.kafka_tls_client_cert_secret, mount_path = local.kafka_client_cert_mount_path } : null,
    ] : v if v != null
  ]

  registry_volumes = concat(
    local.kafka_volumes,
    [
      for v in [
        local.server_tls != null ? { name = "server-tls", secret_name = local.server_tls.secret_name, mount_path = local.server_tls_mount_path } : null,
        local.http_basic != null ? { name = "authfile", secret_name = local.http_basic.secret_name, mount_path = local.authfile_mount_path } : null,
      ] : v if v != null
    ]
  )

  rest_volumes = local.kafka_volumes

  # ---- resources (null-pruned; omitted maps let Kubernetes default) ------------------
  registry_resource_limits = {
    for k, v in {
      cpu    = try(var.spec.resources.limits.cpu, "")
      memory = try(var.spec.resources.limits.memory, "")
    } : k => v if v != ""
  }
  registry_resource_requests = {
    for k, v in {
      cpu    = try(var.spec.resources.requests.cpu, "")
      memory = try(var.spec.resources.requests.memory, "")
    } : k => v if v != ""
  }

  rest_resource_limits = {
    for k, v in {
      cpu    = try(var.spec.rest_proxy.resources.limits.cpu, "")
      memory = try(var.spec.rest_proxy.resources.limits.memory, "")
    } : k => v if v != ""
  }
  rest_resource_requests = {
    for k, v in {
      cpu    = try(var.spec.rest_proxy.resources.requests.cpu, "")
      memory = try(var.spec.rest_proxy.resources.requests.memory, "")
    } : k => v if v != ""
  }

  # ---- probe scheme ---------------------------------------------------------------------
  # PROBES ON /_health: the engine's health endpoint is in config.py's
  # skip-auth path list, so it answers without credentials even when
  # HTTP authentication is enabled — the one path that is always safe to
  # probe (the upstream image's own Docker HEALTHCHECK curls it too).
  # The registry probe scheme follows server_tls; the REST proxy always
  # serves plain HTTP.
  health_check_path     = "/_health"
  registry_probe_scheme = local.server_tls != null ? "HTTPS" : "HTTP"

  # ---- output endpoints -------------------------------------------------------------
  endpoint            = "${local.scheme}://${local.registry_name}.${local.namespace}.svc.cluster.local:${local.registry_port}"
  rest_proxy_endpoint = local.rest_enabled ? "http://${local.rest_name}.${local.namespace}.svc.cluster.local:${local.rest_port}" : ""
}
