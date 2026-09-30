# Cloudflare Worker

Deploys a Cloudflare Worker — a script that runs on Cloudflare's edge — with its bindings, routing, schedules, and runtime settings. Bindings are grouped by type (the wrangler.toml grain) and each cross-resource binding accepts a literal or a reference to the producing resource, so a Worker composes as a real node in the resource graph.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Workers Script** -- the serverless function, from inline `content` or an R2 `r2Bundle`, with optional static `assets`
- **Bindings** -- typed lists (vars, secrets, KV, R2, D1, Hyperdrive, services, queues, Durable Objects, and the rest of the provider's binding types) flattened into the script's bindings array
- **workers.dev subdomain** -- created only when `workersDev.enabled` is true
- **Custom domains** -- one managed hostname (automatic TLS) per `customDomains` entry
- **Routes** -- one pattern-based route per `routes` entry
- **Cron trigger** -- created only when `schedules` is set
- **Observability** -- Workers Logs and traces when `observability` is set

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection with a Cloudflare API token that has Workers Scripts Write (and Routes / DNS Edit if you attach custom domains or routes). A Worker built from an `r2Bundle` also needs the connection's R2 key pair, with Object Read on the bundle bucket.

### Cloudflare Account

- **A script source or assets** -- inline `content`, an R2 `r2Bundle`, a static `assets` directory, or a combination. A Worker with none of these is rejected.
- **Upstream resources** (optional) -- KV, D1, R2, Hyperdrive, another Worker, a Queue, or a Zero Trust tunnel, referenced via `valueFrom`.

## Deploy

### Console

Open the deployment store, find **Cloudflare Worker**, and click **Deploy**. The creation wizard walks through the account and script source, bindings, routing, and runtime settings. Start from the **Minimal Worker** preset in the [Presets](#presets) tab for an inline hello-world, or **Edge API with Custom Domain** for a production-shaped Worker.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorker
metadata:
  name: api-gateway
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  workerName: api-gateway
  content: |
    export default { async fetch() { return new Response("ok"); } }
  workersDev:
    enabled: true
```

```shell
planton apply -f cloudflare-worker.yaml
```

This deploys an inline hello-world Worker reachable on its `workers.dev` subdomain. A Stack Job tracks the provisioning in real time.

### InfraChart

Wire a Worker to a KV namespace deployed in the same pipeline:

```yaml
spec:
  kvNamespaces:
    - name: CONFIG
      namespaceId:
        valueFrom:
          kind: CloudflareKvNamespace
          name: session-cache
          fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, provisions the KV namespace first, then deploys the Worker with the resolved namespace ID bound.

## Key Configuration

These are the most important decisions when configuring a Worker. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Script source** -- `content` for inline ES modules, `r2Bundle` for a CI-built artifact (`bucket` is a CloudflareR2Bucket reference or a literal name; read with the connection's R2 keys, whatever the object's Content-Type), `assets` for a static site or full-stack app.

**Bindings** -- grouped by type. Cross-resource fields take a literal or `valueFrom`. Secrets and secret-key material are sensitive.

**Routing** -- `workersDev` for `*.workers.dev`, `customDomains` for managed hostnames, `routes` for zone patterns.

**Migrations** -- Durable Object class create/rename/transfer/delete. Cloudflare treats the tag as a one-shot; a second apply of the same `newTag` is rejected.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareKvNamespace** (optional) | `kvNamespaces[].namespaceId` | `status.outputs.namespace_id` |
| **CloudflareD1Database** (optional) | `d1Databases[].databaseId` | `status.outputs.database_id` |
| **CloudflareR2Bucket** (optional) | `r2Buckets[].bucketName`, `r2Bundle.bucket` | `status.outputs.bucket_name` |
| **CloudflareHyperdriveConfig** (optional) | `hyperdriveConfigs[].configId` | `status.outputs.hyperdrive_id` |
| **CloudflareQueue** (optional) | `queues[].queueName` | `status.outputs.queue_name` |
| **CloudflareWorker** (optional) | `services[].service`, `durableObjects[].scriptName`, `tailConsumers[].service` | `status.outputs.script_name` |
| **CloudflareDnsZone** (optional) | `customDomains[].zoneId`, `routes[].zoneId` | `status.outputs.zone_id` |
| **CloudflareZeroTrustTunnel** (optional) | `vpcNetworks[].tunnelId` | `status.outputs.tunnel_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `script_name` | The script name | Service bindings, tail consumers, and Durable Object bindings on other Workers; a CloudflareQueue's worker consumer; a Pages project's service binding |

`status.outputs` also carries `script_id`, the attached `custom_domain_hostnames` and `route_patterns`, and the ID maps (`custom_domain_ids`, `route_ids`, `route_zone_ids`) that make the routes and domains importable.

## Common Patterns

**API with custom domain** -- CI-built bundle in R2, KV + D1 by reference, managed custom domain. Start from the **Edge API with Custom Domain** preset.

**Minimal worker** -- inline content, workers.dev only. Start from the **Minimal Worker** preset.

**Static site / full-stack** -- `assets` alone, or `assets` plus a script whose `runWorkerFirst` rules route dynamic paths through code. Start from the **Static Site (Workers Static Assets)** or **Full-Stack App (script + Static Assets)** presets.

## Works With

- [**Cloudflare KV Namespace**](/cloud-catalog/cloudflare-kv-namespace) -- bound for edge key-value reads and writes
- [**Cloudflare D1 Database**](/cloud-catalog/cloudflare-d1-database) -- bound for serverless SQL access
- [**Cloudflare R2 Bucket**](/cloud-catalog/cloudflare-r2-bucket) -- bound for object storage, or holds the CI-built script bundle (`r2Bundle`)
- [**Cloudflare Queue**](/cloud-catalog/cloudflare-queue) -- the Worker produces to it via a `queues` binding, or consumes it as the queue's worker consumer
- [**Cloudflare Hyperdrive Config**](/cloud-catalog/cloudflare-hyperdrive-config) -- bound for pooled access to a regional SQL database
- [**Cloudflare DNS Zone**](/cloud-catalog/cloudflare-dns-zone) -- hosts the Worker's custom domains and route patterns
- [**Cloudflare Zero Trust Tunnel**](/cloud-catalog/cloudflare-zero-trust-tunnel) -- bound through `vpcNetworks` so the Worker reaches private networks
