# Deploying Cloudflare D1: The SQLite Database That Lives at the Edge

## Introduction

For years, the conventional wisdom was simple: serverless functions are stateless. If you need a database, you reach for a traditional connection-pool-based RDBMS like Postgres or MySQL, accept the cold-start penalty of establishing connections from ephemeral compute, and architect around that limitation. The serverless database services that emerged—Aurora Serverless, Neon, PlanetScale—tried to bridge the gap by making connection management faster or serverless-friendly, but the fundamental model remained the same: connection strings, TCP handshakes, and compute separated from storage.

**Cloudflare D1** takes a different approach entirely. It's not a connection-based database pretending to be serverless. It's a genuinely serverless database built on SQLite, designed from the ground up for edge compute. Instead of connection strings, you get bindings. Instead of a single massive database, you design for thousands of smaller, per-tenant databases. Instead of provisioned compute, you get a pay-per-query pricing model with no idle costs. And instead of complex performance tiers (gp3 vs io2, anyone?), you get a straightforward service with generous free tier limits and simple pricing.

But this paradigm shift comes with architectural implications that are easy to miss. D1 is not a drop-in replacement for a traditional RDBMS. It's SQLite, which means no complex transactions (BEGIN/COMMIT), no multi-attach scenarios, and a design philosophy that favors horizontal scaling over vertical. More critically, the tooling for managing D1 reveals a fundamental gap: while Infrastructure-as-Code tools like Terraform and Pulumi can provision the database *resource*, they cannot manage *schema*. That responsibility falls exclusively to Cloudflare's Wrangler CLI and its migration workflow—a design choice that forces production teams into a hybrid orchestration model.

This guide explains how to deploy and manage Cloudflare D1 databases, from manual console experimentation to production-grade CI/CD pipelines. It explores the maturity spectrum of deployment methods, the critical "Orchestration Gap" between IaC and schema management, and the architectural realities that make D1 a specialized tool for edge-first applications—not a general-purpose database.

---

## The Deployment Spectrum: From ClickOps to Production

Not all approaches to managing D1 are created equal. Here's the landscape, from what to avoid to what actually works at scale:

### Level 0: The Dashboard (Anti-Pattern for Production)

**What it is:** Using the Cloudflare web dashboard to manually create a database, specify a name and location hint, and execute SQL queries in the "Console" tab.

**What it solves:** Immediate feedback for learning D1's data model or debugging a specific query. It's fine for initial experimentation or exploring how D1 handles SQLite syntax.

**What it doesn't solve:** Reproducibility, version control, automation, or team collaboration. Manual provisioning is not auditable. If you create a database with a few clicks and populate it with ad-hoc SQL, no one else on your team knows how to recreate it. When that database gets deleted or corrupted, you're starting from scratch.

**Verdict:** Use it to explore D1's interface and understand the workflow. Never for staging or production. Not even for persistent dev environments.

---

### Level 1: The Wrangler CLI (Essential, But Not Sufficient Alone)

**What it is:** Using Cloudflare's official `wrangler` CLI to manage D1 imperatively:

```bash
# Create a database
npx wrangler d1 create my-app-db --location=weur

# Execute SQL
npx wrangler d1 execute my-app-db --remote --command="CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)"

# Create and apply migrations
npx wrangler d1 migrations create my-app-db create_users_table
npx wrangler d1 migrations apply my-app-db --remote
```

**What it solves:** Automation and version control for schema. Wrangler's migration workflow—`d1 migrations create`, `d1 migrations apply`—is the **only** Cloudflare-supported method for managing database schema. It tracks applied migrations in an internal `d1_migrations` table, applies them transactionally, and supports local development via Miniflare (a local D1 runtime).

This is critical: the Wrangler CLI is not just a convenience layer over the API. It implements logic that the API does not expose. There is no REST endpoint for "apply migrations." The CLI reads local `.sql` files, queries the database's migration history, and executes unapplied changes in order. It's a complex, client-side orchestrator.

**What it doesn't solve:** Declarative infrastructure provisioning. Wrangler is imperative. Running `wrangler d1 create` twice with the same name will fail because the database already exists. Deleting a database requires manual cleanup. There's no state file tracking what you created or what changed.

