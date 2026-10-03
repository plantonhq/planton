# Cloudflare KV Namespace

Deploys a Workers KV namespace on Cloudflare for globally replicated, eventually consistent key-value storage at the edge. The namespace is the container: Workers bind it for low-latency reads across Cloudflare's network, and individual entries are seeded as `CloudflareWorkersKvPair` resources or written by the application at runtime. Writes can take up to about 60 seconds to become visible from every edge location, which makes KV right for read-heavy configuration and content and wrong for counters, locks, or read-after-write workloads.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workers KV Namespace** -- a named key-value store in the Cloudflare account, identified by a unique namespace ID that Workers reference as a binding

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Workers KV permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare account** with Workers KV enabled. The `accountId` field identifies which account owns the namespace; it is created at the account level and can be bound to any Worker in that account.

## Deploy

### Console

Open the deployment store, find **Cloudflare KV Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard KV Namespace** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareKvNamespace
metadata:
  name: session-cache
  org: acme-corp
  env: prod
spec:
  namespaceName: session-cache-prod
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
```

```shell
planton apply -f cloudflare-kv-namespace.yaml
```

This creates a Workers KV namespace titled `session-cache-prod` in the account. The namespace ID is exported in outputs for binding to CloudflareWorker resources. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a KV namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace name (`namespaceName`)** -- the title is account-unique identity: creating a second namespace with the same title in the account fails. Name namespaces for their purpose (`app-config-prod`), not their consumer -- several Workers can bind one namespace. Up to 64 characters, visible in the Cloudflare dashboard.

**Eventual consistency is the contract** -- a written value can take up to ~60 seconds to be visible from every edge location. If the workload needs read-after-write consistency, counters, or locks, this is the wrong store -- reach for Durable Objects or D1 instead of fighting KV's replication model.

**Destroy deletes every entry, immediately** -- there is no recycle bin. Entries seeded through `CloudflareWorkersKvPair` are recorded in IaC and can be re-created; runtime-written data is simply gone. Treat a namespace holding runtime data as a stateful resource when planning teardown.

**Seed configuration here, write data at runtime** -- infrastructure seeds slow-changing entries (feature flags, per-environment endpoints) as `CloudflareWorkersKvPair` resources, and the application writes high-churn data through its KV binding. IaC-managing hot data guarantees perpetual drift.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- the Cloudflare account is identified by the `accountId` string.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace_id` | The unique identifier of the created KV namespace | A CloudflareWorker's KV binding and every `CloudflareWorkersKvPair` seeded into the namespace reference this |

## Common Patterns

**Standard namespace** -- one namespace per purpose and environment, bound to the Workers that read it. KV namespaces are simple resources whose complexity lives in the Workers that consume them. Start from the **Standard KV Namespace** preset.

**Seeded configuration store** -- deploy the namespace, then seed feature flags and per-environment endpoints as `CloudflareWorkersKvPair` resources in the same chart, so configuration ships with the infrastructure while runtime data stays out of IaC.

## Works With

- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- binds the namespace (via `namespace_id`) for edge reads and runtime writes
- [**Cloudflare Workers KV Pair**](/infra-catalog/cloudflare-workers-kv-pair) -- seeds individual entries into this namespace as declarative resources
