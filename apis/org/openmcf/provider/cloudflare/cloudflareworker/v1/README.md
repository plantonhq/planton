# Cloudflare Worker

A OpenMCF deployment component for deploying and managing Cloudflare Workers - edge compute functions that run on Cloudflare's global network with true zero cold starts using V8 isolates.

## Overview

Cloudflare Workers revolutionize serverless computing by replacing container-based execution with **V8 Isolates** - the same lightweight isolation technology that Chrome uses for browser tabs. Instead of spinning up new containers for each request (200ms-2s startup time), Workers execute in pre-warmed V8 runtimes in under 5ms, achieving **true zero cold starts** across 330+ global data centers.

This fundamental architectural shift makes Workers ideal for:
- **Edge middleware** (authentication, routing, header manipulation)
- **API gateways** (request validation, rate limiting, transformation)
- **Webhook handlers** (GitHub, Stripe, Twilio integrations)
- **A/B testing and feature flags** (instant routing decisions)
- **Content personalization** (dynamic content delivery)

### Key Features

- **Zero Cold Starts**: Sub-5ms execution start time (vs 200ms+ for Lambda)
- **Global Distribution**: Automatic deployment to 330+ Cloudflare PoPs
- **R2 Bundle Storage**: Immutable artifacts stored in Cloudflare R2
- **KV Integration**: Bind to Cloudflare KV namespaces for edge data
- **Custom Domains**: Automatic DNS and route configuration
- **Environment Variables**: First-class support for config and secrets
- **Compatibility Date Pinning**: Stable runtime behavior across deployments

### Use Cases

Use Cloudflare Workers when you need:

- **Sub-50ms P95 latencies**: Edge execution beats centralized compute
- **Global reach without config**: No region selection, automatic worldwide deployment
- **Cost-efficient high-frequency tasks**: Workers pricing beats Lambda for high-request-count workloads
- **Lightweight HTTP processing**: Request interception, transformation, routing
- **Edge authentication**: JWT validation, API key checks before backend hits

### When NOT to Use

Cloudflare Workers are **not** suitable for:

- Long-running tasks (>30 seconds CPU time limit)
- Heavyweight computations (limited CPU per request)
- File system access (no disk I/O)
- Full Node.js APIs (limited runtime environment)
- Large payload processing (50MB request limit)

**Alternative**: Use AWS Lambda, Google Cloud Functions, or Kubernetes for these scenarios.

## Quick Start

### Minimal Worker

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: hello-worker
spec:
  account_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # Your Cloudflare account ID
  
  script:
    name: hello-worker
    bundle:
      bucket: my-workers-bucket  # R2 bucket containing your worker bundle
      path: builds/hello-worker-v1.0.0.js
  
  compatibility_date: "2025-01-15"
```

This creates a Worker that:
1. Fetches the bundle from R2 bucket `my-workers-bucket/builds/hello-worker-v1.0.0.js`
2. Deploys to all 330+ Cloudflare edge locations
3. Uses the 2025-01-15 compatibility date for stable runtime behavior

### Worker with Custom Domain

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway
spec:
  account_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  
  script:
    name: api-gateway
    bundle:
      bucket: my-workers-bucket
      path: builds/api-gateway-v2.1.0.js
  
  dns:
    enabled: true
    zone_id: { value: "zone123abc" }
    hostname: api.example.com
    route_pattern: api.example.com/*
  
  compatibility_date: "2025-01-15"
```

This additionally:
- Creates DNS record for `api.example.com`
- Attaches Worker to route pattern `api.example.com/*`
- Makes Worker accessible at `https://api.example.com`

## Specification Fields

### Required Fields

- **`account_id`** (string): Cloudflare account ID (exactly 32 hexadecimal characters)
  - Find in: Cloudflare Dashboard → Account Home → Account ID
  - Example: `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`

- **`script`** (object): Worker script configuration
  - **`name`** (string): Unique name for the Worker script
  - **`bundle`** (object): R2 bundle location
    - **`bucket`** (string): R2 bucket name containing the worker bundle
    - **`path`** (string): Path to the bundle file in R2 (e.g., `builds/worker-v1.0.0.js`)

### Optional Fields

- **`kv_bindings`** (list): KV namespace bindings for edge data storage
  - **`name`** (string): Binding name accessible in Worker code
  - **`field_path`** (string): Reference to CloudflareKvNamespace resource
  - Example: Bind `MY_KV` to access data via `env.MY_KV.get("key")`

- **`dns`** (object): Custom domain configuration
  - **`enabled`** (bool): Enable DNS/route configuration - Default: `false`
  - **`zone_id`** (StringValueOrRef): Cloudflare Zone ID for the domain
  - **`hostname`** (string): Fully qualified domain name (e.g., `api.example.com`)
  - **`route_pattern`** (string): URL pattern to match (defaults to `hostname/*`)