More importantly, Wrangler alone doesn't scale to multi-environment patterns (dev/preview/prod) without external orchestration. You need a way to conditionally create multiple databases, bind them to different Workers, and manage lifecycle across environments.

**Verdict:** Essential for schema management and local development. But insufficient as the sole production tool. It must be paired with an IaC layer for resource provisioning.

---

### Level 2: The Cloudflare REST API (The Fragile Foundation)

**What it is:** Calling Cloudflare's v4 REST API directly to create, delete, and query D1 databases:

```bash
# Create a database
curl -X POST "https://api.cloudflare.com/client/v4/accounts/{account_id}/d1/database" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"name":"my-db","primary_location_hint":"weur"}'

# Query a database
curl -X POST "https://api.cloudflare.com/client/v4/accounts/{account_id}/d1/database/{database_id}/query" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"sql":"SELECT * FROM users WHERE id = ?"}'
```

**What it solves:** Programmatic control. The API exposes endpoints for creating, listing, deleting, and querying databases. It's what every other tool—Wrangler, Terraform, Pulumi—ultimately calls under the hood.

**What it doesn't solve:** The migration gap. The API can create the database "container," but it has **no endpoints for schema management**. There's no `/migrations/apply` or `/migrations/list`. The migration workflow exists only in the Wrangler CLI's client-side logic.

This architectural gap is critical for anyone building custom controllers (like OpenMCF). You cannot "just" use the REST API to fully provision a D1 database. You must either replicate Wrangler's migration logic (reading `.sql` files, tracking applied migrations, handling rollback) or shell out to the Wrangler binary.

**Verdict:** Useful for understanding what Terraform/Pulumi are doing under the hood, or for building custom tooling that orchestrates Wrangler. But not suitable as a standalone production tool.

---

### Level 3: Infrastructure-as-Code (Production-Ready for Resources, But Not Schema)

**What it is:** Using Terraform or Pulumi with the official Cloudflare provider to declaratively define D1 databases.

**Terraform example:**

```hcl
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "cloudflare_d1_database" "prod_db" {
  account_id           = var.account_id
  name                 = "my-app-production-db"
  primary_location_hint = "wnam"

  read_replication {
    mode = "auto"
  }
}

# Bind to Worker
resource "cloudflare_workers_script" "app" {
  name    = "my-app"
  content = file("worker.js")

  d1_database_binding {
    name        = "DB"
    database_id = cloudflare_d1_database.prod_db.id
  }
}
```

**Pulumi example (Go):**

```go
import (
    "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

db, _ := cloudflare.NewD1Database(ctx, "prod-db", &cloudflare.D1DatabaseArgs{
    AccountId:           pulumi.String(accountId),
    Name:                pulumi.String("my-app-production-db"),
    PrimaryLocationHint: pulumi.String("wnam"),
})
```

**What it solves:** Declarative, version-controlled, idempotent provisioning of the database resource. Terraform and Pulumi track state, show diffs before applying changes, and make updates predictable. They support multi-environment patterns: create separate `my-app-prod-db` and `my-app-preview-db` resources, bind them to different Workers via environment-specific configurations, and manage lifecycle independently.

Both tools are production-ready. The Cloudflare provider is officially maintained, covers all essential D1 operations (create, delete, update name/region, configure read replication), and handles Worker bindings.

**What it doesn't solve:** The "Orchestration Gap." IaC provisions the database **container**, but it does **not** create tables or manage schema. This is the single most common point of confusion. Users expect `terraform apply` to create a database with tables and data. It doesn't. You get an empty database with no schema.

The reason: Cloudflare does not expose a migration API. The only way to apply schema is via `wrangler d1 migrations apply`. IaC tools can provision the database, but they cannot manage what's *inside* it.

This forces a hybrid workflow: IaC for resources, Wrangler for schema. Some Pulumi users work around this by using a "dynamic provider" to shell out to Wrangler as part of a Pulumi deployment. This works, but it's a workaround acknowledging the gap, not a feature.

**Verdict:** Essential for production. Terraform and Pulumi make multi-environment, team-based D1 provisioning predictable and reproducible. But they must be paired with Wrangler for schema. Pure IaC is impossible.

---

### Level 4: Kubernetes-Native (Crossplane) — The Missing Integration

