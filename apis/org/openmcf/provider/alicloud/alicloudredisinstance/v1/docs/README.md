# AliCloudRedisInstance Research Documentation

## Provider Resource Analysis

### alicloud_kvstore_instance (Terraform) / kvstore.Instance (Pulumi)

The KVStore instance is the sole resource for this component. Key findings from provider analysis:

- **Dual engine**: Supports both Redis and Memcache via the `instance_type` field (default: Redis). The component is named RedisInstance because Redis represents 95%+ of KVStore usage.
- **Engine versions**: 2.8, 4.0, 5.0, 6.0, 7.0 (Redis); default to 7.0 for new deployments
- **Field naming evolution**: Several fields have been renamed in newer provider versions:
  - `instance_name` -> `db_instance_name`
  - `instance_charge_type` -> `payment_type`
  - `availability_zone` -> `zone_id`
  - `parameters` -> `config`
  - `connection_string_prefix` -> use `kvstore.Connection` resource
- **Cluster mode**: The `shard_count` field enables cluster mode for horizontal scaling
- **Connection management**: The `connection_domain` is the VPC-internal endpoint; public endpoints require a separate `alicloud_kvstore_connection` resource

### Related Resources (NOT Bundled)

| Resource | Reason Not Bundled |
|---|---|
| `alicloud_kvstore_account` | Redis ACL accounts (6.0+) -- niche feature; instance password covers 80% of use cases |
| `alicloud_kvstore_connection` | Public endpoint management -- security-sensitive, should be explicit and separate |
| `alicloud_kvstore_backup_policy` | Backup settings inline on instance (`backup_period`, `backup_time`) are sufficient |
| `alicloud_kvstore_audit_log_config` | Operational monitoring concern, not infrastructure definition |

## Design Rationale

### No Account Bundling (Unlike RDS)

RDS bundles databases + accounts because an instance without them is incomplete. Redis is different:

1. The instance `password` field handles authentication for the vast majority of use cases
2. Redis accounts (`alicloud_kvstore_account`) are a 6.0+ ACL feature for fine-grained access control
3. Most Redis deployments use a single password, not multiple per-database accounts
4. Bundling accounts would add complexity without benefiting most users

### backup_period is a String Set, Not an Integer

The initial T02 spec guidance listed `int32 backup_period`. The actual provider schema uses a set of day-name strings (e.g., "Monday", "Wednesday") plus a separate `backup_time` string for the time window. Both are exposed as separate fields.

### period Field is a String, Not an Integer

The Terraform provider defines `period` as a string with string-based validation ("1", "2", ... "9", "12", "24", "36"). The spec follows the provider's native type to avoid unnecessary type conversion.

### Fields Excluded from Spec

| Field | Reason |
|---|---|
| `backup_id`, `restore_time`, `srcdb_instance_id` | Point-in-time restore -- operational, not declarative |
| `dedicated_host_group_id` | Enterprise dedicated host -- very niche |
| `global_instance`, `global_instance_id` | Cross-region distributed cache -- advanced feature |
| `business_info`, `coupon_no`, `auto_use_coupon`, `dry_run` | Billing/operational concerns |
| `order_type`, `force_upgrade`, `effective_time` | Operational change management |
| `kms_encrypted_password`, `kms_encryption_context` | Use external secret management |
| `slave_read_only_count` | Secondary zone read replicas -- extremely advanced |
| `node_type`, `parameters`, `availability_zone`, `instance_name`, `instance_charge_type` | All deprecated |
| `enable_public`, `connection_string_prefix`, `connection_string`, `port` | Deprecated; use `kvstore.Connection` |
| `capacity`, `bandwidth` | Computed from instance_class, not user-configurable |
| `is_auto_upgrade_open` | Minor operational concern |
