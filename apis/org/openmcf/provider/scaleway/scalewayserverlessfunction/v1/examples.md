# Scaleway Serverless Function Examples

Complete, copy-paste ready YAML manifests for common serverless function patterns.

---

## Example 1: Simple HTTP API (Node.js)

**Use Case**: Public REST API endpoint with environment configuration.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: simple-api
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: node20
  handler: handler.handler
  privacy: public
  description: Simple HTTP API function
  memory_limit_mb: 256
  timeout_seconds: 30
  http_option: redirected
  env:
    variables:
      - name: NODE_ENV
        value: production
      - name: LOG_LEVEL
        value: info
```

**Notes:**
- Node.js 20 runtime
- 256 MB memory (default, sufficient for most APIs)
- HTTP redirected to HTTPS for security
- Public endpoint, no authentication required
- No zip_file -- deploy code separately via CLI or CI/CD

---

## Example 2: VPC-Connected Database API (Python)

**Use Case**: API function that securely connects to a managed database via Private Network.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: user-api
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: python311
  handler: handler.handle
  privacy: public
  memory_limit_mb: 512
  timeout_seconds: 60
  http_option: redirected
  env:
    variables:
      - name: PYTHON_ENV
        value: production
      - name: LOG_LEVEL
        value: info
    secrets:
      - name: DATABASE_URL
        value: postgresql://dbuser:password@10.0.1.5:5432/users
      - name: JWT_SECRET
        value: super-secret-jwt-key-12345
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-private-network
      fieldPath: status.outputs.private_network_id
```

**Key points:**
- Private Network connectivity for secure database access
- Database URL uses private IP (only reachable via VPC)
- Secrets stored encrypted, separate from plain variables
- 512 MB memory for database connection handling
- 60 second timeout for complex queries

---

## Example 3: Scheduled Background Job with Cron

**Use Case**: Nightly cleanup task that processes stale data.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: nightly-cleanup
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: python312
  handler: cleanup.run
  privacy: private
  memory_limit_mb: 1024
  timeout_seconds: 300
  min_scale: 0
  env:
    variables:
      - name: CLEANUP_BATCH_SIZE
        value: "1000"
    secrets:
      - name: DATABASE_URL
        value: postgresql://cleanup-user:password@10.0.1.5:5432/analytics
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-private-network
      fieldPath: status.outputs.private_network_id
  cron_triggers:
    - name: nightly-cleanup
      schedule: "0 2 * * *"
      args: '{"cleanup_type": "stale_sessions", "max_age_days": 30}'
```

**Notes:**
- Private function (only invocable via cron trigger, not public HTTP)
- 1 GB memory for processing large datasets
- 5 minute timeout (maximum) for long-running cleanups
- Cron trigger runs at 2:00 AM daily
- JSON args passed to the function specify cleanup parameters

---

## Example 4: Multiple Cron Triggers (Data Sync)

**Use Case**: Function with multiple schedules for different sync tasks.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: data-sync
  org: my-org
  env: production
spec:
  region: nl-ams
  runtime: go124
  handler: Handle
  privacy: private
  memory_limit_mb: 256
  timeout_seconds: 120
  env:
    secrets:
      - name: EXTERNAL_API_KEY
        value: ext-api-key-xxx
      - name: DATABASE_URL
        value: postgresql://sync-db:5432/data
  cron_triggers:
    - name: hourly-incremental
      schedule: "0 * * * *"
      args: '{"sync_mode": "incremental"}'
    - name: daily-full
      schedule: "0 3 * * *"
      args: '{"sync_mode": "full", "batch_size": 5000}'
    - name: weekly-audit
      schedule: "0 6 * * 0"
      args: '{"sync_mode": "audit", "report_email": "ops@example.com"}'
```

**Notes:**
- Go runtime for high performance and low cold start
- Three different cron schedules with different JSON args
- Each trigger invokes the same function with different parameters
- Private function -- no HTTP endpoint, only cron-triggered

---

## Example 5: Zip-Deployed Function

**Use Case**: Fully IaC-managed function with code deployment.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: webhook-handler
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: node22
  handler: index.handler
  privacy: public
  memory_limit_mb: 256
  timeout_seconds: 10
  http_option: redirected
  zip_file: "./dist/webhook-handler.zip"
  zip_hash: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
  env:
    variables:
      - name: NODE_ENV
        value: production
    secrets:
      - name: WEBHOOK_SECRET
        value: whsec_xxxxxxxxxxxxxxxx
```

**Notes:**
- Code deployed via zip archive (IaC-managed)
- `zip_hash` triggers redeployment when the archive changes
- The IaC module automatically sets `deploy = true` when zip_file is provided
- Best for simple functions with infrequent code changes

---

## Example 6: Always-Warm Function (No Cold Starts)

**Use Case**: Latency-sensitive API that must respond instantly.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: realtime-api
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: go124
  handler: Handle
  privacy: public
  memory_limit_mb: 512
  min_scale: 2
  max_scale: 50
  timeout_seconds: 5
  http_option: redirected
  env:
    variables:
      - name: CACHE_TTL_SECONDS
        value: "60"
```