**What it is:** Using Crossplane to manage D1 databases as Kubernetes-native resources.

**What it should be:** A Crossplane provider that offers a `D1Database` managed resource, allowing teams to provision D1 via Kubernetes YAML and kubectl.

**What it actually is:** A gap. The official `crossplane-contrib/provider-cloudflare` (based on the cloudflare-go SDK) does **not** include a managed resource for D1. It focuses on Zones, DNS Records, and Routes—older Cloudflare primitives.

**Workaround:** Teams can use `provider-terraform` to wrap the Terraform `cloudflare_d1_database` resource. This creates a fragile, indirect management path: Crossplane → provider-terraform → Terraform HCL → Cloudflare API. It's technically possible, but it adds layers of indirection and debugging complexity.

**Verdict:** Not a first-class citizen in the Kubernetes ecosystem. If you need Kubernetes-native D1 management, you'll need a custom controller—exactly what OpenMCF provides.

---

## IaC Tool Comparison: Terraform vs. Pulumi

Both Terraform and Pulumi support D1 in production, with identical resource coverage (since Pulumi's provider bridges Terraform's). Here's how they compare:

### Terraform: The Standard Choice

**Strengths:**
- **Maturity:** Terraform is the industry-standard IaC tool. Broad ecosystem, large community, extensive documentation.
- **Familiarity:** Most ops teams already know HCL. Onboarding is straightforward.
- **Resource Coverage:** The official Cloudflare provider includes `cloudflare_d1_database`, supports the new read replication feature (via `read_replication { mode = "auto" }`), and handles Worker bindings via `cloudflare_workers_script`.
- **State Management:** Mature backends (S3, Terraform Cloud, Azure Blob) make multi-team collaboration predictable.

**Limitations:**
- **HCL Expressiveness:** HCL is declarative, but not a full programming language. Complex logic (dynamic resource generation, conditional creation of 50 per-tenant databases) can be verbose or require workarounds (count, for_each, external data sources).
- **Schema Gap:** Like all IaC tools, Terraform cannot manage schema. You must chain `terraform apply` with `wrangler d1 migrations apply` in CI/CD.

**Verdict:** The default choice for teams prioritizing stability, ecosystem maturity, and familiarity. If you're already using Terraform, it's the natural fit.

---

### Pulumi: The Programmer's IaC

**Strengths:**
- **Real Programming Languages:** Write infrastructure in TypeScript, Python, Go, or C#. Use loops, conditionals, and functions to build dynamic configurations.
- **Expressive Logic:** Easier to implement complex patterns (e.g., "create N databases based on a list of tenants") without HCL gymnastics.
- **Native Testing:** Unit test infrastructure code with Jest, pytest, or Go's testing package.
- **Equivalent Coverage:** Pulumi's Cloudflare provider (bridged from Terraform) supports the same D1 resources and features.

**Limitations:**
- **Smaller Community:** Less adoption than Terraform. Fewer public examples and community modules.
- **Runtime Dependency:** Requires Node.js, Python runtime, etc. Slightly more setup overhead than Terraform's single binary.
- **Bridged Provider:** Since Pulumi's Cloudflare provider is bridged from Terraform's, any quirks or bugs in the Terraform provider carry over.
- **Schema Gap:** Same as Terraform. Pulumi can provision the database but not the schema. Some users work around this with dynamic providers that shell out to Wrangler.

**Verdict:** Excellent if your team prefers coding infrastructure in familiar languages or needs complex orchestration logic. Slightly more overhead for simple use cases, but more expressive for complex ones.

---

### Which Should You Choose?

- **Default to Terraform** if you want the most mature, widely adopted solution with straightforward HCL configs.
- **Choose Pulumi** if you prefer writing infrastructure in TypeScript/Python/Go or need advanced logic (dynamic per-tenant database creation, complex conditionals).
- **Both work equally well** for standard D1 provisioning. The choice is more about team preference and existing tooling than capability.
- **Neither solves schema management alone.** Both require Wrangler for migrations. Plan for hybrid CI/CD workflows.

---

## The Multi-Environment Pattern: Preview Is Not a Database Property

One of the most common misconceptions about D1 is that "preview databases" are a feature you enable on a database resource. They are not.

### The Architectural Reality

