# CloudflareEmailRoutingAddress — Research & Design Notes

## What it is

An account-scoped destination address that Email Routing forwards mail to. It is
referenced by routing rules and catch-all rules across any zone in the account,
which is why it is its own kind rather than folded into a zone.

## Verification (external dependency)

Creating an address triggers a verification email from Cloudflare. The address is
not usable as a forwarding target until its owner clicks the link. The resource
itself creates/destroys fine; only the `verified` state depends on the external
mailbox action, which is why end-to-end verified delivery is out of scope for
automated validation.

## Immutability

`email` is `RequiresReplace` — changing it replaces the address.

## Engine parity

A single resource on both engines. Outputs map to `address_id` / `email` /
`verified` / `created`.

## Composition

- Outputs `email` (referenced by rule/catch-all `forward_to`) and `address_id`.