**Notes:**
- `min_scale: 2` keeps 2 instances always running (no cold starts)
- Higher `max_scale: 50` for traffic spikes
- Go runtime for fastest cold starts if additional instances are needed
- **Cost implication**: `min_scale > 0` incurs continuous billing

---

## Example 7: Minimal Function (Development/Testing)

**Use Case**: Simplest possible function for learning or testing.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: hello-world
spec:
  region: fr-par
  runtime: python313
  handler: handler.handle
  privacy: public
```

**Notes:**
- Minimal required fields only
- Defaults: 256 MB memory, 300s timeout, 0-20 scaling, HTTP enabled
- No environment variables, no networking, no cron
- Perfect for Hello World examples

---

## Example 8: Custom Domain via DNS Record

**Use Case**: Expose a function at a custom domain like `api.example.com`.

Deploy both resources -- the function and a CNAME record pointing to it:

```yaml
# Function
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: my-api
  org: my-org
  env: production
spec:
  region: fr-par
  runtime: node20
  handler: handler.handler
  privacy: public
  http_option: redirected
  env:
    variables:
      - name: NODE_ENV
        value: production
---
# DNS Record pointing to the function
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: api-dns
  org: my-org
  env: production
spec:
  zone_name:
    value: "example.com"
  name: "api"
  type: CNAME
  data:
    valueFrom:
      kind: ScalewayServerlessFunction
      name: my-api
      fieldPath: status.outputs.domain_name
  ttl: 300
```

**Notes:**
- The DNS record creates a dependency edge: DNS -> Function
- `domain_name` output is automatically wired via `valueFrom`
- Low TTL (300s) for infrastructure-bound records

---

## Common Patterns Summary

| Use Case | Runtime | Memory | Timeout | Privacy | Cron | Key Features |
|----------|---------|--------|---------|---------|------|--------------|
| Public API | Node.js/Go | 256-512 MB | 5-30s | public | No | HTTP endpoint, env vars |
| Database API | Python | 512-1024 MB | 30-60s | public | No | VPC, secrets, DB access |
| Background Job | Python | 512-2048 MB | 60-300s | private | Yes | Scheduled, long-running |
| Data Sync | Go | 256 MB | 60-120s | private | Yes | Multiple cron triggers |
| Webhook | Node.js | 256 MB | 5-10s | public | No | Fast, event-driven |
| Always-Warm | Go | 256-512 MB | 5s | public | No | min_scale > 0, no cold starts |

---

## Validation Checklist

Before deploying, ensure:

- `runtime` is a valid Scaleway runtime string (e.g., "node20", "python311")
- `handler` matches the runtime convention (e.g., "handler.handle" for Python)
- `privacy` is either `public` or `private`
- `env.secrets` contains all sensitive values (never in `env.variables`)
- `cron_triggers[].schedule` uses valid CRON syntax
- `cron_triggers[].args` is valid JSON
- Database URLs in secrets use private IPs when `private_network_id` is set
- `zip_hash` changes when `zip_file` content changes (for redeployment)

---

## Production Best Practices

### 1. Secret Management

**Do:**
```yaml
env:
  secrets:
    - name: DATABASE_URL
      value: postgresql://10.0.1.5:5432/db
```

**Don't:**
```yaml
env:
  variables:
    - name: DATABASE_URL
      value: postgresql://public-host:5432/db
```

### 2. VPC for Database Access

Always use Private Network when accessing databases:

```yaml
spec:
  private_network_id:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: my-network
      fieldPath: status.outputs.private_network_id
  env:
    secrets:
      - name: DATABASE_URL
        value: postgresql://10.0.1.5:5432/db
```

### 3. Resource Sizing

- **Start small**: 256 MB, increase if cold starts are slow
- **Memory = Speed**: Higher memory allocates more CPU proportionally
- **Timeout tuning**: Set slightly higher than expected execution time
- **Scale-to-zero**: Keep `min_scale: 0` unless latency is critical

### 4. Cron Best Practices

- Use descriptive trigger names for observability
- Always pass valid JSON in `args` (use `"{}"` for no arguments)
- Set private functions for cron-only workloads (no public HTTP needed)
- Monitor cron execution via Scaleway Cockpit

---

## Further Reading

- **Component Overview**: See [README.md](./README.md)
- **Scaleway Functions Runtimes**: [Official Documentation](https://www.scaleway.com/en/docs/serverless-functions/reference-content/functions-runtimes/)
- **CRON Schedule Reference**: [Scaleway Docs](https://www.scaleway.com/en/docs/serverless/functions/reference-content/cron-schedules/)
