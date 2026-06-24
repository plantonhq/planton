# Deploying Cloudflare Workers: Edge Computing Without Containers

## Introduction

For years, serverless computing meant accepting a brutal trade-off: pay for convenience with cold starts measured in hundreds of milliseconds, or manage your own infrastructure. AWS Lambda popularized the serverless paradigm by abstracting away servers, but it couldn't escape the fundamental physics of booting containers—every concurrent execution required spinning up a fresh microVM, initializing a runtime, and loading your code. For latency-sensitive applications, this was unacceptable.

**Cloudflare Workers took a different approach.** Instead of containers or virtual machines, they run code in V8 Isolates—the same lightweight, secure context technology that Google Chrome uses to isolate browser tabs. A single `workerd` process starts once on a physical machine, paying the runtime initialization cost exactly once. When a request arrives, Cloudflare doesn't boot a new VM; it creates a new isolate within that already-warm runtime in under 5 milliseconds.

The result? **True zero cold starts.** Workers execute in all 330+ Cloudflare data centers globally, delivering P95 latencies under 50ms for edge-native applications—performance that container-based FaaS platforms simply cannot match.

But this architectural choice isn't free. Workers don't run arbitrary binaries. They don't have full Node.js APIs or filesystem access. They're designed for a specific class of problem: **intercepting, transforming, and routing HTTP requests at the edge**—API gateways, authentication middleware, A/B testing, webhook handlers, and lightweight data aggregation. Not heavyweight data processing or long-running batch jobs.

This guide explains the deployment landscape for Cloudflare Workers, from manual dashboard clicks to production-grade Infrastructure as Code, and shows how OpenMCF solves the most challenging aspect of Workers deployment: **managing the separation between code artifacts and infrastructure configuration**.

---

## Why Cloudflare Workers Exist: The V8 Isolate Model

### The Architecture That Changes Everything

Traditional serverless platforms (AWS Lambda, Google Cloud Functions, Azure Functions) are built on containers. Each function execution runs in a firecracker microVM or similar isolated environment. This provides maximum flexibility—you can run arbitrary code, access OS-level APIs, and use any runtime—but it introduces unavoidable latency.

Cloudflare Workers replace the container with the **V8 Isolate**:

| Characteristic | V8 Isolate (Workers) | Container/MicroVM (Lambda) |
|---------------|---------------------|---------------------------|
| **Startup Time** | <5ms | 200ms - 2s+ |
| **Memory Overhead** | ~1-2MB per isolate | ~50-100MB per container |
| **Supported Runtimes** | JavaScript, TypeScript, Python, WASM | Node.js, Python, Java, Go, Ruby, C#, etc. |
| **API Surface** | Service Worker API (fetch-based) | Full OS/runtime APIs (fs, net, etc.) |
| **Execution Model** | Shared runtime, lightweight isolation | Dedicated VM per concurrent execution |
| **Primary Use Case** | Edge middleware, request interception | Backend compute, data processing |

This is not an incremental improvement—it's a fundamentally different execution model. The V8 runtime is shared across all Workers on a single machine. Isolation is achieved through V8's built-in security model, not through operating system virtualization.

### Strategic Positioning: Workers vs. Lambda

Cloudflare Workers are not a drop-in replacement for AWS Lambda. They're an alternative designed for a specific architectural pattern:

**Use Workers when:**
- Latency is critical (sub-50ms P95 response times)
- You're building edge middleware (auth, routing, caching, header manipulation)
- Your logic is stateless or uses Cloudflare's native bindings (KV, R2, D1, Durable Objects)
- You need global distribution with zero configuration (automatic deployment to 330+ PoPs)
- Cost efficiency matters for high-frequency, low-CPU tasks

**Use Lambda when:**
- You need long execution times (>5 minutes)
- You require full Node.js/OS APIs (filesystem access, native binaries)
- You're deeply integrated with AWS services (RDS, DynamoDB, S3 with IAM roles)
- You're processing large payloads or running heavyweight computations

