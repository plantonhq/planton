# On-Demand Simple Table

This preset creates a DynamoDB table with on-demand (pay-per-request) billing and a simple partition key. On-demand pricing automatically scales to handle any traffic level without capacity planning. Point-in-time recovery and deletion protection are enabled for production safety. This is the 30-second default for most DynamoDB use cases.

## When to Use

- Key-value stores, session stores, user profiles, or any workload with a simple primary key
- Applications with unpredictable or variable traffic where capacity planning is impractical
- New tables where traffic patterns are not yet established

## Key Configuration Choices

- **On-demand billing** (`billingMode: PAY_PER_REQUEST`) -- No capacity planning; pay only for reads/writes consumed; scales instantly
- **String partition key** (`keySchema: [{attributeName: id, keyType: HASH}]`) -- Simple HASH key on `id`; sufficient for most key-value access patterns
- **Point-in-time recovery** (`pointInTimeRecoveryEnabled: true`) -- Continuous backups enabling restoration to any point in the last 35 days
- **Deletion protection** (`deletionProtectionEnabled: true`) -- Prevents accidental table deletion

## Placeholders to Replace

This preset has no placeholders. The table uses on-demand billing and a generic `id` partition key. Rename the `id` attribute and adjust the type (`S` for string, `N` for number, `B` for binary) based on your data model.

## Related Presets

- **02-provisioned-production** -- Use instead for predictable workloads where provisioned capacity is more cost-effective
