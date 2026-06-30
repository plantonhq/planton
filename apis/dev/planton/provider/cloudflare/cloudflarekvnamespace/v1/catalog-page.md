# Cloudflare KV Namespace

Deploys a Workers KV namespace on Cloudflare. Workers KV is a globally distributed, low-latency key-value store designed for read-heavy workloads at the edge. This component creates a single KV namespace that can be bound to Cloudflare Workers for storing configuration, session data, feature flags, or any other data your edge logic needs.

## What Gets Created

When you deploy a CloudflareKvNamespace resource, Planton provisions:

- **Workers KV Namespace** — a `cloudflare_workers_kv_namespace` resource with the title set to the configured `namespaceName`

The namespace ID is exported as a stack output so that other components (such as CloudflareWorker) can reference it at deploy time.

## Prerequisites

- **Cloudflare credentials** configured via environment variables or Planton provider config
- **A Cloudflare account** with Workers KV enabled on the plan
- **Appropriate permissions** — the API token must have `Workers KV Storage:Edit` access

## Quick Start

Create a file `kv-namespace.yaml`:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareKvNamespace
metadata:
  name: my-kv
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CloudflareKvNamespace.my-kv
spec:
  namespaceName: my-kv-store
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

Deploy:

```shell
planton apply -f kv-namespace.yaml
```

This creates a KV namespace titled `my-kv-store` in your Cloudflare account. Bind it to a Worker using the namespace ID from `status.outputs.namespaceId`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespaceName` | `string` | A human-readable name for the KV namespace. Must be unique within the Cloudflare account. | Required, max 64 characters |
| `accountId` | `string` | The Cloudflare account ID that owns the namespace. | Required, 32 hex characters |

A KV namespace carries only an account and a title; there are no per-namespace
TTL or description settings (TTL is set per write). Seed entries with
`CloudflareWorkersKvPair`, or have the Worker write them at runtime.

## Examples

### Basic KV Namespace

A minimal KV namespace suitable for development or testing:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareKvNamespace
metadata:
  name: dev-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.CloudflareKvNamespace.dev-cache
spec:
  namespaceName: dev-cache
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

### Session Store

A KV namespace intended for session data (TTL is applied per write by the Worker):

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareKvNamespace
metadata:
  name: session-store
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CloudflareKvNamespace.session-store
spec:
  namespaceName: prod-session-store
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

### Feature Flags Namespace

A KV namespace dedicated to feature flag storage, referenced by multiple Workers.
Seed individual flags with `CloudflareWorkersKvPair`:

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareKvNamespace
metadata:
  name: feature-flags
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.CloudflareKvNamespace.feature-flags
spec:
  namespaceName: feature-flags-prod
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespaceId` | `string` | The unique identifier of the created KV namespace in Cloudflare |
| `supportsUrlEncoding` | `bool` | Whether keys in this namespace support URL encoding |

## Related Components

- [CloudflareWorkersKvPair](/docs/catalog/cloudflare/cloudflareworkerskvpair) — seed individual key-value entries into this namespace as managed, composable resources
- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — Workers consume KV namespaces via bindings; use the `namespaceId` output to wire a KV store to your Worker
- [CloudflareR2Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) — object storage for larger or binary data, complementary to KV for small-value, read-heavy access patterns
- [CloudflareD1Database](/docs/catalog/cloudflare/cloudflared1database) — relational SQL storage at the edge; use when you need structured queries rather than simple key-value lookups
- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — manages DNS zones that front the Workers consuming this KV namespace