A "preview database" is a **pattern**, not a property. The `preview_database_id` field found in `wrangler.toml` and Terraform's `cloudflare_workers_script` resource is a **binding-level configuration**, not a database-level one. It tells a Worker: "When you are running in a 'preview' context (e.g., a Git branch preview deployment), bind the `env.DB` variable to this *other* database ID."

Here's the correct multi-environment architecture:

1. **Provision Multiple Databases (IaC):** Use Terraform or Pulumi to create two distinct `cloudflare_d1_database` resources:
   - `my-app-prod` (database ID: `aaaa-bbbb`)
   - `my-app-preview` (database ID: `cccc-dddd`)

2. **Configure Worker Bindings (IaC or wrangler.toml):**
   - The **default** environment's `d1_database_binding` binds `env.DB` to the production database ID (`aaaa-bbbb`).
   - The `preview_database_id` field is set to the preview database ID (`cccc-dddd`).

3. **Deploy Context Determines Binding:**
   - When you run `wrangler deploy` (production), the Worker binds `env.DB` to `aaaa-bbbb`.
   - When you run `wrangler deploy --preview` (or a Git-based preview deployment), the Worker binds `env.DB` to `cccc-dddd`.

**Key Insight:** "Preview" is an attribute of the **binding**, not the **database**. The database is just a database. The Worker's deployment context (production vs. preview) determines which database it connects to.

This means:
- ❌ There is no `preview_branch = true` flag on a database resource.
- ✅ You create separate databases for each environment and configure Worker bindings to point to them based on deployment context.

---

## Production Essentials: Replication, Backups, and Anti-Patterns

### Global Distribution: Read Replication (Beta)

Historically, D1 was a single-region database. You specified a `primary_location_hint` (e.g., `weur`, `wnam`, `apac`) at creation time, and all reads and writes went to that primary location.

The new **D1 Read Replication** (currently in beta) aims to lower global read latency by creating read-only replicas in multiple regions. Enable it in IaC via:

```hcl
resource "cloudflare_d1_database" "global_db" {
  account_id           = var.account_id
  name                 = "my-app-global-db"
  primary_location_hint = "enam"

  read_replication {
    mode = "auto"
  }
}
```

**Critical Warning:** Enabling read replication is **not transparent**. D1 operates on a *sequential consistency* model. To use replicas safely, you **must** refactor your Worker code to use the **D1 Sessions API**. This API passes a "bookmark" (a token representing a point-in-time) between queries to ensure that a user's session reads from a replica that is at least as up-to-date as their last write.

Cloudflare's documentation explicitly warns: failing to use the Sessions API while replication is enabled "will compromise the consistency model provided by your application" and "cause your application to return incorrect results."

**Verdict:** Read replication is powerful for global applications, but it's a two-step deployment: an infrastructure change (enabling replication) and a mandatory application-level code change (using Sessions API). Plan accordingly.

---

### Backup and Restore: D1 Time Travel

D1's backup story is strong. **D1 Time Travel** provides automatic, always-on Point-in-Time Recovery (PITR) at no additional cost:

- **Retention:** 30 days for paid Workers plans, 7 days for free plans.
- **Mechanism:** Uses an internal "bookmark" system, not user-facing snapshots.
- **Restoration:** Destructive, in-place operation via Wrangler CLI:
  ```bash
  npx wrangler d1 time-travel restore my-app-db --timestamp=2025-11-01T12:00:00Z
  ```

**Key Points:**
- Time Travel is enabled by default. You don't configure it; it's just there.
- Restoration is **destructive**. It rewinds the database to the specified timestamp, discarding all changes after that point.
- No granular snapshots. You can't "create a snapshot and restore from it." You specify a timestamp, and D1 rewinds to that moment.

**Verdict:** Robust for disaster recovery and "undo" scenarios. But not a replacement for application-level exports if you need portable backups or cross-account restores.

---

### Monitoring and Observability

D1's observability is high-level, focused on metrics rather than granular query logs:

