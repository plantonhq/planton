# Deploying Cloudflare Workers KV Namespaces: Edge Storage Without the Edge Cases

## Introduction

For years, conventional wisdom said that stateful data and edge computing don't mix. You put your application logic at the edge for speed, but you kept your database centralized—accepting the latency hit as the price of consistency. Cloudflare Workers KV turned that assumption on its head by introducing a globally distributed key-value store that **embraces eventual consistency** rather than fighting it.

**Workers KV** is Cloudflare's answer to edge storage—a globally replicated data store that caches key-value pairs at every one of Cloudflare's 300+ edge locations. It's not trying to be a traditional database. It's a purpose-built caching layer that trades strong consistency for something more valuable at the edge: **sub-millisecond read latency** for hot data, anywhere in the world.

The architecture is elegant: writes go to a central store and propagate to edge caches on demand. The first read from a region (a "cold read") fetches from central storage. Subsequent reads in that region hit the local cache and return in microseconds. This tiered design means KV is **optimized for read-heavy workloads**—think feature flags, configuration, session data, or cached API responses that change infrequently but need to be accessed constantly.

What makes KV compelling isn't just performance—it's **simplicity**. There's no cluster to manage, no replication topology to configure, no consistency level to choose. You create a namespace, bind it to your Worker, and start writing keys. Cloudflare handles global distribution, caching, and resilience. The pricing model matches the simplicity: generous free tier (100K reads/day), then pay-per-operation beyond that. No provisioned capacity, no reserved instances—just usage-based billing that scales to zero.

But this simplicity comes with trade-offs. KV is **eventually consistent**, with propagation delays up to 60 seconds or more. It's not suitable for counters, queues, or anything requiring immediate read-after-write consistency. For those cases, Cloudflare offers Durable Objects (single-location strong consistency) or you look elsewhere. KV shines when you can tolerate a delay between writing a value and seeing it globally—which is true for more use cases than you might think.

This guide walks through the landscape of deploying and managing Workers KV namespaces, from quick prototypes with the Wrangler CLI to production-grade Infrastructure-as-Code with Terraform and Pulumi. We'll show how OpenMCF abstracts these choices into a clean protobuf API that captures the 80% use case while staying out of your way.

---

## Understanding Workers KV: What It Is and When to Use It

Before diving into deployment methods, it's worth understanding what Workers KV actually is—and what it isn't.

### The Data Store Spectrum: KV vs. Durable Objects vs. R2

Cloudflare offers three storage primitives for Workers, each optimized for different use cases:

**Workers KV: Globally Cached Key-Value Store**
- **Best for:** Small data (up to 25 MiB per key) that's read often and written infrequently
- **Consistency:** Eventual (60+ seconds propagation time)
- **Use cases:** Feature flags, cached API responses, session tokens, configuration data, user preferences
- **Anti-patterns:** Counters, real-time coordination, rapidly mutating data, transactional workflows

**Durable Objects: Strongly Consistent State**
- **Best for:** Coordinated state, high-write scenarios, real-time updates
- **Consistency:** Strong (within a single object)
- **Use cases:** Counters, collaborative editing, WebSocket connections, game state
- **Trade-off:** Runs in a specific region per object (not globally replicated like KV)

**R2 Storage: Large Object Storage**
- **Best for:** Large binary files, bulk data, archival storage
- **Consistency:** Strong
- **Use cases:** Images, backups, video files, large JSON blobs
- **Advantage:** S3-compatible API, no egress fees

The key insight: **KV is a distributed cache, not a database**. It's the equivalent of putting a CDN in front of your config service. You get global read speed at the cost of write propagation delay.

### Performance Characteristics: Hot vs. Cold Reads

Workers KV uses a tiered architecture that determines its performance profile:

1. **First read (cold):** Data must be fetched from central storage. Latency: tens to hundreds of milliseconds (comparable to a traditional database query).

