resource "kubernetes_manifest" "database" {
  depends_on = [kubernetes_namespace_v1.postgres_namespace]

  manifest = {
    apiVersion = "acid.zalan.do/v1"
    kind       = "postgresql"
    metadata = {
      # For the Zalando operator, the name must be prefixed by the teamId (which is "db")
      # followed by metadata.name (matching the Pulumi module, which names the CR
      # "<teamId>-<metadata.name>"). The operator creates a service with this same name.
      name      = "db-${var.metadata.name}"
      namespace = local.namespace_name
      labels    = local.final_labels
    }
    spec = merge(
      {
        # Number of PostgreSQL instances (replicas)
        numberOfInstances = var.spec.container.replicas

        # Patroni configuration (empty object to satisfy CRD schema)
        patroni = {}

        # Pod annotations
        podAnnotations = {
          "postgres-cluster-id" = var.metadata.name
        }

        # PostgreSQL settings
        postgresql = {
          version = "14"
          parameters = {
            "max_connections" = "200"
          }
        }

        # Resource allocations
        resources = {
          limits = {
            cpu    = var.spec.container.resources.limits.cpu
            memory = var.spec.container.resources.limits.memory
          }
          requests = {
            cpu    = var.spec.container.resources.requests.cpu
            memory = var.spec.container.resources.requests.memory
          }
        }

        # Team ID is required by the Zalando operator
        teamId = "db"

        # Persistent volume configuration
        volume = {
          size = var.spec.container.disk_size
        }
      },
      # Conditionally add databases if specified
      # Convert list of database objects to map[string]string for Zalando operator
      length(var.spec.databases) > 0 ? {
        databases = { for db in var.spec.databases : db.name => db.owner_role }
      } : {},
      # Conditionally add users if specified
      # Convert list of users to map[string][]string format expected by Zalando operator
      length(var.spec.users) > 0 ? {
        users = { for user in var.spec.users : user.name => user.flags }
      } : {},
      # Disaster-recovery standby block (Zalando spec.standby); present only when
      # backup_config.restore.enabled. Matches the Pulumi restore_config.go behavior.
      local.standby_block != null ? {
        standby = local.standby_block
      } : {},
      # Merged backup + standby environment overrides on the operator pods.
      # Matches the Pulumi backup_config.go + restore_config.go env merge (standby first).
      length(local.all_env) > 0 ? {
        env = local.all_env
      } : {}
    )
  }
}