The platforms are complementary, not competing. A well-designed architecture might use Workers for API gateway logic and authentication, while delegating heavy processing to Lambda or Kubernetes.

---

## The Deployment Spectrum: From Manual to Production

Cloudflare provides a complete range of deployment methods. All programmatic approaches ultimately consume the same underlying **Cloudflare v4 API**, which implements a two-step deployment model: upload a new script version, then activate it.

### Level 0: The Dashboard Quick Edit (Anti-Pattern)

**What it is:** Cloudflare's web console includes a "Quick Edit" feature with an in-browser code editor where you can write and deploy Workers directly.

**What it solves:** Zero-friction experimentation. This is the fastest way to see "hello world" execute at the edge.

**What it doesn't solve:** Everything that matters for production. Manual dashboard edits create **configuration drift**—the deployed code diverges from your source-of-truth (Git). Any hotfix made in the Quick Edit breaks declarative management. If an IaC tool later tries to reconcile state, it will revert your manual changes.

**Verdict:** Use it for learning the platform and testing concepts. **Never use it for staging or production.** The moment you adopt IaC, treat the dashboard as read-only.

---

### Level 1: The Wrangler CLI (Developer's Tool)

**What it is:** Wrangler is the official, canonical CLI for managing the entire Worker lifecycle. It's a Node.js package (`npm install -g wrangler`) that handles bundling, deployment, and local development.

**Core workflow:**

```bash
# Initialize a new project
npx wrangler init my-worker

# Local development (uses actual workerd runtime via Miniflare)
npx wrangler dev

# Deploy to production
npx wrangler deploy
```

The `wrangler deploy` command performs a complex sequence:
1. Reads `wrangler.toml` configuration
2. Bundles the script with esbuild (transpiling TypeScript, inlining dependencies)
3. Constructs metadata defining bindings (KV, R2, D1, secrets, etc.)
4. Uploads the script to the API (creating a new Version)
5. Calls the `/deployments` API to activate that Version

**What it solves:**
- **Best-in-class local development:** `wrangler dev` runs the actual `workerd` runtime locally (not an emulator), with simulated KV/R2/D1 bindings for fully offline development
- **Automatic bundling:** Handles TypeScript transpilation, module resolution, and WASM/asset handling
- **Fast iteration:** The canonical "inner loop" for Worker development

**What it doesn't solve:**
- **State management:** Wrangler is imperative. Running `wrangler deploy` twice doesn't detect drift or manage declarative state.
- **Multi-environment complexity:** While `wrangler.toml` supports `[env.staging]` and `[env.production]` sections, these create *separate Workers* (named `my-worker-staging` and `my-worker-production`), not a single Worker with promoted versions.
- **Separation of build and deploy:** Wrangler tightly couples bundling and deployment. You can't easily "build once, deploy many times."

**Verdict:** Essential for developers. This is how you write and test Workers locally. But for production deployment pipelines, you need IaC tools to consume Wrangler's output, not replace it.

---

### Level 2: CI/CD with Wrangler (Common Production Pattern)

**What it is:** Wrapping `wrangler deploy` in a GitHub Actions or GitLab CI pipeline.

**GitHub Actions example:**

```yaml
name: Deploy Worker
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
```

This pattern is ubiquitous and officially recommended for teams focused on rapid iteration.

**What it solves:**
- **GitOps-driven automation:** Code merged to `main` automatically deploys
- **Credential management:** Secrets stored in CI platform (GitHub Secrets, GitLab CI/CD variables)
- **Fast, familiar workflow:** Developers already use CI/CD; this extends it to Workers

**What it doesn't solve:**
- **Declarative infrastructure:** This is still imperative deployment. The pipeline runs a command, not a reconciliation loop.
- **Multi-resource orchestration:** If your Worker depends on a new KV namespace or R2 bucket, you must provision those separately (or use separate IaC).
- **Artifact immutability:** Each push triggers a rebuild. You can't "deploy the exact artifact from staging to production."

**Verdict:** Perfectly acceptable for small teams, simple Workers, or rapid-iteration projects. But for complex, multi-resource infrastructure or teams that want true declarative management, IaC is the next evolution.