2. **Subsequent reads (hot):** Data is cached at the edge location. Latency: sub-millisecond (comparable to in-memory cache).

3. **Write propagation:** Updates go to central storage immediately, but edge caches expire gradually over 60+ seconds.

4. **Negative caching:** Missing keys are also cached. If you write a new key after checking for it, that edge location may not see the new key for up to 60 seconds.

This means **KV is a poor fit for write-heavy workloads**. If you're updating the same key dozens of times per second, you'll hit rate limits and incur propagation delays. The sweet spot is configuration that changes hourly or daily but gets read thousands of times per second.

### Pricing and Limits: Where the Free Tier Ends

Cloudflare's KV pricing follows a usage-based model with a generous free tier:

**Free Plan:**
- 100,000 read operations/day
- 1,000 write/delete/list operations/day
- 1 GB stored data
- Limits reset daily at 00:00 UTC
- Exceeding limits results in errors (not charges)

**Paid Plans (Bundled Workers, $5/month):**
- 10 million reads/month included
- 1 million writes/month included
- Additional reads: $0.50 per million
- Additional writes: $5.00 per million
- Storage beyond 1 GB: $0.50/GB-month
- No egress charges

**Key takeaway:** Reads are cheap (10× cheaper than writes). This pricing reinforces the design: KV is for read-heavy workloads. If your write costs exceed your compute costs, you're probably using the wrong tool.

---

## The Deployment Spectrum: From Prototype to Production

Not all approaches to managing Workers KV are created equal. Here's how the methods stack up, from quick prototypes to production-ready Infrastructure-as-Code.

### Level 0: Dashboard Clicks (Prototyping Only)

**What it is:** Using the Cloudflare web dashboard to manually create KV namespaces and manage keys.

**What it solves:** Nothing that can't be solved better another way. The dashboard is fine for understanding the UI and testing a concept, but it's manual, error-prone, and leaves no audit trail.

**What it doesn't solve:** Repeatability, version control, automation. If you create a namespace in the dashboard but forget to update your Worker binding, you'll waste an hour debugging why `MY_KV` is undefined.

**Common pitfalls:**
- Creating namespaces but forgetting to bind them to Workers
- Losing track of which namespace belongs to which environment
- Accumulating orphaned namespaces that quietly bill you

**Verdict:** Use it to explore. Never use it for staging or production.

---

### Level 1: Wrangler CLI (Development Workflow)

**What it is:** Using Cloudflare's official `wrangler` CLI to create and manage KV namespaces from your terminal.

**Example workflow:**

```bash
# Create a new namespace
wrangler kv:namespace create "CONFIG_STORE"
# Output: namespace ID and wrangler.toml snippet

# Add the namespace binding to wrangler.toml
[[kv_namespaces]]
binding = "CONFIG_STORE"
id = "abc123..."

# Seed data
wrangler kv:key put --namespace-id=abc123 "feature_flags" '{"darkMode": true}'

# Deploy Worker
wrangler publish
```

**What it solves:**
- Rapid iteration during development
- Quick data inspection and debugging
- Bulk import/export of keys via JSON files
- Preview namespaces for testing without touching production

**What it doesn't solve:**
- State management (no tracking of what was created when)
- Idempotency (running the same command twice may fail or create duplicates)
- Multi-environment orchestration (you must manually track dev/staging/prod namespaces)

**Best practice:** Use Wrangler for local development and debugging. Create a separate namespace for each developer (e.g., `myapp-dev-alice`, `myapp-dev-bob`) to avoid conflicts. Use `wrangler.toml` environment sections to separate dev and prod bindings:

```toml
[[kv_namespaces]]
binding = "CONFIG"
id = "prod-namespace-id"

[env.dev]
[[env.dev.kv_namespaces]]
binding = "CONFIG"
id = "dev-namespace-id"
```

**Verdict:** Excellent for development. Not suitable for production deployments where you need state tracking and rollback capabilities.

---

### Level 2: Cloudflare REST API (Custom Integration)

