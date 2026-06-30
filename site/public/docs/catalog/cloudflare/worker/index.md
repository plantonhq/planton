---
title: "Worker"
description: "Worker deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareworker"
---

# Cloudflare Worker

Deploys a Cloudflare Worker from a pre-built script bundle stored in R2, with optional KV namespace bindings, custom domain routing, environment variables, and encrypted secrets.

## What Gets Created

When you deploy a CloudflareWorker resource, Planton provisions:

- **Workers Script** — the Worker script deployed to Cloudflare's edge network, loaded from an R2 bucket. Configured with `nodejs_compat` compatibility flag and observability enabled.
- **Plain-text Bindings** — environment variables from `env.variables` are bound as plain-text values accessible in the Worker runtime
- **KV Namespace Bindings** — references to CloudflareKvNamespace resources are bound to the Worker
- **Workers Route** — created only when DNS is enabled, attaches the Worker to a URL pattern on a Cloudflare zone
- **DNS A Record** — created only when DNS is enabled, a proxied record pointing the hostname through Cloudflare's network

## Prerequisites

- **A Cloudflare account** with the account ID (32-character hex string)
- **An R2 bucket** containing the pre-built Worker script bundle
- **A Cloudflare zone** (domain) if routing the Worker to a custom domain
- **KV namespaces** created via CloudflareKvNamespace if binding KV storage

## Quick Start

Create a file `worker.yaml`:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareWorker
metadata:
  name: my-worker
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CloudflareWorker.my-worker
spec:
  accountId: "0123456789abcdef0123456789abcdef"
  workerName: my-worker
  scriptBundle:
    bucket: my-worker-builds
    path: builds/my-worker/latest/worker.js
```

Deploy:

```shell
planton apply -f worker.yaml
```

This deploys a Worker script to Cloudflare's edge network from the specified R2 bundle. The Worker is accessible at `my-worker.<account>.workers.dev`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `accountId` | `string` | Cloudflare account ID. | Exactly 32 hex characters |
| `workerName` | `string` | Name of the Worker as shown in the Cloudflare dashboard. | 1-63 characters |
| `scriptBundle.bucket` | `string` | R2 bucket name containing the Worker script bundle. | Required |
| `scriptBundle.path` | `string` | Path to the script bundle within the R2 bucket. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kvBindings` | `ValueFromRef[]` | `[]` | KV namespace references to bind to the Worker. Each references a CloudflareKvNamespace resource. |
| `dns.enabled` | `bool` | `false` | Attach the Worker to a custom domain via a Workers route. |
| `dns.zoneId` | `string` | — | Cloudflare zone ID for the domain. Required when `dns.enabled` is `true`. |
| `dns.hostname` | `string` | — | Fully qualified domain name for the Worker (e.g., `api.example.com`). Required when `dns.enabled` is `true`. |
| `dns.routePattern` | `string` | `"hostname/*"` | URL pattern for the Workers route. Defaults to matching all paths under the hostname. |
| `compatibilityDate` | `string` | — | Compatibility date for the Worker runtime (format: `YYYY-MM-DD`). |
| `usageModel` | `enum` | `BUNDLED` | Billing model: `BUNDLED` (included CPU time) or `UNBOUND` (pay per millisecond). |
| `env.variables` | `map<string, string>` | `{}` | Plain-text environment variables accessible in the Worker. |
| `env.secrets` | `map<string, string>` | `{}` | Encrypted secrets uploaded via the Cloudflare Secrets API. Never logged. |

## Examples

### Worker with Custom Domain

Deploy a Worker accessible at a custom hostname:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareWorker
metadata:
  name: api-worker
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CloudflareWorker.api-worker
spec:
  accountId: "0123456789abcdef0123456789abcdef"
  workerName: api-worker
  scriptBundle:
    bucket: worker-bundles
    path: builds/api-worker/v1.2.3/worker.js
  compatibilityDate: "2026-01-01"
  dns:
    enabled: true
    zoneId: "fedcba9876543210fedcba9876543210"
    hostname: api.example.com
```

### Full-Featured with KV, Environment, and Secrets

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareWorker
metadata:
  name: webhook-handler
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CloudflareWorker.webhook-handler
spec:
  accountId: "0123456789abcdef0123456789abcdef"
  workerName: webhook-handler
  scriptBundle:
    bucket: worker-bundles
    path: builds/webhook-handler/latest/worker.js
  compatibilityDate: "2026-01-01"
  usageModel: UNBOUND
  kvBindings:
    - kind: CloudflareKvNamespace
      name: webhook-cache
      field: status.outputs.namespace_id
  env:
    variables:
      LOG_LEVEL: "info"
      ENVIRONMENT: "production"
    secrets:
      WEBHOOK_SECRET: "whsec_abc123..."
      API_TOKEN: "tok_xyz789..."
  dns:
    enabled: true
    zoneId: "fedcba9876543210fedcba9876543210"
    hostname: webhooks.example.com
    routePattern: "webhooks.example.com/api/*"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `script_id` | `string` | Cloudflare-assigned identifier for the deployed Worker script |
| `route_urls` | `string[]` | URLs or route patterns where the Worker is active (e.g., `webhooks.example.com/*`) |

## Related Components

- [CloudflareKvNamespace](/docs/catalog/cloudflare/kv-namespace) — create KV namespaces to bind to the Worker
- [CloudflareD1Database](/docs/catalog/cloudflare/d1-database) — deploy a D1 database for the Worker to query
- [CloudflareR2Bucket](/docs/catalog/cloudflare/r2-bucket) — R2 storage for the Worker to read and write objects
