# DigitalOceanSpacesKey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanSpacesKeySpec models the full digitalocean_spaces_key resource
surface: an access-key pair for Spaces (DigitalOcean's S3-compatible object
storage), optionally scoped to specific buckets through per-bucket grants.

Keys are managed through DigitalOcean's REST API (not the S3-compatible
endpoint), so provisioning needs only the account API token. The SECRET KEY
IS RETURNED EXACTLY ONCE -- in the create response -- and can never be
retrieved again; rotating a key means destroying and recreating it. The
key name and the grant list both update in place (the grant list is
replaced wholesale on every update); only the key material itself is
immutable.

## Example

```yaml
# Reference manifest for DigitalOceanSpacesKey -- protovalidate-valid,
# embedded as the reference page's Example block, and the document the
# offline tofu plan renders. Two documents: a per-bucket key (the common
# shape) and a full-access key (the empty-bucket grant grammar).
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSpacesKey
metadata:
  name: ci-uploads-key
spec:
  keyName: ci-uploads
  # Literal bucket names; use valueFrom to reference DigitalOceanBucket
  # resources instead.
  grants:
    - bucket:
        value: app-assets
      permission: readwrite
    - bucket:
        value: app-logs
      permission: read
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSpacesKey
metadata:
  name: admin-key
spec:
  keyName: spaces-admin
  # A fullaccess grant covers every bucket in the account and names none.
  grants:
    - permission: fullaccess
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.keyName` | `string` | yes |  |  |
| `spec.grants` | `[]DigitalOceanSpacesKeyGrant` |  |  |  |
| `spec.grants[].bucket` | `string \| valueFrom` |  |  | DigitalOceanBucket (`status.outputs.bucket_id`) |
| `spec.grants[].permission` | `string` | yes |  |  |

## Field Details

### spec.keyName

`string` · required

Human-readable name identifying the key in the DigitalOcean control
panel. Renames apply in place without touching the key material.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.grants

`[]DigitalOceanSpacesKeyGrant`

(Optional) Per-bucket access grants. When empty, DigitalOcean creates
the key with NO access to anything -- grants are the only thing that
authorizes it. Account-wide access is expressed as a single grant with
permission "fullaccess" and no bucket. Updates replace the whole grant
list in one call.

- rule: a fullaccess grant must not name a bucket; read and readwrite grants must name one

### spec.grants[].bucket

`string | valueFrom`

The bucket this grant scopes to. Use a literal bucket name or a
reference to a DigitalOceanBucket resource. Leave unset only for a
"fullaccess" grant, which covers every bucket in the account.

- references: DigitalOceanBucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.grants[].permission

`string` · required

Access level granted: read, readwrite, or fullaccess. This wall exists
only here -- the provider has no validator and silently maps any other
value to an EMPTY permission, creating a grant that authorizes nothing.

- rule: {"required":true,"string":{"in":["read","readwrite","fullaccess"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanSpacesKey, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.access_key` | `string` | The access key ID (also the resource's API identity). Pairs with secret_key as S3-style credentials against the Spaces endpoint. |
| `status.outputs.secret_key` | `string` (sensitive) | The secret access key -- a SECRET. DigitalOcean returns it ONLY in the create response; it can never be read again from the API, so this output is the only place it ever exists. Both provisioners mark it sensitive in their state. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.grants[].bucket` | DigitalOceanBucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
