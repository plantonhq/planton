# CloudflareWorkersKvPair

Seed a single key-value entry into a Workers KV namespace as a first-class,
composable resource — distinct from the high-churn data a Worker writes at
runtime.

## When to use

Use this to manage configuration through infrastructure:

- Feature flags and config keys that should be versioned and reviewed.
- Values derived from other resources (e.g. write a created bucket name or zone
  ID into a config key a Worker reads).

For high-volume application data written by the Worker itself, do not model each
key here — write it from the Worker at runtime.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareWorkersKvPair
metadata:
  name: app-config-entry
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  namespaceId:
    valueFrom:
      kind: CloudflareKvNamespace
      name: app-config
      fieldPath: status.outputs.namespace_id
  keyName: feature.new-dashboard
  value: "true"
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `namespaceId` | yes | KV namespace ID, or a reference to a CloudflareKvNamespace |
| `keyName` | yes | The entry key (≤512 bytes) |
| `value` | yes | The value (≤25 MiB) |
| `metadata` | no | Arbitrary JSON returned with the value on read |

## Outputs

| Output | Description |
|---|---|
| `key_name` | The entry's key name |
| `namespace_id` | The namespace ID the entry was written to |

## A note on secrets

KV values are not secrets. Keep credentials out of KV: use a Worker `secret_text`
binding or Cloudflare Secrets Store, both of which are secret-by-default.

## Related components

- `CloudflareKvNamespace` — the container this entry is written into.
- `CloudflareWorker` — binds the namespace via `kv_namespaces` to read entries.
