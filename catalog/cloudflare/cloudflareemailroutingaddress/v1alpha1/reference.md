# CloudflareEmailRoutingAddress

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareEmailRoutingAddressSpec declares a verified destination address for
Email Routing. Destination addresses are account-scoped (shared across zones)
and referenced by routing rules and catch-all rules as forwarding targets.

Creating an address triggers a verification email to that mailbox; the address
is not usable as a forwarding target until its owner clicks the verification
link. The `verified` output reflects this (null until verified).

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingAddress
metadata:
  name: test-email-routing-address
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  email: ops@example.com
  status: unverified
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.email` | `string` | yes |  |  |
| `spec.status` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this destination address.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.email

`string` · required

The destination email address (e.g. "ops@example.com"). A verification email
is sent here on creation. Immutable — changing it replaces the address.

- rule: {"required":true,"string":{"email":true}}

### spec.status

`string`

Explicit verification-state override. Leave empty to let the normal
verification flow (the emailed link) manage the state. Cloudflare permits
non-admin callers only to set a verified address back to "unverified";
setting "verified" requires admin privileges on the account.

- rule: status must be empty, 'unverified', or 'verified'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareEmailRoutingAddress, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.address_id` | `string` | The Cloudflare-assigned identifier of the destination address. |
| `status.outputs.email` | `string` | The destination email address (echoed; referenced by routing rules). |
| `status.outputs.verified` | `string` | RFC3339 timestamp of when the address was verified, or empty if not yet verified (verification requires the owner to click an emailed link). |
| `status.outputs.created` | `string` | RFC3339 timestamp of when the address was created. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareEmailRoutingRule | `spec.actions[].forwardTo` | `status.outputs.email` |
| CloudflareEmailRoutingZone | `spec.catchAll.actions[].forwardTo` | `status.outputs.email` |

## See Also

- [Overview](../README.md)