---

### Level 3: Infrastructure as Code (The Production Solution)

**What it is:** Using Terraform or Pulumi to declaratively define Workers and their entire ecosystem (KV namespaces, R2 buckets, DNS records, routes).

#### Terraform: The Mature Standard

Terraform's official `cloudflare/cloudflare` provider is automatically generated from Cloudflare's OpenAPI specification, ensuring it stays current with new features. Recent versions have adopted a new, more accurate model that mirrors the platform's architecture:

- `cloudflare_worker`: Defines the Worker service name
- `cloudflare_worker_version`: Defines an immutable Version (script content + bindings)
- `cloudflare_workers_deployment`: Activates a specific Version

**Example:**

```hcl
resource "cloudflare_workers_script" "api" {
  account_id  = var.cloudflare_account_id
  script_name = "api-gateway"
  content     = file("./dist/worker.js")
  main_module = "index.js"

  bindings = [
    {
      name         = "CACHE"
      type         = "kv_namespace"
      namespace_id = cloudflare_workers_kv_namespace.cache.id
    },
    {
      name = "ENVIRONMENT"
      type = "plain_text"
      text = "production"
    }
  ]
}

resource "cloudflare_workers_route" "api_route" {
  zone_id = var.cloudflare_zone_id
  pattern = "api.example.com/*"
  script  = cloudflare_workers_script.api.script_name
}
```

**What it solves:**
- **Declarative state:** Terraform tracks what exists. Running `terraform apply` twice is idempotent.
- **Multi-resource orchestration:** Define Worker + KV + R2 + DNS in one config, with automatic dependency ordering.
- **Plan/preview:** `terraform plan` shows exactly what will change before applying.
- **Mature ecosystem:** Broad community support, extensive documentation, battle-tested in production.

#### Pulumi: The Programmer's IaC

Pulumi's Cloudflare provider is bridged from Terraform's provider, ensuring 100% resource parity. The difference is the interface: instead of HCL, you write infrastructure in TypeScript, Python, or Go.

**Example (TypeScript):**

```typescript
import * as cloudflare from "@pulumi/cloudflare";

const worker = new cloudflare.WorkerScript("api-gateway", {
    accountId: cloudflareAccountId,
    name: "api-gateway",
    content: fs.readFileSync("./dist/worker.js", "utf8"),
    kvNamespaceBindings: [{
        name: "CACHE",
        namespaceId: cacheNamespace.id,
    }],
});
```

**What it adds over Terraform:**
- **Built-in secret encryption:** `pulumi config set cloudflare:apiToken --secret <value>` encrypts secrets in state
- **Managed state by default:** Pulumi Cloud handles state with concurrency locking out-of-the-box
- **Programming language expressiveness:** Loops, conditionals, unit tests, and type safety

**The Critical Bundle Management Problem**

Both Terraform and Pulumi have a fundamental limitation: **they don't build code; they deploy it.**

The `content` argument expects a path to the final, bundled JavaScript artifact. But Workers are typically written in TypeScript with npm dependencies. Someone must run the bundler (esbuild via Wrangler).

The recommended pattern is a **two-step pipeline**:

1. **Build Step:** `npx wrangler deploy --dry-run --outdir=dist` (builds the bundle without deploying)
2. **Deploy Step:** `terraform apply` or `pulumi up` (deploys the artifact from `./dist`)

This separation is essential but introduces a challenge: **how do you version and reference the artifact?**

**Verdict:** IaC is the production standard. Choose Terraform for maximum ecosystem maturity and familiarity. Choose Pulumi for superior secret handling, type safety, and managed state. Both require solving the artifact management problem.

---

## The Secret Management Crisis

Cloudflare Secrets are **write-only**. Once set via `wrangler secret put API_KEY` or the API, their value cannot be read back—not via the API, not via the Dashboard. This is a deliberate security design.

But it's fundamentally incompatible with declarative IaC. Tools like Terraform and Pulumi operate on a reconciliation loop:

