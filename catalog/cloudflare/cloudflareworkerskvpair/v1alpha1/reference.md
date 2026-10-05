# CloudflareWorkersKvPair

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareWorkersKvPairSpec declares a single key-value entry inside a
Workers KV namespace. It exists as a first-class kind so configuration keys
can be seeded and versioned through infrastructure (and reference other
resources' outputs), distinct from the high-churn application data a Worker
writes at runtime.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorkersKvPair
metadata:
  name: test-kv-pair
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  namespaceId:
    value: "0f1e2d3c4b5a69788796a5b4c3d2e1f0"
  keyName: feature.new-dashboard
  value: "true"
  metadata: '{"owner":"platform"}'
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.namespaceId` | `string \| valueFrom` | yes |  | CloudflareKvNamespace (`status.outputs.namespace_id`) |
| `spec.keyName` | `string` | yes |  |  |
| `spec.value` | `string` | yes |  |  |
| `spec.metadata` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns the namespace this entry belongs to.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.namespaceId

`string | valueFrom` · required

The KV namespace this entry is written to. Can be a literal namespace ID or
a reference to a CloudflareKvNamespace resource (defaulting to that
namespace's status.outputs.namespace_id).

- references: CloudflareKvNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareKvNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.keyName

`string` · required

The entry's key. Up to 512 bytes. This is the lookup key a Worker reads
(e.g. env.MY_KV.get(key_name)).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.value

`string` · required

The value stored at key_name (up to 25 MiB). KV is general-purpose storage;
values are not treated as secrets. Keep credentials out of KV — use a Worker
`secret_text` binding or Cloudflare Secrets Store for those.

- rule: {"required":true}

### spec.metadata

`string`

Optional arbitrary JSON metadata associated with the entry (up to 1024
bytes), returned alongside the value on read. Leave empty for none.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareWorkersKvPair, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.key_name` | `string` | The entry's key name. |
| `status.outputs.namespace_id` | `string` | The namespace ID the entry was written to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespaceId` | CloudflareKvNamespace | `status.outputs.namespace_id` |

## See Also

- [Overview](../README.md)
