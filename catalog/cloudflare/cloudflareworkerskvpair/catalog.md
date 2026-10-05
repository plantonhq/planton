# Cloudflare Workers KV Pair

Deploys a single key-value entry inside a Cloudflare Workers KV namespace, managed and versioned as infrastructure. It exists as a first-class Infra Component so configuration keys can be seeded and reviewed in code (and reference other resources' outputs) -- distinct from the high-churn application data a Worker writes at runtime. Each entry belongs to a KV namespace and is account-scoped.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KV Entry** -- a single key/value pair (with optional JSON metadata) written into the referenced namespace

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Workers KV Storage edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A KV namespace** -- an existing CloudflareKvNamespace to write into (reference it), or a literal namespace ID.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Cloudflare Account

- **Account-level access** -- KV namespaces and their entries are account-scoped, so the API token must be scoped to the account.

## Deploy

### Console

Open the deployment store, find **Cloudflare Workers KV Pair**, and click **Deploy**. The creation wizard captures the owning account and parent namespace (a searchable selector lists live namespaces in the account), then the key, value, and optional JSON metadata. A connection diagram shows the Namespace -> Entry edge. Start from the **Standard KV Entry** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorkersKvPair
metadata:
  name: new-checkout-flag
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  namespaceId:
    valueFrom:
      kind: CloudflareKvNamespace
      name: app-config
      fieldPath: status.outputs.namespace_id
  keyName: feature-flags/new-checkout
  value: '{"enabled": true}'
  metadata: '{"updatedBy": "platform"}'
```

```shell
planton apply -f cloudflare-workers-kv-pair.yaml
```

This writes a `feature-flags/new-checkout` key into the `app-config` namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the namespace and its seeded entries together, wiring each entry with ValueFromRef:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: CloudflareKvNamespace
      name: app-config
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, provisions the namespace first, then writes the entry into the resolved namespace ID.

## Key Configuration

These are the most important decisions when configuring a KV pair. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace (`namespaceId`)** -- The KV namespace the entry is written into. Reference a CloudflareKvNamespace (recommended -- the link survives namespace recreation and shows in the resource graph) or paste a literal namespace ID. Immutable -- moving an entry to another namespace replaces it.

**Key (`keyName`)** -- The lookup key a Worker reads (`env.MY_KV.get(key)`), up to 512 bytes. Immutable -- changing it deletes the old key and creates a new one, so keep it stable once Workers depend on it.

**Value (`value`)** -- The stored payload, up to 25 MiB. KV is general-purpose storage and values are NOT secrets -- keep credentials out of KV; use a Worker secret binding or Cloudflare Secrets Store for those.

**Metadata (`metadata`)** -- Optional JSON (up to 1024 bytes) returned alongside the value on read, useful for a version number, owner, or other small annotation a Worker can branch on without parsing the value.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareKvNamespace** | `namespaceId` | `status.outputs.namespace_id` |

The field accepts a literal namespace ID or a ValueFromRef.

### What This Kind Provides

This kind has no consumable outputs: `status.outputs` only echoes `key_name` and `namespace_id` back from the spec. Workers read the entry at runtime through their KV binding to the parent namespace, not through this resource's outputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Feature flags** -- A small set of KV pairs holding flag state a Worker reads at the edge, versioned in code so flips are reviewable and auditable. Start from the **Standard KV Entry** preset.

**Seeded configuration** -- Bootstrap configuration keys (routing tables, allowlists, copy strings) a Worker depends on, kept in infrastructure rather than written ad hoc at runtime.

## Works With

- [**Cloudflare KV Namespace**](/infra-catalog/cloudflare-kv-namespace) -- the namespace this entry is written into (via `namespaceId`)
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- reads this entry at runtime through a `kv` binding to the namespace
