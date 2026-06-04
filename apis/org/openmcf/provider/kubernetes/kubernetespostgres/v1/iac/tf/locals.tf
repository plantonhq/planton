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

  # ── Backup / restore (mirror of the Pulumi backup_config.go + restore_config.go) ──
  backup_config   = try(var.spec.backup_config, null)
  restore         = try(local.backup_config.restore, null)
  restore_enabled = try(local.restore.enabled, false)

  # Zalando spec.standby block for cross-cluster disaster recovery. Present only when
  # restore.enabled; the database then bootstraps read-only from the backup WAL path.
  standby_block = local.restore_enabled ? {
    s3_wal_path = format("s3://%s/%s", local.restore.bucket_name, local.restore.s3_path)
  } : null

  # Per-database backup env overrides (override operator-level settings).
  backup_env = local.backup_config == null ? [] : concat(
    try(local.backup_config.s3_prefix, "") != "" ? [{ name = "WALG_S3_PREFIX", value = format("s3://%s", local.backup_config.s3_prefix) }] : [],
    try(local.backup_config.backup_schedule, "") != "" ? [{ name = "BACKUP_SCHEDULE", value = local.backup_config.backup_schedule }] : [],
    try(local.backup_config.enable_backup, null) != null ? [{ name = "USE_WALG_BACKUP", value = local.backup_config.enable_backup ? "true" : "false" }] : [],
  )

  # STANDBY_* env for R2 access during standby bootstrap. Present only when restore is
  # enabled AND r2_config is supplied (matches restore_config.go).
  standby_env = (local.restore_enabled && try(local.restore.r2_config, null) != null) ? [
    { name = "STANDBY_AWS_ENDPOINT", value = format("https://%s.r2.cloudflarestorage.com", local.restore.r2_config.cloudflare_account_id) },
    { name = "STANDBY_AWS_FORCE_PATH_STYLE", value = "true" },
    { name = "STANDBY_AWS_ACCESS_KEY_ID", value = local.restore.r2_config.access_key_id },
    { name = "STANDBY_AWS_SECRET_ACCESS_KEY", value = local.restore.r2_config.secret_access_key },
  ] : []

  # Pulumi merges standby env first, then backup env.
  all_env = concat(local.standby_env, local.backup_env)
}
