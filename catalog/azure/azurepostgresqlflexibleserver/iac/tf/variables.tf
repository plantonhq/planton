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
  description = "Azure Database for PostgreSQL Flexible Server specification"
  type = object({
    # The Azure region the server lives in.
    region = string

    # The resource group the server lives in. References are resolved to a
    # literal name by the platform before the module runs.
    resource_group = string

    # The server's name: 3-63 lowercase letters/digits/hyphens, globally
    # unique (it becomes {server_name}.postgres.database.azure.com).
    server_name = string

    # How the server comes into existence, as the spec enum's name string
    # (DEFAULT, POINT_IN_TIME_RESTORE, REPLICA, GEO_RESTORE,
    # REVIVE_DROPPED). Unset means DEFAULT (a fresh, empty server).
    create_mode = optional(string)

    # For the replica/restore modes: the resolved ARM ID of the source
    # server.
    source_server_id = optional(string)

    # For the restore modes: the RFC-3339 UTC instant to restore to.
    point_in_time_restore_time_in_utc = optional(string)

    # Replica promotion: "NONE" breaks replication and promotes the
    # replica to a standalone primary. Never sent at creation (Azure
    # rejects it); only meaningful as a day-2 update on a replica.
    replication_role = optional(string)

    # PostgreSQL password-auth admin credentials. Empty on an Entra-only
    # server (password auth disabled) and on replicas/restores, which
    # inherit from the source.
    administrator_login    = optional(string)
    administrator_password = optional(string)

    # The PostgreSQL major version ("11"-"18"). Only sent for a fresh
    # (DEFAULT) server -- replicas and restores inherit the source's
    # version, so sending one would fight the service.
    version = optional(string)

    # The compute SKU ({TIER}_Standard_{SIZE}). Empty on a replica
    # inherits the source's SKU.
    sku_name = optional(string)

    # The provisioned storage size in MB, from Azure's fixed ladder.
    # Unset lets Azure apply its default (32768) -- or, on a replica,
    # inherit the source's size.
    storage_mb = optional(number)

    # The storage performance tier, as the spec enum's name string
    # (P4-P80). Unset applies Azure's default tier for the size.
    storage_tier = optional(string)

    # Whether storage grows automatically (Azure's default is false).
    auto_grow_enabled = optional(bool, false)

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

    # Backup retention days, 7-35 (Azure's default is 7).
    backup_retention_days = optional(number, 7)

    # Whether backups replicate to the paired region (fixed at creation).
    geo_redundant_backup_enabled = optional(bool, false)

    # Whether the server accepts public connections (Azure's default is
    # true). Must be false when the server is VNet-injected.
    public_network_access_enabled = optional(bool, true)

    # VNet injection: the resolved ARM IDs of the delegated subnet and the
    # private DNS zone the server registers in. Set together; fixed at
    # creation.
    delegated_subnet_id = optional(string)
    private_dns_zone_id = optional(string)

    # Which authentication mechanisms the server accepts. Omitted means
    # Azure's default: password auth on, Entra auth off.
    authentication = optional(object({
      active_directory_auth_enabled = optional(bool, false)
      password_auth_enabled         = optional(bool, true)
      tenant_id                     = optional(string)
    }))

    # Microsoft Entra administrator grants, one Azure sub-resource each.
    # principal_type arrives as the spec enum's name string
    # (USER/GROUP/SERVICE_PRINCIPAL); object_id is the resolved directory
    # object ID.
    aad_administrators = optional(list(object({
      object_id      = string
      principal_name = string
      principal_type = string
    })), [])

    # The server's managed identity. type arrives as the spec enum's name
    # string (SYSTEM_ASSIGNED/USER_ASSIGNED/SYSTEM_AND_USER_ASSIGNED);
    # identity_ids are resolved user-assigned-identity ARM IDs.
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    # Customer-managed-key encryption (fixed at creation). Key IDs are
    # resolved Key Vault data-plane key IDs; identity IDs are resolved
    # user-assigned-identity ARM IDs.
    customer_managed_key = optional(object({
      key_vault_key_id                     = string
      primary_user_assigned_identity_id    = optional(string)
      geo_backup_key_vault_key_id          = optional(string)
      geo_backup_user_assigned_identity_id = optional(string)
    }))

    # Elastic cluster (PG 17+): presence provisions a sharded cluster of
    # the declared node count instead of a single node.
    cluster = optional(object({
      size                  = number
      default_database_name = optional(string)
    }))

    # Databases to create, one Azure sub-resource each.
    databases = optional(list(object({
      name      = string
      charset   = optional(string, "UTF8")
      collation = optional(string, "en_US.utf8")
    })), [])

    # Public-endpoint firewall allowlist, one Azure sub-resource each.
    firewall_rules = optional(list(object({
      name             = string
      start_ip_address = string
      end_ip_address   = string
    })), [])

    # PostgreSQL server parameters to override, by parameter name. Each
    # entry is applied as a user override on Azure's per-SKU default.
    server_parameters = optional(map(string), {})

    # Free-form user tags, merged over the metadata-derived tags (user
    # tags win on key collision).
    tags = optional(map(string), {})
  })
}