- **Dashboard Metrics:** The Cloudflare dashboard provides charts for query volume, query latency (p50, p90, p95), and storage size.
- **GraphQL Analytics API:** The same metrics are available programmatically via Cloudflare's GraphQL Analytics API, using the `d1AnalyticsAdaptiveGroups` dataset. This allows integration with custom dashboards (Grafana, Datadog, etc.).
- **Experimental Query Insights:** An experimental `wrangler d1 insights <database_name>` command suggests future support for granular query performance analysis.
- **No Raw Query Logs:** D1 does not provide a raw SQL query log (like Postgres's `pg_stat_statements` or MySQL's slow query log). Debugging must rely on application-level logging within the Worker or high-level metrics.

**Best Practice:** Log query execution times and errors in your Worker code to supplement D1's metrics. Use `console.log()` or structured logging to track query patterns and identify slow queries.

---

### Common Anti-Patterns to Avoid

1. **The "D1 is Postgres" Anti-Pattern:**  
   Treating D1 as a general-purpose, connection-based RDBMS. It's not. It's SQLite, designed for read-heavy, edge-first applications. Write-intensive, complex transactional workloads (full ACID with BEGIN/COMMIT) are not D1's strength.

2. **The "Pure IaC" Anti-Pattern:**  
   Expecting `terraform apply` or `pulumi up` to create tables and populate schema. It won't. IaC provisions the database container. Schema comes from Wrangler migrations.

3. **The "Dynamic Per-Tenant" Anti-Pattern:**  
   Attempting to build a workflow where a Worker dynamically creates D1 databases on-demand (e.g., "new user signs up → Worker creates their personal D1 database"). The tooling doesn't support this. Worker bindings are statically defined at deployment time. The "per-tenant" model Cloudflare describes is a *manual, pre-provisioned sharding strategy* that must be orchestrated externally (by IaC or a custom controller), not a dynamic feature of D1.

4. **The "Transparent Replication" Anti-Pattern:**  
   Enabling read replication without refactoring Worker code to use the D1 Sessions API. This **will** cause data consistency errors and incorrect query results.

5. **The "Ignoring Indexes" Anti-Pattern:**  
   Failing to create indexes for common query patterns. On a "pay-per-row-read" pricing model, a query that scans 1.4 million rows instead of 417 (due to a missing index) is not just slow—it's expensive and avoidable.

---

## The 80/20 Configuration: What You Actually Need

The research reveals a clear 80/20 split: most users need just a few essential fields. Advanced features happen at the application layer or via Wrangler migrations, not at the database resource level.

### Essential Fields (The 80%)

1. **`account_id` (string):** Required. The Cloudflare account ID. This is a path parameter in the API and a required field in IaC.

2. **`database_name` (string):** Required. The human-readable name for the database. Must be unique within the account.

3. **`region` (string):** Optional. Maps to the `primary_location_hint` property in the API. Specifies the geographical region for the database's primary (write) instance. Valid values:
   - `weur` (Western Europe)
   - `eeur` (Eastern Europe)
   - `apac` (Asia Pacific)
   - `oc` (Oceania)
   - `wnam` (Western North America)
   - `enam` (Eastern North America)

   If omitted, Cloudflare selects a default location based on your account settings.

### Production Optional (The 20%)

4. **`read_replication` (object):** Optional. Configures D1 Read Replication (Beta). Based on IaC providers, this object contains a `mode` field:
   - `"auto"`: Enable automatic read replication
   - `"disabled"`: Disable replication (default)

   **Warning:** Enabling this requires application-level code changes to use the D1 Sessions API.

### Fields to Explicitly **Exclude**

5. **`preview_branch` (boolean):** ❌ Architecturally incorrect. Preview environments are a **pattern** implemented by creating a second database and using Worker bindings (`preview_database_id`) to point to it. This is not a property of the database itself.

6. **`primary_key` (string):** ❌ Fundamentally wrong. A primary key is a **schema-level** construct (part of a `CREATE TABLE` statement). The D1 database resource API (and IaC providers) only manage the database "container." Schema (tables, columns, indexes, keys) is managed exclusively via `wrangler d1 migrations`.

---

## Configuration Examples: Dev, Preview, Production

### Development: Minimal Database

**Use Case:** Developer's local sandbox. Small database for testing.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: dev-db
spec:
  account_id: "abc123..."
  database_name: "my-app-dev-db"
  region: "weur"
