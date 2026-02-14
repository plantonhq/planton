# Cloudflare KV Namespace

Deploys a Workers KV namespace on Cloudflare. Workers KV is a globally distributed, low-latency key-value store designed for read-heavy workloads at the edge. This component creates a single KV namespace that can be bound to Cloudflare Workers for storing configuration, session data, feature flags, or any other data your edge logic needs.

## What Gets Created

When you deploy a CloudflareKvNamespace resource, OpenMCF provisions:

- **Workers KV Namespace** — a `cloudflare_workers_kv_namespace` resource with the title set to the configured `namespaceName`

The namespace ID is exported as a stack output so that other components (such as CloudflareWorker) can reference it at deploy time.

## Prerequisites

- **Cloudflare credentials** configured via environment variables or OpenMCF provider config
- **A Cloudflare account** with Workers KV enabled on the plan
- **Appropriate permissions** — the API token must have `Workers KV Storage:Edit` access

## Quick Start

Create a file `kv-namespace.yaml`:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: my-kv
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareKvNamespace.my-kv
spec:
  namespaceName: my-kv-store
```

Deploy:

```shell
openmcf apply -f kv-namespace.yaml
```

This creates a KV namespace titled `my-kv-store` in your Cloudflare account. Bind it to a Worker using the namespace ID from `status.outputs.namespaceId`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespaceName` | `string` | A human-readable name for the KV namespace. Must be unique within the Cloudflare account. | Required, max 64 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttlSeconds` | `int32` | `0` (no expiry) | Default time-to-live for key-value entries, in seconds. A value of `0` or unset means keys never expire. If set, must be at least `60` (Cloudflare minimum for expiring keys). **Note:** this field is stored in the spec for documentation purposes but is not currently enforced by the underlying Pulumi provider resource. |
| `description` | `string` | `""` | A short description of the namespace, useful for identifying its purpose. Maximum 256 characters. **Note:** this field is stored in the spec for documentation purposes but is not currently enforced by the underlying Pulumi provider resource. |

## Examples

### Basic KV Namespace

A minimal KV namespace suitable for development or testing:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: dev-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CloudflareKvNamespace.dev-cache
spec:
  namespaceName: dev-cache
```

### Session Store with TTL Hint

A KV namespace intended for session data, with the spec recording a 1-hour TTL intent:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: session-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareKvNamespace.session-store
spec:
  namespaceName: prod-session-store
  ttlSeconds: 3600
  description: "Session data for authenticated users"
```

### Feature Flags Namespace

A KV namespace dedicated to feature flag storage, referenced by multiple Workers:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareKvNamespace
metadata:
  name: feature-flags
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CloudflareKvNamespace.feature-flags
spec:
  namespaceName: feature-flags-prod
  description: "Global feature flags consumed by edge Workers"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespaceId` | `string` | The unique identifier of the created KV namespace in Cloudflare |

## Related Components

- [CloudflareWorker](/docs/catalog/cloudflare/cloudflareworker) — Workers consume KV namespaces via bindings; use the `namespaceId` output to wire a KV store to your Worker
- [CloudflareR2Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) — object storage for larger or binary data, complementary to KV for small-value, read-heavy access patterns
- [CloudflareD1Database](/docs/catalog/cloudflare/cloudflared1database) — relational SQL storage at the edge; use when you need structured queries rather than simple key-value lookups
- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — manages DNS zones that front the Workers consuming this KV namespace