**What it is:** Calling Cloudflare's HTTP API directly to create namespaces, write keys, and manage lifecycle.

**Example (creating a namespace):**

```bash
curl -X POST "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/storage/kv/namespaces" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  --data '{"title": "my-kv-namespace"}'
```

**What it solves:**
- Maximum flexibility for custom tooling
- Integration into non-IaC workflows (e.g., Ansible playbooks, custom deployment scripts)
- Direct control over API parameters

**What it doesn't solve:**
- Abstraction (you're managing HTTP calls and sequencing yourself)
- State tracking (no automatic record of what exists or what changed)
- Idempotency (you must implement "create if not exists" logic manually)

**Authentication:** The API supports two authentication methods:
1. **API Token** (recommended): Fine-grained permissions, rotatable, scoped to specific resources
2. **Global API Key** (legacy): Full account access, not recommended for security

For KV management, your API token needs the **Workers KV Storage: Edit** permission at the account level.

**Verdict:** Useful for custom integrations or if you're building your own orchestration layer. For most teams, higher-level tools (Terraform, Pulumi) provide better ergonomics.

---

### Level 3: Infrastructure-as-Code (Production Standard)

**What it is:** Using Terraform or Pulumi to declaratively define KV namespaces and their lifecycle.

**Terraform example:**

```hcl
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "cloudflare_workers_kv_namespace" "config" {
  account_id = var.cloudflare_account_id
  title      = "myapp-config-prod"
}

# Output the namespace ID for use in Worker bindings
output "config_kv_id" {
  value = cloudflare_workers_kv_namespace.config.id
}
```

**Pulumi example (Go):**

```go
kvNamespace, err := cloudflare.NewWorkersKvNamespace(ctx, "config", &cloudflare.WorkersKvNamespaceArgs{
    Title: pulumi.String("myapp-config-prod"),
})
if err != nil {
    return err
}

ctx.Export("configKvId", kvNamespace.ID())
```

**What it solves:**
- **Declarative state:** Describe what you want, not how to get there
- **Idempotency:** Running `terraform apply` or `pulumi up` multiple times produces the same result
- **Plan/preview:** See what will change before applying
- **Version control:** Track infrastructure changes in Git with diffs, reviews, and rollbacks
- **Multi-environment support:** Reuse the same config with different variable values for dev/staging/prod
- **Dependency management:** Automatically create namespaces before deploying Workers that reference them

**What it doesn't solve:**
- The underlying limitations of KV (eventual consistency, single-tier storage, no multi-attach)
- Managing the contents of KV (keys/values are typically managed by application code, not IaC)

**Terraform vs. Pulumi:**

| Aspect | Terraform | Pulumi |
|--------|-----------|--------|
| **Maturity** | Older, widely adopted, battle-tested | Newer, production-ready, growing ecosystem |
| **Language** | HCL (declarative DSL) | Real programming languages (Go, TypeScript, Python) |
| **Cloudflare Support** | Official provider, comprehensive KV resource support | Bridged from Terraform provider, equivalent coverage |
| **State Management** | Local or remote backends (S3, Terraform Cloud) | Pulumi Cloud or self-managed (S3, Azure Blob) |
| **Strengths** | Simple for standard use cases, huge community | Expressive for complex logic, native testing |
| **Limitations** | HCL less expressive than full language | Smaller community, requires runtime (Node/Python/Go) |

**Which to choose?**
- Default to **Terraform** if you want the most mature, widely-adopted solution with a large ecosystem.
- Choose **Pulumi** if your team prefers coding infrastructure in familiar languages or needs advanced orchestration (dynamic resource generation, complex conditionals).

Both are production-ready. The choice is team preference, not capability.

**Verdict:** This is the production standard. Use Wrangler for development, but always deploy to staging/prod via IaC.

---

### Level 4: Crossplane (Kubernetes-Native IaC)

**What it is:** Using Crossplane (a Kubernetes-based IaC tool) to manage KV namespaces as Kubernetes Custom Resources.

**Example:**

```yaml
apiVersion: kv.cloudflare.crossplane.io/v1alpha1
kind: Namespace
metadata:
  name: myapp-config
spec:
  forProvider:
    title: myapp-config-prod
    accountId: abc123
```

**What it solves:**
- Unified infrastructure management for teams already using Crossplane
- GitOps workflows (declare resources in YAML, apply via CI/CD)
- Integration with Kubernetes-native tooling

**What it doesn't solve:**
- Adds complexity if you're not already invested in Crossplane
- Cloudflare provider is less mature than Terraform/Pulumi equivalents

**Verdict:** Use it if you're all-in on Crossplane for multi-cloud orchestration. Otherwise, stick with Terraform or Pulumi.

---

## Production Best Practices: Making KV Reliable

Managing Workers KV in production requires understanding its constraints and designing around them.

### Namespace Organization: Separation and Naming

**Pattern 1: Separate namespaces per environment**

```toml
# wrangler.toml
[[kv_namespaces]]
binding = "CONFIG"
id = "prod-config-id"  # Production

[env.staging]
[[env.staging.kv_namespaces]]
binding = "CONFIG"
id = "staging-config-id"  # Staging

[env.dev]
[[env.dev.kv_namespaces]]
binding = "CONFIG"
id = "dev-config-id"  # Development
```

**Rationale:** Complete isolation. No risk of dev data leaking into prod. Easy to audit and monitor per environment.

**Pattern 2: Separate namespaces per concern**

Instead of one giant namespace, create multiple namespaces for different data types:

- `myapp-config`: Feature flags and configuration
- `myapp-cache`: Cached API responses
- `myapp-session`: User session tokens

**Rationale:** Easier to apply different policies (TTLs, monitoring) and clearer separation of concerns.

**Naming convention:**

```
{app}-{purpose}-{environment}
Examples:
- myapp-config-prod
- myapp-cache-staging
- myapp-session-dev
```

Must be unique within the Cloudflare account and under 64 characters.

---

### Handling Eventual Consistency

Because KV propagates writes gradually over 60+ seconds, design your application to tolerate stale reads.

**Strategy 1: Embrace staleness**

For configuration that changes infrequently (hourly or daily), a 60-second delay is acceptable. Example: Feature flags that control UI elements. Users won't notice if a flag takes a minute to propagate globally.

**Strategy 2: Version your data**

Include a timestamp or version number in your values:

```json
{
  "version": 2,
  "updated_at": "2025-11-08T12:00:00Z",
  "dark_mode_enabled": true
}
```

This lets you detect stale reads and potentially retry or refresh.

**Strategy 3: Use Durable Objects for coordination**

For data that requires immediate consistency (e.g., counters, coordination), use a Durable Object to handle writes and synchronize state. Then write to KV for fast global reads:

```javascript
// Write: Durable Object updates state and KV
await durableObject.increment();
await CONFIG.put("counter", value);

// Read: KV for speed, Durable Object for guaranteed freshness
const cached = await CONFIG.get("counter");
if (needsFreshData) {
  const fresh = await durableObject.getCounter();
}
```

**Anti-pattern:** Don't build workflows that require reading your own writes immediately. If you write a key and then immediately read it from a different edge location, you may get stale data or a cache miss.

---

### TTL and Expiration

Cloudflare KV supports per-key TTL (time-to-live). Use it for ephemeral data:

```javascript
// Set a key with 1-hour expiration
await SESSION.put("user-token-abc", sessionData, { expirationTtl: 3600 });
```

**Minimum TTL:** 60 seconds. Anything below is rounded up.

**Use cases for TTL:**
- Session tokens (expire after inactivity)
- Cached API responses (refresh every N minutes)
- Temporary data (rate limit counters, temporary access codes)

**Use cases for no TTL:**
- Configuration that persists indefinitely
- Feature flags managed externally
- User preferences

**Note:** OpenMCF's spec includes a `ttl_seconds` field to document a default TTL policy for a namespace. While Cloudflare's API doesn't support namespace-wide TTL, this field serves as documentation and could be enforced by tooling that writes to KV.

---

### Monitoring and Quota Management

Workers KV doesn't provide detailed per-namespace metrics out of the box. You must track usage yourself:

**1. Monitor at the account level:**
- Use Cloudflare's dashboard or API to track daily/monthly read and write operations
- Set up alerts if you approach quota limits (e.g., 80% of free tier)

**2. Instrument your Worker:**

```javascript
// Log KV operations for debugging
async function getConfig(key) {
  const start = Date.now();
  const value = await CONFIG.get(key);
  const duration = Date.now() - start;
  console.log(`KV read: ${key}, duration: ${duration}ms, hit: ${value !== null}`);
  return value;
}
```

**3. Budget for overages:**

If you exceed the free tier, ensure you have a paid plan. On the free tier, operations fail after hitting limits. On paid plans, you incur overage charges.

**4. Audit for orphaned namespaces:**

Periodically list all KV namespaces and verify they're still needed. Delete unused namespaces to avoid unnecessary storage charges.

---

### Common Anti-Patterns

**Anti-pattern 1: Using KV as a real-time database**

Don't use KV for counters, queues, or frequently mutating data. The write rate limits and eventual consistency make it unsuitable.

**Solution:** Use Durable Objects for high-write scenarios.

**Anti-pattern 2: Storing large blobs**

KV supports values up to 25 MiB, but large values increase read latency and Worker memory usage.

**Solution:** For large files, use R2 Storage. For moderate-size data (100 KB - 1 MB), KV works but monitor performance.

**Anti-pattern 3: Assuming immediate consistency**

Don't rely on reading your own writes within 60 seconds across all regions.

**Solution:** Design for eventual consistency or use Durable Objects for immediate coordination.

**Anti-pattern 4: Committing namespace IDs to source control**

Hard-coding KV namespace IDs in `wrangler.toml` makes it hard to manage multiple environments.

**Solution:** Use environment variables or generate `wrangler.toml` dynamically in CI/CD from IaC outputs.

---

## OpenMCF's Approach: Simplicity by Design

OpenMCF abstracts Workers KV into a clean protobuf API that captures the 80% use case while staying out of your way.

### The 80/20 Configuration

The `CloudflareKvNamespaceSpec` includes a small set of fields:

**1. `namespace_name` (required)**
- The human-readable identifier for the namespace
- Must be unique within the Cloudflare account
- Limited to 64 characters
- Example: `myapp-config-prod`

**2. `account_id` (required)**
- The Cloudflare account ID (32 hex characters) that owns the namespace
- A KV namespace is account-scoped, so the account must be specified
- Example: `0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d`

**3. `ttl_seconds` (optional)**
- Default TTL for keys in this namespace (documentation/convention)
- Minimum: 60 seconds (enforced by Cloudflare when writing keys)
- Use `0` or omit for no default expiration
- Example: `3600` (1 hour)

**4. `description` (optional)**
- Short human-readable description of the namespace's purpose
- Max 256 characters
- Example: `"Feature flags and configuration for MyApp production"`

**Rationale:** These fields cover the vast majority of KV namespace configurations. We intentionally omit:
- **Network IDs:** Not applicable to Workers KV (unlike compute resources)
- **Performance tiers:** KV has a single storage tier
- **Access controls:** Managed via Cloudflare API tokens, not per-namespace
- **Replication settings:** KV is globally replicated by default
- **Quota limits:** Set at the account level, not per namespace

This follows the **80/20 principle**: expose the 20% of configuration that 80% of users need.

### Default Choices and Opinions

**Default: No automatic TTL**

We default `ttl_seconds` to 0 (no expiration) because most KV use cases involve persistent configuration. If you need TTL, you can set it per-key when writing.

**Opinion: Explicit naming over auto-generation**

We require you to specify `namespace_name` rather than auto-generating it. This forces intentional naming conventions and avoids "namespace-abc123" clutter.

**Opinion: Description as documentation**

The `description` field isn't required by Cloudflare's API, but we include it to encourage teams to document their namespaces. When you list 20 KV stores in production, descriptive names and descriptions are invaluable.

### Under the Hood: Pulumi for Orchestration

OpenMCF uses **Pulumi (Go)** to provision KV namespaces. The implementation is straightforward:

```go
kvNamespace, err := cloudflare.NewWorkersKvNamespace(ctx, spec.NamespaceName, &cloudflare.WorkersKvNamespaceArgs{
    Title: pulumi.String(spec.NamespaceName),
})
```

**Why Pulumi over Terraform?**
- **Language consistency:** OpenMCF's codebase is Go-based, so Pulumi Go fits naturally
- **Protobuf integration:** Easier to map protobuf specs to Pulumi resources programmatically
- **Equivalent coverage:** Pulumi's Cloudflare provider (bridged from Terraform) supports all KV operations

That said, Terraform would work equally well. The protobuf API remains the same regardless of the underlying IaC tool.

### What We Don't Manage

OpenMCF manages the **namespace lifecycle** (create, update, delete), but not the **contents** (keys and values). Why?

**1. Scale:** A namespace might contain thousands or millions of keys. Managing them in IaC state would be impractical.

**2. Dynamism:** Keys are typically written by application code at runtime, not defined statically in infrastructure config.

**3. Separation of concerns:** Infrastructure defines the container (namespace), application defines the content (keys).

If you need to seed initial data, use Wrangler's bulk import or custom scripts during deployment:

```bash
# Seed feature flags after creating namespace
wrangler kv:bulk put --namespace-id=$NAMESPACE_ID flags.json
```

---

## Configuration Examples: Dev, Staging, Production

### Development: Minimal Namespace

**Use case:** A developer's local sandbox for testing Workers with KV.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-dev-alice
spec:
  namespace_name: myapp-dev-alice
  account_id: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  ttl_seconds: 300  # 5-minute default TTL for quick cache expiration
  description: "Alice's dev environment for MyApp"
```

**Rationale:**
- Short TTL for rapid iteration (cache expires quickly during testing)
- Developer's name in namespace for easy identification
- No production data

---

### Staging: Production-Like Configuration

**Use case:** Staging environment mirroring production setup.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config-staging
spec:
  namespace_name: myapp-config-staging
  account_id: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  ttl_seconds: 0  # No default expiration (persistent config)
  description: "Feature flags and configuration for MyApp staging"
```

**Rationale:**
- No TTL (data persists like production)
- Clear staging designation in name
- Descriptive purpose for clarity

---

### Production: Multi-Namespace Strategy

**Use case:** Production with separate namespaces for different concerns.

**Config namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-config-prod
spec:
  namespace_name: myapp-config-prod
  account_id: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  ttl_seconds: 0  # Persistent configuration
  description: "Feature flags and application configuration (production)"
```

**Cache namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-cache-prod
spec:
  namespace_name: myapp-cache-prod
  account_id: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  ttl_seconds: 3600  # 1-hour default TTL for cache entries
  description: "Cached API responses and transient data (production)"
```

**Session namespace:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: myapp-session-prod
spec:
  namespace_name: myapp-session-prod
  account_id: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  ttl_seconds: 1800  # 30-minute session expiration
  description: "User session tokens and temporary auth data (production)"
```

**Rationale:**
- Separation by purpose (config, cache, session)
- Different TTLs for different use cases
- Easy to monitor and audit per namespace
- Clear naming convention: `{app}-{purpose}-{env}`

---

## Integrating KV with Your Deployment Pipeline

### CI/CD Pattern: IaC First, Then Worker Deployment

**Step 1: Provision KV namespace via IaC (Terraform/Pulumi)**

```bash
# Create namespace in CI/CD
terraform apply
# Output: namespace ID
```

**Step 2: Inject namespace ID into Worker config**

```bash
# Generate wrangler.toml dynamically
cat > wrangler.toml <<EOF
name = "myapp-worker"

[[kv_namespaces]]
binding = "CONFIG"
id = "$NAMESPACE_ID"
EOF
```

**Step 3: Deploy Worker**

```bash
wrangler publish
```

**Key insight:** IaC manages the namespace, Wrangler (or Terraform's Worker script resource) deploys the code. Keep them separate but coordinated.

---

### Wrangler + IaC: How to Avoid Conflicts

**Problem:** If you create a namespace with Wrangler, then define it in Terraform, Terraform may try to create it again (conflict) or adopt it (import required).

**Solution: Import existing resources**

If a namespace was created manually or via Wrangler, import it into Terraform:

```bash
terraform import cloudflare_workers_kv_namespace.config $ACCOUNT_ID/$NAMESPACE_ID
```

Now Terraform manages it.

**Better solution: Choose one source of truth**

- **Development:** Developers use Wrangler freely to create personal namespaces
- **Staging/Production:** All namespaces defined in IaC

Mark dev namespaces with a clear prefix (`dev-username-`) to avoid confusion.

---

### Backup and Migration

**Problem:** KV doesn't support native snapshots or backups.

**Solution: Bulk export via Wrangler**

```bash
# Export all keys in a namespace to JSON
wrangler kv:bulk export --namespace-id=$NAMESPACE_ID > backup.json

# Import into another namespace
wrangler kv:bulk import --namespace-id=$NEW_NAMESPACE_ID backup.json
```

**Best practice:** Schedule periodic backups for critical namespaces (e.g., weekly exports to S3).

---

## Key Takeaways

1. **Workers KV is a globally distributed cache, not a database.** It trades consistency for read speed. Design for eventual consistency from day one.

2. **Use the right tool for the job.** KV excels at read-heavy, infrequently changing data (config, feature flags). For counters or real-time coordination, use Durable Objects.

3. **IaC is the production standard.** Use Wrangler for development, Terraform or Pulumi for staging and production. Both tools are mature and production-ready—choose based on team preference.

4. **Separate namespaces by environment and purpose.** Don't mix dev and prod data. Don't lump config, cache, and session data into one namespace.

5. **Monitor usage and set quotas.** KV's free tier is generous, but exceeding it results in errors. Track operations and storage to avoid surprises.

6. **Implement TTLs where appropriate.** Use expiration for ephemeral data (sessions, cache). Leave config data without TTL.

7. **OpenMCF simplifies the API.** A small field set (`namespace_name`, `account_id`, `ttl_seconds`, `description`) covers 80% of use cases. Advanced patterns happen at the application layer, not in infrastructure config.

---

## Further Reading

- **Cloudflare Workers KV Documentation:** [How KV Works](https://developers.cloudflare.com/kv/concepts/how-kv-works/)
- **Wrangler CLI Reference:** [KV Commands](https://developers.cloudflare.com/workers/wrangler/commands/#kv)
- **Terraform Cloudflare Provider:** [KV Namespace Resource](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/workers_kv_namespace)
- **Pulumi Cloudflare Provider:** [WorkersKvNamespace](https://www.pulumi.com/registry/packages/cloudflare/api-docs/workerskvnamespace/)
- **Workers KV Pricing:** [Cloudflare Pricing](https://developers.cloudflare.com/workers/platform/pricing/#workers-kv)

---

**Bottom Line:** Cloudflare Workers KV gives you globally distributed edge storage with minimal operational overhead. Manage namespaces with IaC, embrace eventual consistency in your application design, and use Wrangler for development velocity. OpenMCF reduces the API surface to three essential fields, letting you focus on building edge applications instead of wrangling infrastructure.