```

**Rationale:**
- Minimal config: just account, name, and region.
- Western Europe for lowest latency to EU developers.
- No replication (unnecessary for dev).
- Schema managed via `wrangler d1 migrations apply my-app-dev-db --local` during local development.

---

### Preview: Separate Database for Staging

**Use Case:** Preview environment for testing changes before production.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: preview-db
spec:
  account_id: "abc123..."
  database_name: "my-app-preview-db"
  region: "wnam"
```

**Worker Binding (wrangler.toml):**

```toml
# Production binding
[[d1_databases]]
binding = "DB"
database_name = "my-app-prod-db"
database_id = "aaaa-bbbb-..."

# Preview binding (points to separate database)
preview_database_id = "cccc-dddd-..."
```

**Rationale:**
- Separate database resource (`my-app-preview-db`).
- Worker binding configured with `preview_database_id` pointing to this database.
- When deploying preview branches (`wrangler deploy --preview`), Worker connects to the preview database.
- Schema applied via CI/CD: `wrangler d1 migrations apply my-app-preview-db --remote`.

---

### Production: Global Database with Replication

**Use Case:** Production application serving global users.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareD1Database
metadata:
  name: prod-db
spec:
  account_id: "abc123..."
  database_name: "my-app-production-db"
  region: "enam"
  read_replication:
    mode: "auto"
```

**Rationale:**
- Primary region: Eastern North America (`enam`).
- Read replication enabled (`mode: auto`) for global low-latency reads.
- Worker code **must** use D1 Sessions API to maintain consistency.
- Schema applied via CI/CD: `wrangler d1 migrations apply my-app-production-db --remote`.
- Backup strategy: Rely on D1 Time Travel (30-day PITR) for disaster recovery.

---

## The Orchestration Gap: Solving Production Deployment

The critical insight for production D1 deployment is this: **provisioning a D1 database is a multi-step, multi-tool process**. A pure IaC workflow is impossible. You must orchestrate IaC (for the database resource) and Wrangler CLI (for schema migrations).

### The Hybrid CI/CD Pattern

The correct, idempotent production workflow is:

1. **Trigger:** Git push to main (or feature branch for preview).

2. **Step 1: Provision Database Resource (IaC)**
   ```bash
   # Using Terraform
   terraform apply -auto-approve

   # Or using OpenMCF
   planton apply
   ```
   This ensures the `cloudflare_d1_database` resource exists. It creates or updates the database "container" and outputs the `database_name` for the next step.

3. **Step 2: Apply Schema Migrations (Wrangler CLI)**
   ```bash
   npx wrangler d1 migrations apply my-app-production-db --remote
   ```
   Using the `database_name` from Step 1, this connects to the provisioned database and applies any new `.sql` migration files found in the repository's `migrations/` directory.

4. **Step 3: Deploy Application (Wrangler or IaC)**
   ```bash
   npx wrangler deploy
   ```
   This deploys the Worker code, which is now compatible with the schema applied in Step 2.

### Why This Matters

- **Declarative + Imperative:** IaC (declarative) for resources, Wrangler (imperative but version-controlled via `.sql` files) for schema.
- **Idempotent:** Running this pipeline twice produces the same result. IaC no-ops if the database exists. Wrangler only applies unapplied migrations.
- **Multi-Environment:** The same pattern works for dev, preview, and prod—just with different database names.
- **No Manual Steps:** Everything is automated. No SSH into servers, no manual SQL execution.

---

## OpenMCF's Approach: Bridging the Gap

OpenMCF provides a clean, protobuf-defined API for Cloudflare D1 that abstracts the orchestration complexity while respecting D1's architectural realities.

### What We Abstract

**The `CloudflareD1DatabaseSpec` includes:**

```protobuf
message CloudflareD1DatabaseSpec {
  // (Required) The Cloudflare Account ID.
  string account_id = 1;

  // (Required) The name for the D1 database.
  string database_name = 2;

  // (Optional) Location hint for the primary database.
  // Valid values: "weur", "eeur", "apac", "oc", "wnam", "enam".
  // If omitted, Cloudflare selects a default location.
  string region = 3;

  // (Optional) Configures D1 Read Replication (Beta).
  // If omitted, replication is disabled.
  ReplicationConfig replication_config = 4;
}

