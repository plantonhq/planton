# DigitalOcean Function Examples

Complete, copy-paste ready YAML manifests for common serverless function patterns.

**Important**: These examples use App Platform deployment (production-ready) with VPC integration, monitoring, and IaC support. Standalone Functions are intentionally not shown as they lack production features.

---

## Example 1: Simple HTTP API Function (Node.js)

**Use Case**: Public REST API endpoint for handling user requests.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: simple-api
spec:
  function_name: simple-api
  region: nyc1
  runtime: nodejs_20
  github_source:
    repo: myorg/my-functions
    branch: main
    deploy_on_push: true
  source_directory: /functions/simple-api
  memory_mb: 256
  timeout_ms: 3000
  is_web: true
  environment_variables:
    NODE_ENV: production
    LOG_LEVEL: info
```

**Notes:**
- Uses Node.js 20 (latest LTS)
- 256 MB memory (default, sufficient for most APIs)
- 3 second timeout (default for HTTP functions)
- Auto-deploys on Git push
- Exposed as public HTTPS endpoint

---

## Example 2: Production API with Database Access

**Use Case**: API function that securely connects to a DigitalOcean Managed Database via VPC.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: user-api
spec:
  function_name: user-api
  region: nyc1
  runtime: nodejs_18
  github_source:
    repo: mycompany/production-apis
    branch: main
    deploy_on_push: true
  source_directory: /functions/user-api
  memory_mb: 512
  timeout_ms: 15000
  is_web: true
  environment_variables:
    NODE_ENV: production
    LOG_LEVEL: info
    API_VERSION: v2
  secret_environment_variables:
    DATABASE_URL: postgresql://username:password@private-db-host:5432/users
    JWT_SECRET_KEY: super-secret-jwt-key-12345
    STRIPE_API_KEY: sk_live_xxxxxxxxxxxxxxxx
```

**Notes:**
- **VPC Integration**: Database URL uses private network address (only possible via App Platform)
- **Secret Management**: Sensitive values stored in `secret_environment_variables` (encrypted by App Platform)
- 512 MB memory for database connection pooling
- 15 second timeout for complex queries
- Production-ready monitoring via DigitalOcean Insights

**Security**: Never use Standalone Functions for database access. They cannot connect to VPCs and force you to expose databases to `0.0.0.0/0`.

---

## Example 3: Scheduled Background Job (Python)

**Use Case**: Nightly cleanup task that runs at midnight.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: nightly-cleanup
spec:
  function_name: nightly-cleanup
  region: nyc1
  runtime: python_311
  github_source:
    repo: myorg/background-jobs
    branch: production
    deploy_on_push: false
  source_directory: /jobs/cleanup
  memory_mb: 1024
  timeout_ms: 300000  # 5 minutes (max allowed)
  is_web: false
  cron_schedule: "0 0 * * *"  # At midnight daily
  environment_variables:
    PYTHON_ENV: production
    CLEANUP_BATCH_SIZE: "1000"
  secret_environment_variables:
    DATABASE_URL: postgresql://cleanup-user:password@private-db:5432/analytics
    S3_ACCESS_KEY: aws-access-key
    S3_SECRET_KEY: aws-secret-key
```

**Notes:**
- **Scheduled Execution**: Cron schedule triggers function automatically
- `is_web: false` disables HTTP endpoint (background job only)
- 1 GB memory for processing large datasets
- 5 minute timeout (maximum) for long-running tasks
- `deploy_on_push: false` for manual deployment control

---

## Example 4: Webhook Handler (Go)

**Use Case**: High-performance webhook processor for third-party integrations (Stripe, GitHub, etc.).

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: webhook-processor
spec:
  function_name: webhook-processor
  region: sfo3
  runtime: go_121
  github_source:
    repo: myorg/webhooks
    branch: main
    deploy_on_push: true
  source_directory: /webhooks/processor
  entrypoint: main
  memory_mb: 256
  timeout_ms: 10000
  is_web: true
  environment_variables:
    GO_ENV: production
  secret_environment_variables:
    STRIPE_WEBHOOK_SECRET: whsec_xxxxxxxxxxxxxxxx
    GITHUB_WEBHOOK_SECRET: ghp_xxxxxxxxxxxx
    REDIS_URL: redis://private-redis:6379
```

**Notes:**
- Go runtime for maximum performance and low cold start latency
- Explicit `entrypoint: main` for Go functions
- 10 second timeout for webhook processing
- Secret verification keys stored securely
- Redis connection via VPC private network

---

## Example 5: Image Processing Function (Python with High Memory)

