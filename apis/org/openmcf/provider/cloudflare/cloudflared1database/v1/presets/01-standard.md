# Standard D1 Database

Creates a Cloudflare D1 serverless SQLite database with optional region and read replication. Use for edge databases, Workers backends, or lightweight relational storage. Schema is managed via Wrangler migrations, not at the resource level.

## When to Use

- Serverless SQLite database for Workers or edge apps
- Key-value or relational data with SQL at the edge
- Low-latency reads when region matches your traffic

## Key Configuration Choices

- **region** (`region: wnam`) -- Primary location; values: weur, eeur, apac, oc, wnam, enam. Omit for Cloudflare default.
- **readReplication** (`readReplication.mode`) -- Set to `auto` for global read replicas; `disabled` for single-region. Enabling requires D1 Sessions API in your app.
- **databaseName** (`databaseName`) -- Unique name within account; max 64 chars.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |
| `<database-name>` | Unique name for the D1 database | Choose a descriptive name (e.g., app-cache, user-sessions) |