- **`compatibility_date`** (string): Runtime compatibility date (YYYY-MM-DD format)
  - Locks Worker to specific runtime version for stability
  - Recommended: Set explicitly to avoid automatic updates
  - Example: `"2025-01-15"`

- **`usage_model`** (enum): Billing model
  - `0` = **BUNDLED** (default): Included requests + CPU time, pay for overages
  - `1` = **UNBOUND**: Pay only for CPU time, no included requests

- **`env`** (object): Environment configuration
  - **`variables`** (map[string]string): Plain-text environment variables
    - Accessible in Worker as `env.VARIABLE_NAME`
    - Example: `LOG_LEVEL: "info"`
  - **`secrets`** (map[string]string): Encrypted secrets
    - Uploaded via Cloudflare Secrets API
    - Never logged or visible in dashboard
    - Example: `API_KEY: "secret-value"`

## Common Patterns

### API Gateway with KV Storage

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway
spec:
  account_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  
  script:
    name: api-gateway
    bundle:
      bucket: workers-prod
      path: api-gateway/v2.1.0.js
  
  kv_bindings:
    - name: RATE_LIMIT_KV
      field_path: rate-limit-namespace-id
    - name: CACHE_KV
      field_path: cache-namespace-id
  
  dns:
    enabled: true
    zone_id: { value: "zone123" }
    hostname: api.example.com
  
  env:
    variables:
      LOG_LEVEL: "info"
      ENVIRONMENT: "production"
    secrets:
      BACKEND_API_KEY: "$secrets-group/backend/api-key"
  
  compatibility_date: "2025-01-15"
```

**Access in Worker**:
```javascript
export default {
  async fetch(request, env) {
    // Check rate limit
    const ip = request.headers.get('CF-Connecting-IP');
    const limit = await env.RATE_LIMIT_KV.get(ip);
    
    // Use environment variables
    console.log(`Environment: ${env.ENVIRONMENT}`);
    
    // Call backend with secret
    const response = await fetch('https://backend.com/api', {
      headers: { 'Authorization': `Bearer ${env.BACKEND_API_KEY}` }
    });
    
    return response;
  }
};
```

### Webhook Handler

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: github-webhook
spec:
  account_id: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
  
  script:
    name: github-webhook-handler
    bundle:
      bucket: workers-prod
      path: webhooks/github-v1.2.0.js
  
  dns:
    enabled: true
    zone_id: { value: "zone123" }
    hostname: webhooks.example.com
    route_pattern: webhooks.example.com/github/*
  
  env:
    secrets:
      GITHUB_WEBHOOK_SECRET: "$secrets-group/github/webhook-secret"
  
  compatibility_date: "2025-01-15"
```

### Multi-Environment Deployment

Deploy separate Workers for staging and production:

```yaml
---
# Staging Worker
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-staging
spec:
  account_id: "staging-account-32-hex-characters"
  script:
    name: api-staging
    bundle:
      bucket: workers-staging
      path: api/v2.1.0.js
  dns:
    enabled: true
    zone_id: { value: "staging-zone-id" }
    hostname: api-staging.example.com
  env:
    variables:
      ENVIRONMENT: "staging"
  compatibility_date: "2025-01-15"

---
# Production Worker
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-prod
spec:
  account_id: "prod-account-32-hex-characters"
  script:
    name: api-prod
    bundle:
      bucket: workers-prod
      path: api/v2.1.0.js
  dns:
    enabled: true
    zone_id: { value: "prod-zone-id" }
    hostname: api.example.com
  env:
    variables:
      ENVIRONMENT: "production"
    secrets:
      API_KEY: "$secrets-group/prod/api-key"
  compatibility_date: "2025-01-15"
```

## R2 Bundle Storage

Workers require JavaScript bundles stored in Cloudflare R2 (S3-compatible object storage). The deployment process:

1. **Build Worker**: `wrangler build` produces bundled JavaScript
2. **Upload to R2**: Store bundle in R2 bucket (e.g., `s3://workers-prod/api-v1.0.0.js`)
3. **Deploy Worker**: OpenMCF fetches bundle from R2 and deploys

**Why R2?**
- **Artifact immutability**: Bundles are versioned, enabling rollbacks
- **Zero egress fees**: Free to download bundles during deployment
- **S3 compatibility**: Use standard S3 tools (AWS CLI, SDKs)

**Upload Example**:
```bash
aws s3 cp dist/worker.js s3://my-workers-bucket/builds/worker-v1.0.0.js \
  --endpoint-url https://<account-id>.r2.cloudflarestorage.com
```