```
current_state (read from API) vs. desired_state (from config) → calculate diff → apply changes
```

If the current state cannot be read, the tool cannot detect drift. You can set a secret once, but you can't verify it matches your config. If someone manually changes it via `wrangler secret put`, the IaC tool will never know.

### Anti-Pattern: Secrets in the Spec

Including secrets directly in a Worker resource (e.g., `spec.env.secrets: {API_KEY: "sk-xyz"}`) creates two problems:

1. **Security:** Secrets leak into state files (Terraform stores them in plaintext; Pulumi encrypts but still persists)
2. **False declarativeness:** The tool can *set* the secret but can't *verify* it, making this an imperative "fire-and-forget" command disguised as declarative config

### The Solution: Separation of Concerns

The correct pattern is to split secret management from Worker deployment:

- **Worker Deployment Resource (declarative):** Manages the script, bindings, routes, and non-sensitive environment variables
- **Secret Provisioning (imperative, write-only):** Handled via a separate process (CI/CD pipeline step, sealed secrets in Kubernetes, external secret manager)

OpenMCF follows this pattern, as explained below.

---

## OpenMCF's Approach: R2-Based Artifact Storage

OpenMCF's `CloudflareWorker` API solves the two hardest problems in Workers deployment: **artifact management** and **secret handling**.

### The R2 Bundle Model

Instead of requiring a local file path (`./dist/worker.js`), the `CloudflareWorkerSpec` references a **bundle stored in Cloudflare R2**:

```protobuf
message CloudflareWorkerScript {
  string name = 1;
  CloudflareWorkerScriptBundleR2Object bundle = 2;
}

message CloudflareWorkerScriptBundleR2Object {
  string bucket = 1;  // R2 bucket name
  string path = 2;    // Path to the bundled script (e.g., "my-worker/v1.2.3.js")
}
```

**Why this matters:**

1. **Decouples build from deploy:** Your CI pipeline builds the bundle with `wrangler deploy --dry-run --outdir=dist`, uploads the result to R2, and commits the updated R2 path to Git.
2. **Immutable artifacts:** The same artifact (identified by its R2 path) can be deployed to staging, validated, and promoted to production without rebuilding.
3. **Versioned history:** R2 becomes an artifact registry. You can reference older versions for rollbacks.
4. **GitOps-native:** The Git commit references the artifact URL, not the artifact content. ArgoCD or Flux can reconcile the Worker by fetching the bundle from R2.

**Example workflow:**

```yaml
# cloudflareworker.yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway-prod
spec:
  account_id: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  script:
    name: api-gateway
    bundle:
      bucket: planton-worker-bundles
      path: api-gateway/v1.2.3/worker.js
  compatibility_date: "2024-03-01"
  dns:
    enabled: true
    zone_id: z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4
    hostname: api.example.com
    route_pattern: api.example.com/*
```

CI pipeline:
1. Build: `npx wrangler deploy --dry-run --outdir=dist`
2. Upload: `aws s3 cp dist/worker.js s3://planton-worker-bundles/api-gateway/v1.2.3/worker.js` (R2 is S3-compatible)
3. Update Git: Change `spec.script.bundle.path` to `api-gateway/v1.2.3/worker.js` and commit
4. Deploy: `planton apply -f cloudflareworker.yaml`

The Planton controller fetches the bundle from R2 and deploys it via the Cloudflare API.

### Separate Secret Management

The `CloudflareWorkerEnv` message includes both `variables` (plaintext) and `secrets` (encrypted), but secrets are treated differently:

```protobuf
message CloudflareWorkerEnv {
  map<string, string> variables = 1;  // Deployed with the Version
  map<string, string> secrets = 2;    // Uploaded separately via Secrets API
}
```

**How it works:**

- **Variables** are declarative. They're included in the Worker Version and reconciled on every `planton apply`.
- **Secrets** are imperative. They're uploaded once via a separate API call and *never* read back. The controller sets them if they're present in the spec, but it cannot verify they're correct on subsequent runs.

Best practice: Reference secrets from external stores using the `$secrets-group/...` syntax:

```yaml
spec:
  env:
    variables:
      LOG_LEVEL: info
    secrets:
      API_KEY: $secrets-group/external-apis/stripe-key
```

The controller resolves these references at deployment time, fetching the actual secret from a secure store (Kubernetes Secrets, HashiCorp Vault, etc.).

### The 80/20 Configuration Surface

The `CloudflareWorkerSpec` exposes only the fields that 80% of users need:

| Field | Type | Purpose |
|-------|------|---------|
| `account_id` | string (required) | Cloudflare Account ID (32-char hex) |
| `script.name` | string (required) | Worker service name |
| `script.bundle` | R2Object (required) | Reference to bundled artifact |
| `compatibility_date` | string (optional) | Runtime version (e.g., "2024-03-01") |
| `dns.enabled` | bool | Enable custom domain routing |
| `dns.zone_id` | string | Cloudflare Zone for DNS |
| `dns.hostname` | string | FQDN where Worker is accessible |
| `kv_bindings` | repeated | Bindings to KV namespaces |
| `env.variables` | map | Plaintext environment variables |
| `env.secrets` | map | Encrypted secrets (write-only) |

**What we omit:**

- D1 database bindings (20% use case)
- R2 bucket bindings (20% use case)
- Durable Object bindings (advanced use case)
- Service bindings (Worker-to-Worker communication)
- Cron triggers (advanced use case)
- Tail consumers (log processing)

These can be added as optional fields when needed, but most Workers are simple request handlers with KV storage.

---

## Production Configuration Examples

### Example 1: Minimal Worker (Dev/Test)

**Use Case:** Simple webhook handler for a staging environment.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: webhook-handler-dev
spec:
  account_id: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  script:
    name: webhook-handler
    bundle:
      bucket: dev-worker-bundles
      path: webhook-handler/v0.1.0/worker.js
  compatibility_date: "2024-03-01"
  dns:
    enabled: false  # Deploy without a route (testing only)
```

**Rationale:**
- No DNS routing (accessed via `*.workers.dev` subdomain for testing)
- Minimal configuration, fast iteration
- Bundle versioned in R2 for repeatability

---

### Example 2: API Gateway (Production)

**Use Case:** Production API gateway with KV caching and custom domain.

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway-prod
spec:
  account_id: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  script:
    name: api-gateway
    bundle:
      bucket: prod-worker-bundles
      path: api-gateway/v1.2.3/worker.js
  compatibility_date: "2024-03-01"
  kv_bindings:
    - valueFrom:
        kind: CloudflareKvNamespace
        name: api-cache-prod
        field_path: status.outputs.namespace_id
  dns:
    enabled: true
    zone_id: z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4
    hostname: api.example.com
    route_pattern: api.example.com/*
  env:
    variables:
      ENVIRONMENT: production
      LOG_LEVEL: info
    secrets:
      UPSTREAM_API_KEY: $secrets-group/external-apis/upstream-key
```

**Rationale:**
- Custom domain (`api.example.com`) with automatic DNS record creation
- KV binding for response caching
- Secrets referenced from external secret manager
- Explicit compatibility date for production stability
- Immutable artifact (R2 bundle)

---

### Example 3: Multi-Environment Deployment

**Pattern:** Separate Worker resources for staging and production, sharing the same bundle but with different configs.

**staging.yaml:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway-staging
spec:
  account_id: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  script:
    name: api-gateway-staging
    bundle:
      bucket: staging-worker-bundles
      path: api-gateway/v1.2.3/worker.js  # Same bundle as prod
  kv_bindings:
    - valueFrom:
        kind: CloudflareKvNamespace
        name: api-cache-staging  # Different KV namespace
        field_path: status.outputs.namespace_id
  dns:
    enabled: true
    zone_id: z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4
    hostname: staging-api.example.com
  env:
    variables:
      ENVIRONMENT: staging
