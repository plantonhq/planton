# Cloudflare Email Routing Address

Declare a verified destination address for Email Routing.

## What Gets Created

- A `cloudflare_email_routing_address` (account-scoped destination). A
  verification email is sent to the mailbox on creation.

## Prerequisites

- A Cloudflare account ID.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `email` — destination email address (immutable).

## Stack Outputs

| Output | Description |
|---|---|
| `address_id` | The address identifier |
| `email` | The destination email |
| `verified` | Verification timestamp (empty until verified) |
| `created` | Creation timestamp |

## Related Components

- CloudflareEmailRoutingRule
- CloudflareEmailRoutingZone