## KV Namespace Bindings

Bind Cloudflare KV namespaces to Workers for edge key-value storage:

```yaml
kv_bindings:
  - name: MY_KV
    field_path: "namespace-id-from-cloudflare"
```

**Access in Worker**:
```javascript
// Read
const value = await env.MY_KV.get("key");

// Write
await env.MY_KV.put("key", "value");

// Delete
await env.MY_KV.delete("key");

// List keys
const keys = await env.MY_KV.list({ prefix: "user:" });
```

## Environment Variables and Secrets

### Variables (Plain Text)

Accessible in Worker code but visible in Cloudflare Dashboard:

```yaml
env:
  variables:
    LOG_LEVEL: "debug"
    API_VERSION: "v2"
```

### Secrets (Encrypted)

Encrypted at rest, never logged or visible:

```yaml
env:
  secrets:
    DATABASE_URL: "$secrets-group/db/connection-string"
    API_KEY: "actual-secret-value"
```

**Best Practice**: Use `$secrets-group/...` references for centralized secret management.

## Compatibility Dates

Cloudflare Workers runtime evolves over time. The `compatibility_date` field locks your Worker to specific runtime behavior:

```yaml
compatibility_date: "2025-01-15"
```

**Why it matters**:
- **Stability**: Prevents unexpected behavior changes from automatic runtime updates
- **Controlled upgrades**: Test new runtime versions in staging before production

**Recommendation**: Always set explicitly. Review Cloudflare's compatibility changelog before updating.

## Cost Model

### BUNDLED Plan (Default)

- **Included**: 100,000 requests/day + 10ms CPU time per request
- **Overages**: $0.50 per million requests, $30 per million GB-seconds CPU

**Best for**: Moderate traffic workloads

### UNBOUND Plan

- **Pay per use**: $0.125 per million requests + $12.50 per million GB-seconds CPU
- **No included requests**: Pure usage-based pricing

**Best for**: High-traffic workloads with low CPU per request

## Best Practices

1. **Always set `compatibility_date`**: Avoid unexpected runtime changes
2. **Version your bundles**: Use semantic versioning in R2 paths (e.g., `v1.2.3.js`)
3. **Use custom domains**: Don't rely on `*.workers.dev` URLs for production
4. **Separate environments**: Use different accounts or workers for dev/staging/prod
5. **Monitor KV usage**: KV has separate pricing (read/write operations)
6. **Keep Workers lightweight**: Optimize bundle size (<1MB recommended)
7. **Test locally**: Use `wrangler dev` before deploying

## Anti-Patterns to Avoid

❌ **Hardcoding secrets in bundle**: Use `env.secrets` instead  
❌ **Heavy computation**: Workers have CPU time limits  
❌ **Large payloads**: 50MB request limit  
❌ **Filesystem access**: Workers don't have disk I/O  
❌ **Long-running tasks**: Use Durable Objects or offload to backend  

## Integration with Cloudflare Services

### With Cloudflare KV

Bind KV namespaces for edge data storage (see examples above).

### With Cloudflare R2

Workers can access R2 buckets via S3-compatible API.

### With Cloudflare D1

Bind D1 databases for edge SQL queries (coming soon).

## Deployment

OpenMCF handles:
1. Fetching bundle from R2
2. Creating Worker script with proper configuration
3. Setting up DNS records (if custom domain enabled)
4. Configuring routes
5. Uploading environment variables and secrets

## Examples

See [examples.md](./examples.md) for complete, working examples including:
- Minimal Worker
- API Gateway with KV and custom domain
- Webhook Handler
- Multi-environment deployments

## Documentation

- **[Research Documentation](./docs/README.md)**: Deep dive into V8 isolates, deployment methods, and architectural decisions
- **[Pulumi Module](./iac/pulumi/README.md)**: Pulumi deployment guide
- **[Terraform Module](./iac/tf/README.md)**: Terraform deployment guide

## Support

- Cloudflare Workers: [Official Docs](https://developers.cloudflare.com/workers/)
- Workers Runtime API: [API Reference](https://developers.cloudflare.com/workers/runtime-apis/)
- Pricing: [Workers Pricing](https://developers.cloudflare.com/workers/platform/pricing/)

## What's Next?

- Explore [examples.md](./examples.md) for complete usage patterns
- Read [docs/README.md](./docs/README.md) for V8 isolates and architectural deep dive
- Deploy using [iac/pulumi/](./iac/pulumi/) or [iac/tf/](./iac/tf/)

---

**Bottom Line**: Cloudflare Workers provide true zero-cold-start serverless computing with sub-5ms execution times across 330+ global locations. Perfect for edge middleware, API gateways, and lightweight HTTP processing.

