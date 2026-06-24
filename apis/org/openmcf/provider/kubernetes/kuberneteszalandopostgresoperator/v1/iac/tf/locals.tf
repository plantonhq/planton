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
    "resource_kind" = "KubernetesZalandoPostgresOperator"
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

  # Namespace for Zalando Postgres Operator (from spec)
  namespace = var.spec.namespace

  # Service name
  service_name = "postgres-operator"

  # Helm chart configuration
  helm_chart_name       = "postgres-operator"
  helm_chart_repository = "https://opensource.zalando.com/postgres-operator/charts/postgres-operator"
  helm_chart_version    = "1.12.2"

  # Backup configuration
  has_backup_config = var.spec.backup_config != null

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  backup_secret_name    = "${var.metadata.name}-backup-credentials"
  backup_configmap_name = "${var.metadata.name}-backup-config"

  # R2 configuration (when backup is enabled). The bucket arrives as a resolved plain
  # name (Planton resolves the value-or-ref before tfvars).
  r2_account_id    = try(var.spec.backup_config.credentials.cloudflare_account_id, "")
  r2_bucket        = try(var.spec.backup_config.bucket, "")
  r2_object_prefix = try(var.spec.backup_config.object_prefix, "")
  r2_access_key_id = try(var.spec.backup_config.credentials.access_key_id, "")
  r2_secret_key    = try(var.spec.backup_config.credentials.secret_access_key, "")
  r2_endpoint      = local.has_backup_config ? "https://${local.r2_account_id}.r2.cloudflarestorage.com" : ""

  # Backup settings
  backup_schedule            = try(var.spec.backup_config.schedule, "")
  enable_wal_g_backup        = try(var.spec.backup_config.enable_wal_g_backup, true)
  enable_wal_g_restore       = try(var.spec.backup_config.enable_wal_g_restore, true)
  enable_clone_wal_g_restore = try(var.spec.backup_config.enable_clone_wal_g_restore, true)

  # WAL-G push target: s3://<bucket>[/<object_prefix>]/$(SCOPE)/$(PGVERSION). One operator
  # configmap serves every database, so Spilo/Patroni substitutes the suffix at runtime.
  walg_s3_prefix = !local.has_backup_config ? "" : (
    local.r2_object_prefix != "" ?
    "s3://${local.r2_bucket}/${local.r2_object_prefix}/$(SCOPE)/$(PGVERSION)" :
    "s3://${local.r2_bucket}/$(SCOPE)/$(PGVERSION)"
  )

  # Labels to be inherited by all PostgreSQL databases
  inherited_labels = [
    "resource",
    "organization",
    "environment",
    "resource_kind",
    "resource_id"
  ]

  # Cluster endpoint
  kube_endpoint = "${local.service_name}.${local.namespace}.svc.cluster.local"

  # Port-forward command
  port_forward_command = "kubectl port-forward svc/${local.service_name} -n ${local.namespace} 8080:8080"
}