**Use Case**: Image resizing and optimization triggered by S3 uploads.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: image-processor
spec:
  function_name: image-processor
  region: nyc1
  runtime: python_311
  github_source:
    repo: myorg/image-processing
    branch: main
    deploy_on_push: true
  source_directory: /functions/image-processor
  memory_mb: 2048  # Maximum memory for image processing
  timeout_ms: 60000  # 1 minute for large images
  is_web: true
  environment_variables:
    PYTHON_ENV: production
    MAX_IMAGE_SIZE_MB: "50"
    QUALITY_PRESET: high
  secret_environment_variables:
    SPACES_ACCESS_KEY: spaces-access-key
    SPACES_SECRET_KEY: spaces-secret-key
    SPACES_BUCKET: my-image-bucket
```

**Notes:**
- 2 GB memory (maximum) for processing large images
- 60 second timeout for complex transformations
- DigitalOcean Spaces (S3-compatible) for storage
- High memory critical for libraries like Pillow, ImageMagick

---

## Example 6: Multi-Environment Setup (Staging vs Production)

**Staging Function:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: api-staging
spec:
  function_name: api-staging
  region: nyc1
  runtime: nodejs_20
  github_source:
    repo: mycompany/apis
    branch: develop  # Staging branch
    deploy_on_push: true
  source_directory: /functions/api
  memory_mb: 256
  timeout_ms: 5000
  is_web: true
  environment_variables:
    NODE_ENV: staging
    LOG_LEVEL: debug
  secret_environment_variables:
    DATABASE_URL: postgresql://staging-db:5432/users
```

**Production Function:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: api-production
spec:
  function_name: api-production
  region: nyc1
  runtime: nodejs_20
  github_source:
    repo: mycompany/apis
    branch: main  # Production branch
    deploy_on_push: true
  source_directory: /functions/api
  memory_mb: 512  # Higher memory for production
  timeout_ms: 15000
  is_web: true
  environment_variables:
    NODE_ENV: production
    LOG_LEVEL: info
  secret_environment_variables:
    DATABASE_URL: postgresql://prod-db:5432/users
    API_KEY: prod-api-key-12345
```

**Notes:**
- Same codebase, different branches (`develop` vs `main`)
- Different resource allocations (staging uses less memory)
- Separate databases and secrets per environment
- Different log levels for debugging

---

## Example 7: Hourly Data Sync (Scheduled Task)

**Use Case**: Sync data from external API every hour.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: hourly-sync
spec:
  function_name: hourly-sync
  region: ams3
  runtime: python_310
  github_source:
    repo: myorg/data-sync
    branch: main
    deploy_on_push: true
  source_directory: /sync/hourly
  memory_mb: 512
  timeout_ms: 120000  # 2 minutes
  is_web: false
  cron_schedule: "0 * * * *"  # Every hour at minute 0
  environment_variables:
    SYNC_BATCH_SIZE: "500"
  secret_environment_variables:
    EXTERNAL_API_KEY: ext-api-key-xxx
    DATABASE_URL: postgresql://sync-db:5432/data
```

**Notes:**
- Runs every hour automatically
- `is_web: false` - no HTTP endpoint needed
- 2 minute timeout for API rate limits
- External API key stored securely

---

## Example 8: PHP Web Function (WordPress REST API)

**Use Case**: PHP function for WordPress API integration.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: wordpress-api
spec:
  function_name: wordpress-api
  region: lon1
  runtime: php_82
  github_source:
    repo: myorg/wordpress-functions
    branch: main
    deploy_on_push: true
  source_directory: /functions/wp-api
  memory_mb: 512
  timeout_ms: 10000
  is_web: true
  environment_variables:
    PHP_ENV: production
  secret_environment_variables:
    WP_API_KEY: wp-api-key-xxx
    WP_DB_HOST: private-mysql:3306
    WP_DB_USER: wordpress
    WP_DB_PASSWORD: secure-password
```

**Notes:**
- PHP 8.2 runtime
- WordPress database via VPC private network
- API key authentication
- 512 MB for WordPress libraries

---

## Example 9: Event-Driven Function (Database Trigger Simulation)

**Use Case**: Function triggered by database changes (simulated via cron + change detection).

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: db-change-handler
spec:
  function_name: db-change-handler
  region: nyc1
  runtime: go_120
  github_source:
    repo: myorg/event-handlers
    branch: main
    deploy_on_push: true
  source_directory: /handlers/db-changes
  entrypoint: main
  memory_mb: 256
  timeout_ms: 30000
  is_web: false
  cron_schedule: "*/5 * * * *"  # Every 5 minutes
  secret_environment_variables:
    DATABASE_URL: postgresql://events-db:5432/changes
    REDIS_URL: redis://private-redis:6379
    WEBHOOK_URL: https://api.example.com/webhook
```

