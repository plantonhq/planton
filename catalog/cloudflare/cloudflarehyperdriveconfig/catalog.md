# Cloudflare Hyperdrive Config

Deploys a Cloudflare Hyperdrive -- a connection pooler and global query cache that lets a Worker reach a regional SQL database (PostgreSQL or MySQL) with low latency. A Worker binds to this config and queries the origin without paying the full connection-setup round trip on every request; Hyperdrive reuses warm, pooled connections and can cache read-query results at the edge. Hyperdrive configs are account-scoped, and creation is a live connection test: Cloudflare dials the origin when the config is created, so an unreachable host or wrong credentials fail the deploy, not the first query.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Hyperdrive Config** -- an account-scoped configuration pointing at your origin database, with pooling and caching behavior; Cloudflare validates connectivity to the origin at create time
- **Origin Credentials** -- the database password (and optional Cloudflare Access service-token secret) resolved just-in-time from managed secrets at deploy, never stored in plaintext

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Hyperdrive edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Managed secret for the database password** -- store the origin password as an org secret and reference it; Hyperdrive never accepts a plaintext password.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Cloudflare Account

- **Network reachability** -- the origin must be reachable from Cloudflare's network before this config is created: a public endpoint, a private one fronted by Cloudflare Access (use the Access client ID + secret), or a Workers VPC Service (`origin.serviceId`). A create failure here is a connectivity report, not a module defect.

## Deploy

### Console

Open the deployment store, find **Cloudflare Hyperdrive Config**, and click **Deploy**. The creation wizard captures the owning account and name, the required origin connection (engine, host, port, database, user, and the reference-only password secret), and optional caching, tuning, and mTLS settings. Start from the **Basic PostgreSQL Hyperdrive** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareHyperdriveConfig
metadata:
  name: prod-postgres
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: prod-postgres
  origin:
    scheme: postgres
    host: db.example.com
    database: app_production
    user: hyperdrive
    password:
      value: "$secret/prod-db-password"
  caching:
    maxAge: 60
```

```shell
planton apply -f cloudflare-hyperdrive-config.yaml
```

This creates a Hyperdrive config fronting a PostgreSQL origin with 60-second result caching. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Hyperdrive config. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Database Engine (`origin.scheme`)** -- `postgres`/`postgresql` (default port 5432) or `mysql` (default port 3306). Determines the wire protocol Hyperdrive uses.

**Password (`origin.password`)** -- the database user's password, provided as a managed-secret reference; the runner resolves it just-in-time at deploy. The password is write-only on Cloudflare's side: the API never returns it, so an imported config shows a password diff on its first plan -- that re-assert is expected, not drift.

**Reaching a private origin -- Access or VPC Service** -- set `origin.accessClientId`/`origin.accessClientSecret` when the origin is published behind Cloudflare Access, or `origin.serviceId` to egress over a Workers VPC Service instead of dialing a public host. A VPC Service origin manages TLS on the VPC side, so the spec forbids combining `serviceId` with the `mtls` block -- pick one trust path.

**Caching (`caching`)** -- `disabled`, `maxAge`, and `staleWhileRevalidate` control edge caching of read-query results; enabled by default (60s / 15s). It is the single biggest latency win for read-heavy workloads -- and wrong for anything that must read its own writes. Disable caching for those configs rather than tuning the windows down.

**Origin Connection Limit (`originConnectionLimit`)** -- caps pooled connections to the origin. Floors at 5, ceilings by plan (about 20 on free, up to 100 on paid); 0 takes the plan default. Set it only when the origin database's own `max_connections` budget demands it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no Infra Component foreign-key dependencies; it points directly at an external origin database, and its credentials are managed-secret references resolved at deploy.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `hyperdrive_id` | The Cloudflare-assigned identifier of the config | Referenced by a CloudflareWorker's `hyperdrive` binding |

## Common Patterns

**Serverless Postgres access** -- a Hyperdrive config fronting a managed PostgreSQL database (RDS, Neon, Supabase, Cloud SQL) so Workers query it without per-request connection overhead. Start from the **Basic PostgreSQL Hyperdrive** preset.

**Private origin via Access** -- a database published behind Cloudflare Access, reached using an Access service token (client ID + secret) rather than a public endpoint.

**Private origin via VPC Service** -- set `origin.serviceId` to route through a Workers VPC Service for private connectivity with TLS managed on the VPC side. Start from the **PostgreSQL Hyperdrive over a Workers VPC Service** preset.

**Strict TLS** -- an mTLS configuration with `sslmode: verify-full` for origins that require client certificates and full hostname verification. Start from the **PostgreSQL Hyperdrive with mTLS** preset.

## Works With

- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- binds this config (a `hyperdrive` binding) to query the origin database with pooled, cached connections
