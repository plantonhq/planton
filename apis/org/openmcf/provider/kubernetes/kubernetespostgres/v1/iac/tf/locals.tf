locals {
  # The Pulumi module exports the resource_id LABEL only when metadata.id is set;
  # everything else (CR name, service, secret names) keys off metadata.name. We
  # follow the same basis so tofu- and pulumi-provisioned Postgres are identical.
  # (metadata.id is optional, so it is null when unset -- normalize to "" because
  # `try` only catches errors, not nulls.)
  resource_id = var.metadata.id == null ? "" : var.metadata.id

  # Resource identity labels mirror the Pulumi module exactly: the kuberneteslabelkeys
  # keys are planton.ai/<key> and resource_kind is the CloudResourceKind enum string
  # ("KubernetesPostgres"). Kept byte-for-byte identical to avoid drift between engines.
  final_labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      "planton.ai/kind"     = "KubernetesPostgres"
    },
    local.resource_id != "" ? { "planton.ai/id" = local.resource_id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {},
  )

  # The external LoadBalancer selects the Spilo pods the Zalando operator creates.
  # Those pods carry application=spilo -- NOT our resource labels -- so the selector
  # must be application=spilo (matching the Pulumi module), or it matches zero pods.
  postgres_pod_selector_labels = {
    application = "spilo"
  }

  # Namespace is the requested spec namespace (Pulumi reads spec.namespace too).
  namespace = var.spec.namespace

  # Service name (matches the Pulumi `service` output).
  kube_service_name = "${var.metadata.name}-master"

  # Fully qualified domain name for the service.
  kube_service_fqdn = "${local.kube_service_name}.${local.namespace}.svc.cluster.local"

  # Handy port-forward command.
  kube_port_forward_command = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8080:8080"

  # Ingress configuration.
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)

  # Namespace reference: use created namespace name if created, otherwise use local.namespace.
  namespace_name = var.spec.create_namespace ? kubernetes_namespace_v1.postgres_namespace[0].metadata[0].name : local.namespace

  # Computed resource names to avoid conflicts when multiple instances share a namespace.
  external_lb_service_name = "${var.metadata.name}-external-lb"

  # ── Backup / restore ──
  backup_config   = try(var.spec.backup_config, null)
  backup_enabled  = try(local.backup_config.enabled, false)
  backup_creds    = try(local.backup_config.credentials, null)
  restore_config  = try(local.backup_config.restore, null)
  restore_enabled = try(local.restore_config.enabled, false)
  restore_creds   = local.restore_enabled ? try(local.restore_config.credentials, null) : null

  backup_creds_present  = local.backup_enabled && local.backup_creds != null
  restore_creds_present = local.restore_enabled && local.restore_creds != null

  # Names of the per-database R2 credential Secrets (created in credentials.tf).
  backup_r2_secret_name  = "${var.metadata.name}-backup-r2-credentials"
  restore_r2_secret_name = "${var.metadata.name}-restore-r2-credentials"

  # WAL-G push target: s3://<bucket>[/<object_prefix>]/$(SCOPE)/$(PGVERSION). Spilo/Patroni
  # substitutes the suffix at runtime. Empty when backups are disabled or no bucket is set.
  backup_bucket        = try(local.backup_config.bucket, "")
  backup_object_prefix = try(local.backup_config.object_prefix, "")
  backup_walg_s3_prefix = (local.backup_enabled && local.backup_bucket != "") ? (
    local.backup_object_prefix != "" ?
    format("s3://%s/%s/$(SCOPE)/$(PGVERSION)", local.backup_bucket, local.backup_object_prefix) :
    format("s3://%s/$(SCOPE)/$(PGVERSION)", local.backup_bucket)
  ) : ""

  # Zalando spec.standby block for cross-cluster disaster recovery. Present only when
  # restore is enabled; the database then bootstraps read-only from the source backup.
  standby_block = local.restore_enabled ? {
    s3_wal_path = format("s3://%s/%s", try(local.restore_config.bucket, ""), try(local.restore_config.object_prefix, ""))
  } : null

  # Per-database backup env. When credentials are set, the dedicated R2 target
  # (endpoint/region/path-style as plain values; credentials via secretKeyRef) is added.
  # Each entry is appended via a single-element conditional so plain ({name,value}) and
  # secretKeyRef ({name,valueFrom}) entries can coexist (concat preserves per-element
  # tuple types; a single mixed-shape ternary would not type-check).
  backup_env = !local.backup_enabled ? [] : concat(
    [{ name = "USE_WALG_BACKUP", value = "true" }],
    local.backup_walg_s3_prefix != "" ? [{ name = "WALG_S3_PREFIX", value = local.backup_walg_s3_prefix }] : [],
    try(local.backup_config.schedule, "") != "" ? [{ name = "BACKUP_SCHEDULE", value = local.backup_config.schedule }] : [],
    try(local.backup_config.retain_count, null) != null ? [{ name = "BACKUP_NUM_TO_RETAIN", value = tostring(local.backup_config.retain_count) }] : [],
    local.backup_creds_present ? [{ name = "AWS_ENDPOINT", value = format("https://%s.r2.cloudflarestorage.com", local.backup_creds.cloudflare_account_id) }] : [],
    local.backup_creds_present ? [{ name = "AWS_REGION", value = "auto" }] : [],
    local.backup_creds_present ? [{ name = "AWS_FORCE_PATH_STYLE", value = "true" }] : [],
    local.backup_creds_present ? [{ name = "USE_WALG_RESTORE", value = "true" }] : [],
    local.backup_creds_present ? [{ name = "AWS_ACCESS_KEY_ID", valueFrom = { secretKeyRef = { name = local.backup_r2_secret_name, key = "access_key_id" } } }] : [],
    local.backup_creds_present ? [{ name = "AWS_SECRET_ACCESS_KEY", valueFrom = { secretKeyRef = { name = local.backup_r2_secret_name, key = "secret_access_key" } } }] : [],
  )

  # STANDBY_* env for R2 access during standby bootstrap. Present only when restore is
  # enabled AND credentials are supplied; credentials via secretKeyRef.
  standby_env = concat(
    local.restore_creds_present ? [{ name = "STANDBY_AWS_ENDPOINT", value = format("https://%s.r2.cloudflarestorage.com", local.restore_creds.cloudflare_account_id) }] : [],
    local.restore_creds_present ? [{ name = "STANDBY_AWS_FORCE_PATH_STYLE", value = "true" }] : [],
    local.restore_creds_present ? [{ name = "STANDBY_AWS_ACCESS_KEY_ID", valueFrom = { secretKeyRef = { name = local.restore_r2_secret_name, key = "access_key_id" } } }] : [],
    local.restore_creds_present ? [{ name = "STANDBY_AWS_SECRET_ACCESS_KEY", valueFrom = { secretKeyRef = { name = local.restore_r2_secret_name, key = "secret_access_key" } } }] : [],
  )

  # Pulumi merges standby env first, then backup env.
  all_env = concat(local.standby_env, local.backup_env)
}
