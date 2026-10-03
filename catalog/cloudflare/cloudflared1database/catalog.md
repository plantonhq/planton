# Cloudflare D1 Database

Deploys a Cloudflare D1 serverless SQLite database with configurable region placement and optional read replication. The database is the container a Worker queries through a `d1` binding; schema (tables, indexes, migrations) is managed by the application via Wrangler, not at this layer. Placement is a creation-time decision: changing the region hint or jurisdiction later replaces the database and destroys its data.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **D1 Database** -- a serverless SQLite database created in the specified Cloudflare account, with an optional primary location hint (or data-residency jurisdiction) fixing where the primary instance lives
- **Read Replication** -- configured only when `readReplication` is set; enables D1 Read Replication to place read-only replicas across multiple regions for lower global read latency

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has D1 permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare account** with D1 access enabled. The `accountId` field identifies which Cloudflare account owns the database.
- **Schema management** -- D1 tables, indexes, and migrations are managed via the Wrangler CLI, not at the resource level. Deploy the database first, then run migrations against the `database_id` it outputs.

## Deploy

### Console

Open the deployment store, find **Cloudflare D1 Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard D1 Database** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareD1Database
metadata:
  name: app-cache
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  databaseName: app-cache
```

```shell
planton apply -f cloudflare-d1-database.yaml
```

This creates a D1 database named `app-cache` with Cloudflare selecting the default storage region. No read replication is configured. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a D1 database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Placement is fixed at creation** -- both `region` and `jurisdiction` are creation-time decisions; changing either replaces the database, which destroys its data. Pick placement before the database holds anything real, and treat a placement change in a plan as the destructive event it is.

**Region (`region`)** -- the primary location hint: `weur` (Western Europe), `eeur` (Eastern Europe), `apac` (Asia Pacific), `oc` (Oceania), `wnam` (Western North America), or `enam` (Eastern North America). Omit to let Cloudflare choose. Place the primary close to the Workers that write to it -- reads can be replicated later, writes cannot.

**Jurisdiction (`jurisdiction`)** -- a data-residency constraint: `eu` or `fedramp`. Mutually exclusive with `region` -- both answer the same "where does the primary live" question, so the spec rejects manifests that set both. If a residency regime applies, jurisdiction wins and the exact location belongs to Cloudflare within that boundary.

**Read replication (`readReplication.mode`)** -- `auto` places read-only replicas across regions for lower global read latency, but it changes the Worker's contract: a Worker reading a replicated database must use the D1 Sessions API for sequential consistency. Enable it together with the application change, not ahead of it. Set `disabled` or omit for a single primary.

**Database naming (`databaseName`)** -- unique within the account, 64 characters max. It appears in Wrangler commands and Worker binding configuration, so choose a name you want to type.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- the Cloudflare account is identified by the `accountId` string.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_id` | The unique identifier (UUID) of the created D1 database | A Worker's `d1` binding references this value; Wrangler migrations target it |
| `database_name` | The database name as confirmed by Cloudflare | Wrangler CLI commands and application configuration that address the database by name |
| `version` | The D1 engine version Cloudflare reports (server-assigned) | Informational; not used for wiring |

## Common Patterns

**Standard database** -- a single-primary D1 database with an optional region hint, backing a Worker or lightweight relational workload. Start from the **Standard D1 Database** preset.

**Residency-pinned database** -- set `jurisdiction: eu` (or `fedramp`) instead of a region when compliance dictates where data lives; leave the exact placement to Cloudflare inside that boundary.

**Globally replicated reads** -- set `readReplication.mode: auto` for read-heavy, globally distributed Workers -- and ship the D1 Sessions API change in the Worker in the same release, because replication without it breaks read consistency assumptions.

## Works With

- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- the primary consumer; a Worker's `d1` binding references the `database_id` output
