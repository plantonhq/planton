# Computed values. Every resolution here has an exact twin in the Pulumi
# module (locals.go / bao_config.go / values.go) — keep them in lockstep.

locals {
  namespace    = var.spec.namespace
  release_name = var.metadata.name

  # Chart identity — the SERVED index truth
  # (https://openbao.github.io/openbao-helm); chart 0.28.6 = OpenBao
  # v2.6.1.
  helm_chart_name       = "openbao"
  helm_chart_repo       = "https://openbao.github.io/openbao-helm"
  default_chart_version = "0.28.6"
  chart_version         = try(coalesce(var.spec.chart_version), "") != "" ? var.spec.chart_version : local.default_chart_version

  # Chart constants: ports, mount paths (the config's stanzas and the
  # PVC/TLS mounts must agree).
  api_port        = 8200
  cluster_port    = 8201
  data_mount_path = "/openbao/data"
  tls_mount_path  = "/openbao/tls"

  # Planton governance labels for module-created satellites (namespace,
  # seal-credentials Secret) — never injected into the chart's own
  # resources; Helm owns those.
  labels = merge(
    {
      "planton.ai/resource"      = "true"
      "planton.ai/resource-name" = var.metadata.name
      "planton.ai/resource-kind" = "KubernetesOpenBao"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/resource-id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {},
  )

  # ------------------------------ mode ----------------------------------
  # The spec's mode oneof arrives as three nullable blocks; unset =
  # standalone (the chart's own default). Precedence dev > ha mirrors
  # nothing — the proto oneof guarantees at most one.
  mode = (
    try(var.spec.server.dev, null) != null ? "dev" :
    try(var.spec.server.ha, null) != null ? "ha" : "standalone"
  )
  ha_replicas = local.mode == "ha" ? coalesce(try(var.spec.server.ha.replicas, null), 3) : 1

  # ------------------------------ tls ------------------------------------
  tls_enabled     = try(var.spec.tls.enabled, false)
  tls_secret_name = local.tls_enabled ? try(var.spec.tls.cert_secret_name, "") : ""
  scheme          = local.tls_enabled ? "https" : "http"

  # ---------------------- synthesized server config ----------------------
  # The chart takes config as a raw HCL string written to a ConfigMap —
  # this module OWNS synthesizing it from typed fields (twin of
  # bao_config.go). SENSITIVE-MATERIAL RULE: only NON-credential seal
  # parameters render here; credential material rides env vars from the
  # module-owned Secret.
  ui_enabled      = coalesce(try(var.spec.ui_enabled, null), true)
  metrics_enabled = try(var.spec.metrics.enabled, false)

  config_listener_tls_lines = local.tls_enabled ? [
    "  tls_cert_file = \"${local.tls_mount_path}/tls.crt\"",
    "  tls_key_file = \"${local.tls_mount_path}/tls.key\"",
  ] : ["  tls_disable = 1"]

  config_listener_telemetry_lines = local.metrics_enabled ? [
    "  telemetry {",
    "    unauthenticated_metrics_access = \"true\"",
    "  }",
  ] : []

  config_listener_block = join("\n", concat(
    ["listener \"tcp\" {"],
    local.config_listener_tls_lines,
    [
      "  address = \"[::]:${local.api_port}\"",
      "  cluster_address = \"[::]:${local.cluster_port}\"",
    ],
    local.config_listener_telemetry_lines,
    ["}"],
  ))

  # THE RETRY_JOIN SYNTHESIS: the chart ships NO retry_join — without
  # these blocks a multi-replica Raft install never forms a cluster
  # (verified at chart 0.28.6). Peers are the StatefulSet pods' stable
  # DNS names through the headless `-internal` Service (fullnameOverride
  # pins the names). Joins are idempotent.
  config_retry_join_lines = local.mode == "ha" ? flatten([
    for i in range(local.ha_replicas) : concat(
      [
        "  retry_join {",
        "    leader_api_addr = \"${local.scheme}://${local.release_name}-${i}.${local.release_name}-internal:${local.api_port}\"",
      ],
      local.tls_enabled ? ["    leader_ca_cert_file = \"${local.tls_mount_path}/ca.crt\""] : [],
      ["  }"],
    )
  ]) : []

  config_storage_block = (
    local.mode == "standalone" ? join("\n", [
      "storage \"file\" {",
      "  path = \"${local.data_mount_path}\"",
      "}",
      ]) : local.mode == "ha" ? join("\n", concat(
      [
        "storage \"raft\" {",
        "  path = \"${local.data_mount_path}\"",
      ],
      local.config_retry_join_lines,
      [
        "}",
        "",
        # The server patches openbao-active/openbao-sealed labels onto
        # its own pod — what the chart's active/standby Services select
        # on.
        "service_registration \"kubernetes\" {}",
      ],
    )) : ""
  )

  # Seal stanza (non-credential parameters only). The proto oneof
  # guarantees at most one arm; each arm renders its own lines.
  seal_aws     = try(var.spec.auto_unseal.aws_kms, null)
  seal_gcp     = try(var.spec.auto_unseal.gcp_kms, null)
  seal_azure   = try(var.spec.auto_unseal.azure_key_vault, null)
  seal_transit = try(var.spec.auto_unseal.transit, null)

  config_seal_block = (
    local.seal_aws != null ? join("\n", [
      "seal \"awskms\" {",
      "  region = \"${local.seal_aws.region}\"",
      "  kms_key_id = \"${local.seal_aws.kms_key_id}\"",
      "}",
      ]) : local.seal_gcp != null ? join("\n", [
      "seal \"gcpckms\" {",
      "  project = \"${local.seal_gcp.project}\"",
      "  region = \"${local.seal_gcp.region}\"",
      "  key_ring = \"${local.seal_gcp.key_ring}\"",
      "  crypto_key = \"${local.seal_gcp.crypto_key}\"",
      "}",
      ]) : local.seal_azure != null ? join("\n", concat(
      [
        "seal \"azurekeyvault\" {",
        "  vault_name = \"${local.seal_azure.vault_name}\"",
        "  key_name = \"${local.seal_azure.key_name}\"",
        "  tenant_id = \"${local.seal_azure.tenant_id}\"",
      ],
      try(coalesce(local.seal_azure.client_id), "") != "" ? ["  client_id = \"${local.seal_azure.client_id}\""] : [],
      ["}"],
      )) : local.seal_transit != null ? join("\n", [
      "seal \"transit\" {",
      "  address = \"${local.seal_transit.address}\"",
      "  key_name = \"${local.seal_transit.key_name}\"",
      "  mount_path = \"${try(coalesce(local.seal_transit.mount_path), "") != "" ? local.seal_transit.mount_path : "transit/"}\"",
      "}",
    ]) : ""
  )

  config_telemetry_block = local.metrics_enabled ? join("\n", [
    "telemetry {",
    "  prometheus_retention_time = \"30s\"",
    "  disable_hostname = true",
    "}",
  ]) : ""

  # Dev mode renders NO config (`bao server -dev` ignores it). The
  # trailing newline matches bao_config.go's TrimRight+"\n" ending —
  # the rendered ConfigMaps must stay byte-identical across engines.
  bao_config_hcl = local.mode == "dev" ? "" : "${join("\n\n", compact([
    "ui = ${local.ui_enabled}",
    local.config_listener_block,
    local.config_storage_block,
    local.config_seal_block,
    local.config_telemetry_block,
  ]))}\n"

  # -------------------- seal credentials (env wiring) --------------------
  # Credential material a declared seal arm carries, keyed by the ENV VAR
  # it reaches the server as (the cloud SDKs' standard variables; transit
  # follows the wrapper's token env). Empty for keyless postures.
  seal_secret_data = merge(
    local.seal_aws != null && try(coalesce(local.seal_aws.secret_access_key), "") != "" ? { "AWS_SECRET_ACCESS_KEY" = local.seal_aws.secret_access_key } : {},
    local.seal_azure != null && try(coalesce(local.seal_azure.client_secret), "") != "" ? { "AZURE_CLIENT_SECRET" = local.seal_azure.client_secret } : {},
    local.seal_transit != null && try(coalesce(local.seal_transit.token), "") != "" ? { "VAULT_TOKEN" = local.seal_transit.token } : {},
  )
  seal_credentials_secret_name = length(local.seal_secret_data) > 0 ? "${var.metadata.name}-seal-credentials" : ""

  # Non-secret seal env (identifiers only).
  seal_plain_env = local.seal_aws != null ? merge(
    { "AWS_REGION" = local.seal_aws.region },
    try(coalesce(local.seal_aws.access_key_id), "") != "" ? { "AWS_ACCESS_KEY_ID" = local.seal_aws.access_key_id } : {},
  ) : {}

  # ServiceAccount annotations: the GCP seal arm's declared workload
  # identity contributes iam.gke.io/gcp-service-account; explicit
  # service_account.annotations win on conflict (twin of values.go).
  sa_annotations = merge(
    try(coalesce(local.seal_gcp.workload_identity_service_account), "") != "" ? {
      "iam.gke.io/gcp-service-account" = local.seal_gcp.workload_identity_service_account
    } : {},
    try(var.spec.service_account.annotations, {}) != null ? try(var.spec.service_account.annotations, {}) : {},
  )

  # ------------------------------ values ---------------------------------
  # The typed chart values (twin of values.go's buildHelmValues) — one
  # object literal per block, conditional keys null-pruned.
  server_block_raw = {
    dev = local.mode == "dev" ? { enabled = true } : null
    standalone = local.mode == "standalone" ? {
      enabled = true
      config  = local.bao_config_hcl
    } : null
    ha = local.mode == "ha" ? {
      enabled  = true
      replicas = local.ha_replicas
      raft = {
        enabled = true
        # Stable, human-readable Raft node IDs = pod names (without
        # this the server generates a GUID — persisted on the data PVC,
        # but opaque in every peer listing).
        setNodeId = true
        config    = local.bao_config_hcl
      }
    } : null

    resources = try(var.spec.server.resources, null) != null ? {
      for k, v in {
        requests = try(var.spec.server.resources.requests, null) != null ? {
          for k2, v2 in {
            cpu    = try(coalesce(var.spec.server.resources.requests.cpu), "")
            memory = try(coalesce(var.spec.server.resources.requests.memory), "")
          } : k2 => v2 if v2 != ""
        } : null
        limits = try(var.spec.server.resources.limits, null) != null ? {
          for k2, v2 in {
            cpu    = try(coalesce(var.spec.server.resources.limits.cpu), "")
            memory = try(coalesce(var.spec.server.resources.limits.memory), "")
          } : k2 => v2 if v2 != ""
        } : null
      } : k => v if v != null && v != {}
    } : null

    logLevel  = try(coalesce(var.spec.server.log_level), "") != "" ? var.spec.server.log_level : null
    logFormat = try(coalesce(var.spec.server.log_format), "") != "" ? var.spec.server.log_format : null

    nodeSelector = length(try(var.spec.server.scheduling.node_selector, {})) > 0 ? var.spec.server.scheduling.node_selector : null
    tolerations = length(try(var.spec.server.scheduling.tolerations, [])) > 0 ? [
      for t in var.spec.server.scheduling.tolerations : {
        for k, v in {
          key               = try(coalesce(t.key), "")
          operator          = try(coalesce(t.operator), "")
          value             = try(coalesce(t.value), "")
          effect            = try(coalesce(t.effect), "")
          tolerationSeconds = try(t.toleration_seconds, null)
        } : k => v if v != "" && v != null
      }
    ] : null

    # Data volume: consumed by the chart only in standalone/ha+raft (dev
    # is in-memory) — rendered unconditionally for explicitness.
    dataStorage = {
      for k, v in {
        enabled      = true
        size         = try(coalesce(var.spec.server.data_storage.size), "") != "" ? var.spec.server.data_storage.size : null
        storageClass = try(coalesce(var.spec.server.data_storage.storage_class), "") != "" ? var.spec.server.data_storage.storage_class : null
      } : k => v if v != null
    }

    auditStorage = try(var.spec.server.audit_storage, null) != null ? {
      for k, v in {
        enabled      = true
        size         = try(coalesce(var.spec.server.audit_storage.size), "") != "" ? var.spec.server.audit_storage.size : null
        storageClass = try(coalesce(var.spec.server.audit_storage.storage_class), "") != "" ? var.spec.server.audit_storage.storage_class : null
      } : k => v if v != null
    } : null

    # TLS: mount the certificate Secret where the synthesized listener
    # expects it. global.tlsDisable alone changes only probe schemes and
    # URLs — the listener lines above are the other half of the
    # composite switch.
    volumes = local.tls_enabled ? [{
      name   = "tls"
      secret = { secretName = local.tls_secret_name }
    }] : null
    volumeMounts = local.tls_enabled ? [{
      name      = "tls"
      mountPath = local.tls_mount_path
      readOnly  = true
    }] : null

    extraEnvironmentVars = length(local.seal_plain_env) > 0 ? local.seal_plain_env : null
    # Credential material reaches the server as env vars from the
    # module-owned Secret — never through the config ConfigMap. Keys
    # sorted for a deterministic rendering (twin of values.go).
    extraSecretEnvironmentVars = length(local.seal_secret_data) > 0 ? [
      for envName in sort(keys(local.seal_secret_data)) : {
        envName    = envName
        secretName = local.seal_credentials_secret_name
        secretKey  = envName
      }
    ] : null

    # ServiceAccount annotations: the GCP seal arm's declared workload
    # identity contributes iam.gke.io/gcp-service-account (the spec
    # field promises exactly this); explicit service_account.annotations
    # win on conflict. NOTE dev mode drops SA annotations (chart
    # behavior — taught on the spec field).
    serviceAccount = length(local.sa_annotations) > 0 ? {
      annotations = local.sa_annotations
    } : null
    authDelegator = try(var.spec.service_account.auth_delegator_enabled, null) != null ? {
      enabled = var.spec.service_account.auth_delegator_enabled
    } : null

    networkPolicy = try(var.spec.network_policy_enabled, false) ? { enabled = true } : null
  }

  server_block = { for k, v in local.server_block_raw : k => v if v != null }

  # THE INJECTOR IS OPT-IN — a deliberate divergence from the chart
  # default (which installs a cluster-wide mutating webhook on every
  # install); rendered explicitly either way.
  #
  # SINGLE null-pruned object, never `cond ? {rich} : {enabled=false}`:
  # two-arm conditionals with different object shapes are the HCL
  # type-unification class this program keeps catching — Terraform
  # cannot unify the arms and the plan fails the moment the injector is
  # enabled. `enabled` is the boolean itself; every other key resolves
  # null unless the injector is on, and the prune drops it.
  injector_enabled = try(var.spec.injector.enabled, false)
  injector_block = { for k, v in {
    enabled  = local.injector_enabled
    replicas = local.injector_enabled && try(var.spec.injector.replicas, null) != null ? var.spec.injector.replicas : null
    webhook = local.injector_enabled && try(coalesce(var.spec.injector.failure_policy), "") != "" ? {
      failurePolicy = var.spec.injector.failure_policy
    } : null
    resources = local.injector_enabled && try(var.spec.injector.resources, null) != null ? {
      for k2, v2 in {
        requests = try(var.spec.injector.resources.requests, null) != null ? {
          for k3, v3 in {
            cpu    = try(coalesce(var.spec.injector.resources.requests.cpu), "")
            memory = try(coalesce(var.spec.injector.resources.requests.memory), "")
          } : k3 => v3 if v3 != ""
        } : null
        limits = try(var.spec.injector.resources.limits, null) != null ? {
          for k3, v3 in {
            cpu    = try(coalesce(var.spec.injector.resources.limits.cpu), "")
            memory = try(coalesce(var.spec.injector.resources.limits.memory), "")
          } : k3 => v3 if v3 != ""
        } : null
      } : k2 => v2 if v2 != null && v2 != {}
    } : null
  } : k => v if v != null }

  typed_helm_values_raw = {
    global = { tlsDisable = !local.tls_enabled }
    server = local.server_block
    # The ui Service toggle; the listener-side `ui = true` lives in the
    # synthesized config — one spec field drives both.
    ui       = { enabled = local.ui_enabled }
    injector = local.injector_block
    serverTelemetry = try(var.spec.metrics.service_monitor_enabled, false) ? {
      serviceMonitor = { enabled = true }
    } : null
  }

  typed_helm_values = { for k, v in local.typed_helm_values_raw : k => v if v != null }

  # ------------------------------ outputs --------------------------------
  api_endpoint = "${local.scheme}://${local.release_name}.${local.namespace}.svc.cluster.local:${local.api_port}"

  # NAME BUDGET (chart truth at 0.28.6): the chart truncates its
  # fullname at 63 then APPENDS Service suffixes — `-internal` (9)
  # always, `-agent-injector-svc` (19) with the injector; Services cap
  # at 63. Enforced by the release precondition in main.tf.
  max_name_length = try(var.spec.injector.enabled, false) ? 44 : 54
}