message ReplicationConfig {
  // (Required if replication_config is set) The replication mode.
  // "auto": Enable automatic read replication.
  // "disabled": Disable read replication.
  string mode = 1;
}
```

This follows the **80/20 principle**: 80% of users need only `account_id`, `database_name`, and optionally `region`. Advanced scenarios (read replication) are supported but optional.

### What We **Don't** Include

- **`preview_branch`:** Architecturally incorrect. Previews are a binding-level pattern, not a database property.
- **`primary_key`:** Schema-level construct. Managed via Wrangler migrations, not at the database resource level.

### Default Choices

- **Region:** If omitted, Cloudflare selects a default. Users can specify based on primary user geography.
- **Read Replication:** Disabled by default. Users enable it explicitly when they're ready to refactor their Worker code to use the Sessions API.

### Under the Hood: Pulumi (Go)

OpenMCF uses **Pulumi (Go)** for D1 provisioning. Why?

- **Language Consistency:** Pulumi's Go SDK integrates naturally with OpenMCF's broader multi-cloud orchestration (also Go-based).
- **Equivalent Coverage:** Pulumi's Cloudflare provider (bridged from Terraform) supports all D1 operations we need: create, delete, update, and configure read replication.
- **Future-Proofing:** Pulumi's programming model makes it easier to add conditional logic, multi-database strategies, or custom integrations (e.g., automatically pairing `wrangler d1 migrations apply` with resource provisioning).

That said, Terraform would work equally well. The protobuf API remains the same regardless of the IaC engine underneath.

---

## Key Takeaways

1. **D1 is SQLite-at-the-edge, not a traditional RDBMS.** It's designed for read-heavy, edge-first applications with horizontal scaling across many small databases, not vertical scaling of a single massive database.

2. **Manual management (dashboard) is an anti-pattern.** Use IaC (Terraform or Pulumi) for production database provisioning. The Cloudflare provider is mature and supports all essential operations.

3. **Schema management requires Wrangler.** IaC provisions the database "container," but only `wrangler d1 migrations apply` can create tables and manage schema. This is not a bug; it's by design. Plan for hybrid CI/CD workflows: IaC → Wrangler migrations → Worker deployment.

4. **Preview databases are a pattern, not a property.** You create separate databases for dev/preview/prod and use Worker binding configuration (`preview_database_id`) to point to them based on deployment context.

5. **Read replication is powerful but not transparent.** Enabling it requires refactoring Worker code to use the D1 Sessions API. Otherwise, you'll get data consistency errors.

6. **The 80/20 config is account_id, database_name, and region.** Advanced features (replication, schema, indexes) happen via optional config or Wrangler migrations, not at the database resource level.

7. **D1 Time Travel provides robust PITR backups.** 30-day retention for paid plans, 7-day for free. Restoration is destructive and in-place, but it's automatic and free.

8. **OpenMCF abstracts the complexity** into a clean protobuf API, making multi-cloud deployments consistent while respecting D1's unique architecture. We provision the database via Pulumi, but we document and support the hybrid workflow (IaC + Wrangler migrations) that production requires.

---

## Further Reading

- **Cloudflare D1 Documentation:** [D1 Overview](https://developers.cloudflare.com/d1/)
- **Wrangler CLI Guide:** [Getting Started with D1](https://developers.cloudflare.com/d1/get-started/)
- **D1 Migrations Workflow:** [Local Development Best Practices](https://developers.cloudflare.com/d1/best-practices/local-development/)
- **Terraform Cloudflare Provider:** [cloudflare_d1_database Resource](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/d1_database)
- **D1 Time Travel (Backups):** [Time Travel and Backups](https://developers.cloudflare.com/d1/platform/time-travel/)
- **D1 Read Replication:** [Workers Binding API](https://developers.cloudflare.com/d1/worker-api/)

---

**Bottom Line:** Cloudflare D1 is a genuinely serverless, SQLite-based database designed for edge compute. It's not a drop-in replacement for Postgres, but for the right use case—read-heavy, edge-first applications tightly integrated with Cloudflare Workers—it's a powerful, cost-effective choice. Manage it with IaC (Terraform or Pulumi) for resource provisioning and Wrangler for schema migrations. OpenMCF simplifies this with a protobuf API that hides the orchestration complexity while exposing the essential configuration you actually need.

