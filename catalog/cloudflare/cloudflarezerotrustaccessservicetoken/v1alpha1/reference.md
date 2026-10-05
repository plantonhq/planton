# CloudflareZeroTrustAccessServiceToken

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustAccessServiceTokenSpec defines an Access service token: a
machine credential (a client-ID / client-secret pair) that non-human clients
present in the `CF-Access-Client-ID` / `CF-Access-Client-Secret` request
headers to pass through Access-protected applications without an identity
provider login.

THE SECRET IS RETURNED ONLY AT CREATION AND AT ROTATION -- Cloudflare never
returns it again on reads, and an imported token cannot recover it. Capture
the `client_secret` output into a secret store at deploy time; a lost
secret means rotating the token.

Rotation is first-class: increment client_secret_version to mint a new secret,
and set previous_client_secret_expires_at to control how long the OLD secret
keeps working (extend it to give services time to migrate, or set it in the
past to kill a compromised secret immediately). The two fields work as a pair.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessServiceToken
metadata:
  name: test-service-token
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ci-deployer
  duration: 8760h
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.duration` | `string` |  |  |  |
| `spec.clientSecretVersion` | `int32` |  |  |  |
| `spec.previousClientSecretExpiresAt` | `string` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account ID that owns this service token. Set this for an
account-scoped token (the common case). Mutually exclusive with zone_id.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The Cloudflare zone this token is scoped to, as a literal zone ID or a
reference to a CloudflareDnsZone. Set this for a zone-scoped token. Mutually
exclusive with account_id.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

The display name of the service token.

- rule: {"string":{"minLen":"1"}}

### spec.duration

`string`

How long the token is valid, as a Go-style duration ("300ms", "2h45m",
"8760h") or the special value "forever" for a non-expiring token. Leave
empty for the Cloudflare default of one year (8760h).

- rule: duration must be a Go-style duration (e.g. 8760h, 2h45m) or 'forever'

### spec.clientSecretVersion

`int32` · optional (explicit presence)

Version number of the current client secret. Incrementing it triggers a
ROTATION: Cloudflare mints a new secret (returned once in the
client_secret output) and keeps accepting the previous secret until
previous_client_secret_expires_at. Leave unset until the first rotation
(Cloudflare treats the initial secret as version 1). Creating a token
with a higher version directly also works (measured live 2026-08-26);
the version only triggers a rotation when it increases past the token's
current one, so re-asserting an equal version is a no-op.

- rule: {"int32":{"gte":1}}

### spec.previousClientSecretExpiresAt

`string`

When the PREVIOUS client secret stops being accepted after a rotation, as an
RFC3339 timestamp (e.g. "2026-09-01T00:00:00Z"). Extend it into the future
to give services time to migrate to the new secret; set it in the past to
invalidate a compromised secret immediately. Required whenever
client_secret_version is set.

- rule: previous_client_secret_expires_at must be an RFC3339 timestamp (e.g. 2026-09-01T00:00:00Z) or empty

## Validation Rules

- `spec.account_xor_zone`: set exactly one of account_id or zone_id
- `spec.rotation_pair`: client_secret_version and previous_client_secret_expires_at must be set together (or both left unset)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustAccessServiceToken, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_token_id` | `string` | The UUID of the service token (the API identity used for import and policy `service_token` rules). |
| `status.outputs.client_id` | `string` | The Client ID the machine client presents in the CF-Access-Client-ID request header. |
| `status.outputs.client_secret` | `string` (sensitive) | The Client Secret the machine client presents in the CF-Access-Client-Secret request header -- sensitive both here (machine-readable) and in the modules' output registration. Cloudflare returns it ONLY at creation and at rotation -- it can never be read back later, and it does not survive an import. Capture it into a secret store at deploy time. |
| `status.outputs.expires_at` | `string` | When the token expires, as an RFC3339 timestamp. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
