# Provisioned Production Table

This preset creates a DynamoDB table with provisioned capacity, a composite primary key (partition + sort key), and server-side encryption with a customer-managed key. Provisioned mode is more cost-effective than on-demand for predictable, steady-state workloads. The composite key supports rich query patterns like hierarchical data and time-series lookups.

## When to Use

- Workloads with predictable, steady read/write traffic where provisioned capacity saves money vs on-demand
- Data models requiring range queries, sorting, or hierarchical access patterns (partition + sort key)
- Applications needing customer-managed encryption keys for compliance

## Key Configuration Choices

- **Provisioned billing** (`billingMode: PROVISIONED`) -- 25 RCUs and 25 WCUs as a starting point; enable auto-scaling in your application's scaling policy
- **Composite key** (`pk` HASH + `sk` RANGE) -- Enables efficient range queries within a partition (e.g., all orders for a customer sorted by date)
- **Server-side encryption** (`serverSideEncryption.enabled: true`) -- Uses a customer-managed KMS key for audit trails; omit to use AWS-owned key (free)
- **Point-in-time recovery** -- Continuous backups for the last 35 days
- **Deletion protection** -- Prevents accidental table deletion

## Placeholders to Replace

This preset has no placeholders. Adjust `pk`/`sk` names and types to match your data model, and tune `readCapacityUnits`/`writeCapacityUnits` based on expected traffic. Consider enabling DynamoDB auto-scaling for production.

## Related Presets

- **01-on-demand-simple** -- Use instead for unpredictable traffic or simple key-value access patterns
