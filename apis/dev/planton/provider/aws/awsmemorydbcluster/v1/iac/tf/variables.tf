variable "metadata" {
  description = "Resource metadata (name, id, org, env, labels)"
  type = object({
    name = string
    id   = string
    org  = string
    env  = string
  })
}

variable "spec" {
  description = "AwsMemorydbCluster specification"
  type = object({
    # The AWS region where the resource will be created.
    region                     = string
    engine                     = string
    engine_version             = optional(string)
    description                = optional(string)
    node_type                  = string
    port                       = optional(number, 6379)
    num_shards                 = optional(number, 1)
    num_replicas_per_shard     = optional(number, 1)
    acl_name                   = optional(string, "open-access")
    subnet_ids                 = optional(list(string), [])
    security_group_ids         = optional(list(string), [])
    tls_enabled                = optional(bool, true)
    kms_key_id                 = optional(string)
    maintenance_window         = optional(string)
    snapshot_retention_limit   = optional(number, 0)
    snapshot_window            = optional(string)
    final_snapshot_name        = optional(string)
    snapshot_arns              = optional(list(string), [])
    snapshot_name              = optional(string)
    parameter_group_family     = optional(string)
    parameters                 = optional(list(object({ name = string, value = string })), [])
    sns_topic_arn              = optional(string)
    auto_minor_version_upgrade = optional(bool, true)
    data_tiering               = optional(bool, false)
  })
}