**Notes:**
- Polls database every 5 minutes for changes
- Uses Redis to track processed records (via VPC)
- Sends webhooks for new changes
- Go for high performance

---

## Example 10: Minimal Function (Development/Testing)

**Use Case**: Simplest possible function for learning or testing.

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanFunction
metadata:
  name: hello-world
spec:
  function_name: hello-world
  region: nyc1
  runtime: nodejs_20
  github_source:
    repo: myorg/test-functions
    branch: main
    deploy_on_push: true
  source_directory: /functions/hello
  memory_mb: 128  # Minimum memory
  timeout_ms: 1000  # 1 second
  is_web: true
```

**Notes:**
- Minimal resources for cost efficiency
- No environment variables
- Perfect for Hello World examples
- Still gets VPC, monitoring (unlike Standalone Functions)

---

## Common Patterns Summary

| Use Case | Runtime | Memory | Timeout | is_web | Cron | Key Features |
|----------|---------|--------|---------|--------|------|--------------|
| Public API | Node.js | 256-512 MB | 3-15s | true | - | HTTP endpoint, secrets, DB access |
| Background Job | Python | 512-2048 MB | 60-300s | false | Yes | Scheduled, long-running, high memory |
| Webhook Handler | Go | 256 MB | 5-10s | true | - | Fast, event-driven, secret verification |
| Image Processing | Python | 2048 MB | 60s | true | - | Maximum memory, S3 integration |
| Data Sync | Python | 512 MB | 120s | false | Yes | Scheduled, external API, batch processing |

---

## Validation Checklist

Before deploying, ensure:

- ✅ `function_name` is unique and 1-64 characters
- ✅ `region` matches your database/VPC region
- ✅ `github_source.repo` is in `owner/repo` format
- ✅ `source_directory` contains `project.yml` file
- ✅ `memory_mb` is one of: 128, 256, 512, 1024, 2048
- ✅ `timeout_ms` is ≤ 300000 (5 minutes)
- ✅ Secrets are in `secret_environment_variables` (never in `environment_variables`)
- ✅ Database URLs use private network addresses (VPC)
- ✅ `cron_schedule` is valid cron syntax if set

---

## Production Best Practices

### 1. Secret Management
**✅ Do:**
```yaml
secret_environment_variables:
  DATABASE_URL: postgresql://private-host:5432/db
  API_KEY: secret-key
```

**❌ Don't:**
```yaml
environment_variables:
  DATABASE_URL: postgresql://public-host:5432/db  # Exposed in logs!
```

### 2. VPC Integration
**✅ Do (App Platform):**
```yaml
secret_environment_variables:
  DATABASE_URL: postgresql://10.116.0.5:5432/db  # Private VPC IP
```

**❌ Don't (Standalone Functions):**
```yaml
# Standalone Functions cannot access VPC
DATABASE_URL: postgresql://db-public.digitalocean.com:25060/db?sslmode=require
```

### 3. Resource Allocation
- **Start small**: Begin with 256 MB, increase if cold starts are slow
- **Timeout tuning**: Set timeout slightly higher than expected execution time
- **Memory = Speed**: Higher memory = faster CPU allocation

---

## Troubleshooting

### Issue: Function times out
**Cause**: Timeout too low or function doing too much work.

**Solution**:
```yaml
timeout_ms: 30000  # Increase from 3000 to 30000 (30 seconds)
```

### Issue: Out of memory errors
**Cause**: Memory allocation too low for workload.

**Solution**:
```yaml
memory_mb: 512  # Increase from 256 to 512 MB
```

### Issue: Cannot connect to database
**Cause**: Using Standalone Functions or public database URL without SSL.

**Solution**: Use App Platform (this component) with private VPC address:
```yaml
secret_environment_variables:
  DATABASE_URL: postgresql://private-db:5432/db  # VPC private address
```

---

## Further Reading

- **Component Overview**: See [README.md](./README.md)
- **Comprehensive Guide**: See [docs/README.md](./docs/README.md)
- **Pulumi Module**: See [iac/pulumi/README.md](./iac/pulumi/README.md)

---

**TL;DR**: Always use App Platform deployment (this component) for production. Use `secret_environment_variables` for sensitive data. Match function resources (memory, timeout) to workload. Enable VPC for database access.

