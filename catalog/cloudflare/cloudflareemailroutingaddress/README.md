# CloudflareEmailRoutingAddress

Declare a verified destination address for Email Routing. Destination addresses
are account-scoped (shared across zones) and referenced as forwarding targets by
`CloudflareEmailRoutingRule` and a zone's catch-all.

## When to use

- Registering the real mailbox(es) that Email Routing forwards to.

> Creating an address sends a verification email to the mailbox. The address can
> only receive forwarded mail after its owner clicks the verification link; the
> `verified` output is empty until then.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareEmailRoutingAddress
metadata:
  name: ops-mailbox
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  email: ops@example.com
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `email` | yes | Destination email address (immutable) |
| `status` | no | Explicit verification-state override (`unverified` \| `verified`). Empty leaves the state to the emailed-link flow; Cloudflare permits non-admin callers only to flip a verified address back to `unverified` |

## Outputs

| Output | Description |
|---|---|
| `address_id` | The destination address identifier |
| `email` | The destination email |
| `verified` | Verification timestamp, or empty if not yet verified |
| `created` | Creation timestamp |

## Related kinds

- `CloudflareEmailRoutingRule` — references this address as a forwarding target.
- `CloudflareEmailRoutingZone` — enables Email Routing and a catch-all that can
  forward to this address.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
