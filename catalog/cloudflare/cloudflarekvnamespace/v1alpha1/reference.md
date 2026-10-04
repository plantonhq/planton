# CloudflareKvNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareKvNamespaceSpec provisions a Workers KV namespace: a low-latency,
eventually-consistent key-value store readable from Workers at the edge. The
namespace is the container; individual entries are seeded as
CloudflareWorkersKvPair resources (or written by the application at runtime).

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareKvNamespace
metadata:
  name: example-kv-namespace
spec:
  namespaceName: myapp-config-prod
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespaceName` | `string` | yes |  |  |
| `spec.accountId` | `string` | yes |  |  |

## Field Details

### spec.namespaceName

`string` · required

A human-readable title for the KV namespace, unique within the account. This
maps to the namespace's `title` in the Cloudflare API.

- rule: {"required":true,"string":{"maxLen":"64"}}

### spec.accountId

`string` · required

The Cloudflare account ID that owns this KV namespace.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareKvNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_id` | `string` | The unique identifier (ID) of the created KV namespace. A Worker's `kv_namespace` binding and a CloudflareWorkersKvPair both reference this. |
| `status.outputs.supports_url_encoding` | `bool` | Whether keys in this namespace support URL encoding (reported by Cloudflare). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflarePagesProject | `spec.deploymentConfigs.preview.kvNamespaces[].namespaceId` | `status.outputs.namespace_id` |
| CloudflarePagesProject | `spec.deploymentConfigs.production.kvNamespaces[].namespaceId` | `status.outputs.namespace_id` |
| CloudflareWorker | `spec.kvNamespaces[].namespaceId` | `status.outputs.namespace_id` |
| CloudflareWorkersKvPair | `spec.namespaceId` | `status.outputs.namespace_id` |

## See Also

- [Overview](../README.md)