```

**production.yaml:**

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorker
metadata:
  name: api-gateway-prod
spec:
  account_id: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  script:
    name: api-gateway-prod
    bundle:
      bucket: prod-worker-bundles
      path: api-gateway/v1.2.3/worker.js  # Promoted from staging
  kv_bindings:
    - valueFrom:
        kind: CloudflareKvNamespace
        name: api-cache-prod
        field_path: status.outputs.namespace_id
  dns:
    enabled: true
    zone_id: z9y8x7w6v5u4t3s2r1q0p9o8n7m6l5k4
    hostname: api.example.com
  env:
    variables:
      ENVIRONMENT: production
```

**Key insight:** These are **separate Workers** (different service names), not different deployments of the same Worker. This mirrors Wrangler's `[env.staging]` behavior and prevents cross-environment contamination.

---

## Production Best Practices

### 1. Compatibility Dates Are Critical

Always set `compatibility_date`. Cloudflare maintains strict backward compatibility by releasing runtime bug fixes behind "compatibility flags." Setting the date opts your Worker into all fixes up to that date.

Without it, you may inadvertently depend on buggy behavior that gets fixed in a future runtime version.

### 2. Bundle Immutability

Never rebuild the same version. Use semantic versioning in R2 paths (`api-gateway/v1.2.3/worker.js`). When promoting to production, reference the exact staging artifact.

### 3. Observability

For production Workers, enable **Logpush**:

```yaml
# Future API extension (not yet in spec)
observability:
  logpush_enabled: true
  logpush_destination: r2://my-logs-bucket/workers/
```

Use `wrangler tail <worker-name>` for live debugging during incidents.

### 4. KV Binding Strategy

Create separate KV namespaces per environment:
- `api-cache-dev`
- `api-cache-staging`
- `api-cache-prod`

Never share KV data across environments.

### 5. Secrets Rotation

Because secrets are write-only, implement a rotation process:
1. Update secret in external store
2. Trigger a pipeline that calls `wrangler secret put` or the Secrets API
3. Verify the Worker can still authenticate to external services

---

## Key Takeaways

1. **Cloudflare Workers are not general-purpose FaaS.** They're optimized for edge middleware, delivering sub-50ms latency by running in V8 isolates, not containers. Use them for request interception, not batch processing.

2. **Wrangler is essential for development.** Use `wrangler dev` for local testing. The `workerd` runtime provides a high-fidelity simulation of production.

3. **IaC is essential for production.** Terraform and Pulumi provide declarative state management, but both require solving the artifact management problem.

4. **Secrets are write-only.** This is a security feature, not a bug. Treat secret management as a separate, imperative process, not part of the declarative Worker spec.

5. **OpenMCF solves artifact storage.** By referencing R2-hosted bundles, the `CloudflareWorker` API decouples build from deploy, enabling immutable artifacts and GitOps workflows.

6. **Workers are primitives, not platforms.** Cloudflare Pages is the opinionated, GitOps-native platform. Workers are the low-level primitive for teams that want declarative control.

---

## Further Reading

- **Cloudflare Workers Documentation:** [developers.cloudflare.com/workers](https://developers.cloudflare.com/workers/)
- **Wrangler CLI Reference:** [developers.cloudflare.com/workers/wrangler](https://developers.cloudflare.com/workers/wrangler/)
- **Workers API Reference:** [developers.cloudflare.com/api/resources/workers](https://developers.cloudflare.com/api/resources/workers/)
- **Terraform Cloudflare Provider:** [registry.terraform.io/providers/cloudflare/cloudflare](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- **How Workers Work (V8 Isolates):** [developers.cloudflare.com/workers/reference/how-workers-works](https://developers.cloudflare.com/workers/reference/how-workers-works/)

---

**Bottom Line:** Cloudflare Workers deliver true edge computing with zero cold starts by running code in V8 isolates instead of containers. They're production-ready for latency-sensitive request interception, but they require careful artifact management and a clear understanding of their constraints. OpenMCF's R2-based bundle model and separation of secret management make Workers deployable in a fully declarative, GitOps-native way—solving the hardest problems in Workers IaC while keeping the API minimal and focused on what 80% of users actually need.

