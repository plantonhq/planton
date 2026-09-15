variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Database for MySQL Flexible Server specification"
  type = object({
    # The Azure region the server lives in.
    region = string

    # The resource group the server lives in. References are resolved to a
    # literal name by the platform before the module runs.
    resource_group = string

    # The server's name: 3-63 lowercase letters/digits/hyphens, globally
    # unique (it becomes {server_name}.mysql.database.azure.com).
    server_name = string

    # How the server comes into existence, as the spec enum's name string
    # (DEFAULT, POINT_IN_TIME_RESTORE, REPLICA, GEO_RESTORE). Unset means
    # DEFAULT (a fresh, empty server).
    create_mode = optional(string)

    # For the replica/restore modes: the resolved ARM ID of the source
    # server.
    source_server_id = optional(string)

    # For POINT_IN_TIME_RESTORE: the RFC-3339 UTC instant to restore to.
    point_in_time_restore_time_in_utc = optional(string)

    # Replica promotion: "NONE" breaks replication and promotes the
    # replica to a standalone primary. Never sent at creation (Azure
    # rejects it); only meaningful as a day-2 update on a replica.
    replication_role = optional(string)

    # MySQL password-auth admin credentials. Empty on replicas/restores,
    # which inherit from the source. The login is fixed once set.
    administrator_login    = optional(string)
    administrator_password = optional(string)

    # The MySQL version ("5.7", "8.0.21", "8.4"). Only sent for a fresh
    # (DEFAULT) server -- replicas and restores inherit the source's
    # version, so sending one would fight the service.
    version = optional(string)

    # The compute SKU ({TIER}_Standard_{SIZE}). Empty on a replica
    # inherits the source's SKU.
    sku_name = optional(string)

    # The storage profile: capacity in GiB, provisioned IOPS or elastic
    # IOPS scaling (mutually exclusive), auto-grow, and slow-query-log
    # placement. Omitted applies Azure's defaults.
    storage = optional(object({
      size_gb             = optional(number)
      iops                = optional(number)
      auto_grow_enabled   = optional(bool, true)
      io_scaling_enabled  = optional(bool, false)
      log_on_disk_enabled = optional(bool, false)
    }))

    # The primary's availability zone ("1", "2", "3"); unset lets Azure
    # choose.
    zone = optional(string)

    # High availability: presence enables a standby with automatic
    # failover. mode arrives as the spec enum's name string
    # (ZONE_REDUNDANT/SAME_ZONE).
    high_availability = optional(object({
      mode                      = string
      standby_availability_zone = optional(string)
    }))

    # The weekly maintenance window; omit for a system-managed window.
    maintenance_window = optional(object({
      day_of_week  = optional(number)
      start_hour   = optional(number)
      start_minute = optional(number)
    }))

    # Backup retention days, 1-35 (Azure's default is 7).
    backup_retention_days = optional(number, 7)

    # Whether backups replicate to the paired region (fixed at creation).
    geo_redundant_backup_enabled = optional(bool, false)

    # Public-endpoint control, as the spec enum's name string
    # (ENABLED/DISABLED). Unset lets Azure derive it: enabled for a
    # public server, disabled when VNet-injected.
    public_network_access = optional(string)

    # VNet injection: the resolved ARM IDs of the delegated subnet and the
    # private DNS zone the server registers in. Set together; fixed at
    # creation.
    delegated_subnet_id = optional(string)
    private_dns_zone_id = optional(string)

    # The user-assigned identities attached to the server, by resolved
    # ARM ID (MySQL supports user-assigned only). Required for CMK and
    # for the Entra administrator.
    user_assigned_identity_ids = optional(list(string), [])

    # Customer-managed-key encryption. Key IDs are resolved Key Vault
    # data-plane key IDs; identity IDs are resolved
    # user-assigned-identity ARM IDs.
    customer_managed_key = optional(object({
      key_vault_key_id                     = string
      primary_user_assigned_identity_id    = optional(string)
      geo_backup_key_vault_key_id          = optional(string)
      geo_backup_user_assigned_identity_id = optional(string)
    }))

    # The server's single Microsoft Entra administrator. identity_id is
    # the resolved ARM ID of a user-assigned identity attached to the
    # server; object_id is the resolved directory object ID; tenant_id
    # falls back to the deploying credential's tenant when unset.
    aad_administrator = optional(object({
      identity_id = string
      login       = string
      object_id   = string
      tenant_id   = optional(string)
    }))

    # Databases to create, one Azure sub-resource each.
    databases = optional(list(object({
      name      = string
      charset   = optional(string, "utf8mb4")
      collation = optional(string, "utf8mb4_0900_ai_ci")
    })), [])

    # Public-endpoint firewall allowlist, one Azure sub-resource each.
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])

    # MySQL server parameters to override, by parameter name. Each entry
    # is applied as a user override on Azure's per-SKU default.
    server_parameters = optional(map(string), {})

    # Free-form user tags, merged over the metadata-derived tags (user
    # tags win on key collision).
    tags = optional(map(string), {})
  })
}
